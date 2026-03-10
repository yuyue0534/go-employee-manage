package main

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ═══════════════════════════════════════════════════════════════
// Repository
// ═══════════════════════════════════════════════════════════════

type departmentRepo struct{ db *pgxpool.Pool }

func (r *departmentRepo) list(ctx context.Context) ([]*Department, error) {
	rows, err := r.db.Query(ctx, "SELECT dept_no, dept_name FROM emp_department ORDER BY dept_no")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Department
	for rows.Next() {
		d := &Department{}
		if err := rows.Scan(&d.DeptNo, &d.DeptName); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *departmentRepo) getByID(ctx context.Context, deptNo string) (*Department, error) {
	d := &Department{}
	err := r.db.QueryRow(ctx,
		"SELECT dept_no, dept_name FROM emp_department WHERE dept_no = $1", deptNo,
	).Scan(&d.DeptNo, &d.DeptName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

func (r *departmentRepo) create(ctx context.Context, d *Department) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO emp_department (dept_no, dept_name) VALUES ($1,$2)", d.DeptNo, d.DeptName)
	return err
}

func (r *departmentRepo) update(ctx context.Context, deptNo, deptName string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE emp_department SET dept_name = $1 WHERE dept_no = $2", deptName, deptNo)
	return err
}

func (r *departmentRepo) delete(ctx context.Context, deptNo string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM emp_department WHERE dept_no = $1", deptNo)
	return err
}

func (r *departmentRepo) listEmployees(ctx context.Context, deptNo string) ([]*Employee, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.emp_no, e.birth_date, e.first_name, e.last_name, e.gender, e.hire_date
		 FROM emp_employee e
		 JOIN emp_dept_emp de ON de.emp_no = e.emp_no
		 WHERE de.dept_no = $1 AND de.to_date = '9999-01-01'
		 ORDER BY e.emp_no`, deptNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Employee
	for rows.Next() {
		e := &Employee{}
		if err := rows.Scan(&e.EmpNo, &e.BirthDate, &e.FirstName, &e.LastName, &e.Gender, &e.HireDate); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *departmentRepo) getManager(ctx context.Context, deptNo string) (*Employee, error) {
	e := &Employee{}
	err := r.db.QueryRow(ctx,
		`SELECT e.emp_no, e.birth_date, e.first_name, e.last_name, e.gender, e.hire_date
		 FROM emp_employee e
		 JOIN emp_dept_manager dm ON dm.emp_no = e.emp_no
		 WHERE dm.dept_no = $1 AND dm.to_date = '9999-01-01' LIMIT 1`, deptNo,
	).Scan(&e.EmpNo, &e.BirthDate, &e.FirstName, &e.LastName, &e.Gender, &e.HireDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return e, err
}

func (r *departmentRepo) assignManager(ctx context.Context, deptNo string, req *AssignManagerReq) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO emp_dept_manager (emp_no, dept_no, from_date, to_date)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (emp_no, dept_no) DO UPDATE
		 SET from_date = EXCLUDED.from_date, to_date = EXCLUDED.to_date`,
		req.EmpNo, deptNo, req.FromDate, req.ToDate,
	)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Handlers
// ═══════════════════════════════════════════════════════════════

type departmentHandler struct{ repo *departmentRepo }

// GET /api/v1/departments
func (h *departmentHandler) list(c *gin.Context) {
	list, err := h.repo.list(c.Request.Context())
	if err != nil {
		respError(c, err.Error())
		return
	}
	if list == nil {
		list = []*Department{}
	}
	respOK(c, list)
}

// GET /api/v1/departments/:id
func (h *departmentHandler) get(c *gin.Context) {
	d, err := h.repo.getByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respError(c, err.Error())
		return
	}
	if d == nil {
		respNotFound(c, "department not found")
		return
	}
	respOK(c, d)
}

// POST /api/v1/departments
func (h *departmentHandler) create(c *gin.Context) {
	var req CreateDepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	existing, _ := h.repo.getByID(c.Request.Context(), req.DeptNo)
	if existing != nil {
		respConflict(c, "dept_no already exists")
		return
	}
	if err := h.repo.create(c.Request.Context(), &Department{DeptNo: req.DeptNo, DeptName: req.DeptName}); err != nil {
		respError(c, err.Error())
		return
	}
	respCreated(c, gin.H{"dept_no": req.DeptNo})
}

// PATCH /api/v1/departments/:id
func (h *departmentHandler) update(c *gin.Context) {
	var req UpdateDepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	if err := h.repo.update(c.Request.Context(), c.Param("id"), req.DeptName); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"updated": true})
}

// DELETE /api/v1/departments/:id
func (h *departmentHandler) delete(c *gin.Context) {
	if err := h.repo.delete(c.Request.Context(), c.Param("id")); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"deleted": true})
}

// GET /api/v1/departments/:id/employees
func (h *departmentHandler) listEmployees(c *gin.Context) {
	list, err := h.repo.listEmployees(c.Request.Context(), c.Param("id"))
	if err != nil {
		respError(c, err.Error())
		return
	}
	if list == nil {
		list = []*Employee{}
	}
	respOK(c, list)
}

// GET /api/v1/departments/:id/manager
func (h *departmentHandler) getManager(c *gin.Context) {
	mgr, err := h.repo.getManager(c.Request.Context(), c.Param("id"))
	if err != nil {
		respError(c, err.Error())
		return
	}
	if mgr == nil {
		respNotFound(c, "no current manager")
		return
	}
	respOK(c, mgr)
}

// PUT /api/v1/departments/:id/manager
func (h *departmentHandler) assignManager(c *gin.Context) {
	var req AssignManagerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	if err := h.repo.assignManager(c.Request.Context(), c.Param("id"), &req); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"assigned": true})
}
