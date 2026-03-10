package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ═══════════════════════════════════════════════════════════════
// Repository
// ═══════════════════════════════════════════════════════════════

type employeeRepo struct{ db *pgxpool.Pool }

func (r *employeeRepo) list(ctx context.Context, page, size int, gender, name string) ([]*Employee, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if gender != "" {
		where += fmt.Sprintf(" AND gender = $%d", idx)
		args = append(args, gender)
		idx++
	}
	if name != "" {
		where += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d)", idx, idx+1)
		args = append(args, "%"+name+"%", "%"+name+"%")
		idx += 2
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM emp_employee "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT emp_no, birth_date, first_name, last_name, gender, hire_date
		 FROM emp_employee %s ORDER BY emp_no LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, size, (page-1)*size)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*Employee
	for rows.Next() {
		e := &Employee{}
		if err := rows.Scan(&e.EmpNo, &e.BirthDate, &e.FirstName, &e.LastName, &e.Gender, &e.HireDate); err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

func (r *employeeRepo) getByID(ctx context.Context, empNo int) (*Employee, error) {
	e := &Employee{}
	err := r.db.QueryRow(ctx,
		`SELECT emp_no, birth_date, first_name, last_name, gender, hire_date
		 FROM emp_employee WHERE emp_no = $1`, empNo,
	).Scan(&e.EmpNo, &e.BirthDate, &e.FirstName, &e.LastName, &e.Gender, &e.HireDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return e, err
}

func (r *employeeRepo) getDetail(ctx context.Context, empNo int) (*EmployeeDetail, error) {
	e, err := r.getByID(ctx, empNo)
	if err != nil || e == nil {
		return nil, err
	}
	detail := &EmployeeDetail{Employee: *e}

	// current department
	var deptNo, deptName string
	if err := r.db.QueryRow(ctx,
		`SELECT d.dept_no, d.dept_name FROM emp_dept_emp de
		 JOIN emp_department d ON d.dept_no = de.dept_no
		 WHERE de.emp_no = $1 AND de.to_date = '9999-01-01' LIMIT 1`, empNo,
	).Scan(&deptNo, &deptName); err == nil {
		detail.Department = &Department{DeptNo: deptNo, DeptName: deptName}
	}

	// current title
	var title string
	if err := r.db.QueryRow(ctx,
		`SELECT title FROM emp_title
		 WHERE emp_no = $1 AND (to_date IS NULL OR to_date = '9999-01-01')
		 ORDER BY from_date DESC LIMIT 1`, empNo,
	).Scan(&title); err == nil {
		detail.CurrentTitle = &title
	}

	// current salary
	var amount int
	if err := r.db.QueryRow(ctx,
		`SELECT amount FROM emp_salary
		 WHERE emp_no = $1 AND to_date = '9999-01-01'
		 ORDER BY from_date DESC LIMIT 1`, empNo,
	).Scan(&amount); err == nil {
		detail.CurrentSalary = &amount
	}

	return detail, nil
}

func (r *employeeRepo) create(ctx context.Context, e *Employee) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO emp_employee (birth_date, first_name, last_name, gender, hire_date)
		 VALUES ($1,$2,$3,$4,$5) RETURNING emp_no`,
		e.BirthDate, e.FirstName, e.LastName, e.Gender, e.HireDate,
	).Scan(&id)
	return id, err
}

func (r *employeeRepo) update(ctx context.Context, empNo int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	set, args := "", []any{}
	idx := 1
	for k, v := range fields {
		if set != "" {
			set += ", "
		}
		set += fmt.Sprintf("%s = $%d", k, idx)
		args = append(args, v)
		idx++
	}
	args = append(args, empNo)
	_, err := r.db.Exec(ctx,
		fmt.Sprintf("UPDATE emp_employee SET %s WHERE emp_no = $%d", set, idx), args...)
	return err
}

func (r *employeeRepo) delete(ctx context.Context, empNo int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM emp_employee WHERE emp_no = $1", empNo)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Handlers
// ═══════════════════════════════════════════════════════════════

type employeeHandler struct{ repo *employeeRepo }

// GET /api/v1/employees
func (h *employeeHandler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	list, total, err := h.repo.list(c.Request.Context(), page, size, c.Query("gender"), c.Query("name"))
	if err != nil {
		respError(c, err.Error())
		return
	}
	if list == nil {
		list = []*Employee{}
	}
	respPaged(c, list, total, page, size)
}

// GET /api/v1/employees/:id
func (h *employeeHandler) get(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	detail, err := h.repo.getDetail(c.Request.Context(), empNo)
	if err != nil {
		respError(c, err.Error())
		return
	}
	if detail == nil {
		respNotFound(c, "employee not found")
		return
	}
	respOK(c, detail)
}

// POST /api/v1/employees
func (h *employeeHandler) create(c *gin.Context) {
	var req CreateEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	birth, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		respBadRequest(c, "birth_date must be YYYY-MM-DD")
		return
	}
	hire, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		respBadRequest(c, "hire_date must be YYYY-MM-DD")
		return
	}
	id, err := h.repo.create(c.Request.Context(), &Employee{
		BirthDate: birth, FirstName: req.FirstName,
		LastName: req.LastName, Gender: req.Gender, HireDate: hire,
	})
	if err != nil {
		respError(c, err.Error())
		return
	}
	respCreated(c, gin.H{"emp_no": id})
}

// PATCH /api/v1/employees/:id
func (h *employeeHandler) update(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	existing, err := h.repo.getByID(c.Request.Context(), empNo)
	if err != nil {
		respError(c, err.Error())
		return
	}
	if existing == nil {
		respNotFound(c, "employee not found")
		return
	}

	var req UpdateEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	fields := map[string]any{}
	if req.FirstName != "" {
		fields["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		fields["last_name"] = req.LastName
	}
	if req.Gender != "" {
		fields["gender"] = req.Gender
	}
	if req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			respBadRequest(c, "birth_date must be YYYY-MM-DD")
			return
		}
		fields["birth_date"] = t
	}
	if req.HireDate != "" {
		t, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			respBadRequest(c, "hire_date must be YYYY-MM-DD")
			return
		}
		fields["hire_date"] = t
	}
	if err := h.repo.update(c.Request.Context(), empNo, fields); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"updated": true})
}

// DELETE /api/v1/employees/:id
func (h *employeeHandler) delete(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	existing, err := h.repo.getByID(c.Request.Context(), empNo)
	if err != nil {
		respError(c, err.Error())
		return
	}
	if existing == nil {
		respNotFound(c, "employee not found")
		return
	}
	if err := h.repo.delete(c.Request.Context(), empNo); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"deleted": true})
}
