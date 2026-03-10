package main

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ═══════════════════════════════════════════════════════════════
// Repository
// ═══════════════════════════════════════════════════════════════

type salaryRepo struct{ db *pgxpool.Pool }

func (r *salaryRepo) listByEmployee(ctx context.Context, empNo int) ([]*Salary, error) {
	rows, err := r.db.Query(ctx,
		`SELECT emp_no, amount, from_date, to_date FROM emp_salary
		 WHERE emp_no = $1 ORDER BY from_date DESC`, empNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Salary
	for rows.Next() {
		s := &Salary{}
		if err := rows.Scan(&s.EmpNo, &s.Amount, &s.FromDate, &s.ToDate); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *salaryRepo) create(ctx context.Context, empNo int, req *CreateSalaryReq) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO emp_salary (emp_no, amount, from_date, to_date) VALUES ($1,$2,$3,$4)`,
		empNo, req.Amount, req.FromDate, req.ToDate,
	)
	return err
}

func (r *salaryRepo) updateCurrent(ctx context.Context, empNo, amount int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE emp_salary SET amount = $1 WHERE emp_no = $2 AND to_date = '9999-01-01'`,
		amount, empNo,
	)
	return err
}

func (r *salaryRepo) delete(ctx context.Context, empNo int, fromDate string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM emp_salary WHERE emp_no = $1 AND from_date = $2", empNo, fromDate)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Handlers
// ═══════════════════════════════════════════════════════════════

type salaryHandler struct{ repo *salaryRepo }

// GET /api/v1/employees/:id/salaries
func (h *salaryHandler) list(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	list, err := h.repo.listByEmployee(c.Request.Context(), empNo)
	if err != nil {
		respError(c, err.Error())
		return
	}
	if list == nil {
		list = []*Salary{}
	}
	respOK(c, list)
}

// POST /api/v1/employees/:id/salaries
func (h *salaryHandler) create(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	var req CreateSalaryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	if err := h.repo.create(c.Request.Context(), empNo, &req); err != nil {
		respError(c, err.Error())
		return
	}
	respCreated(c, gin.H{"created": true})
}

// PATCH /api/v1/employees/:id/salaries/current
func (h *salaryHandler) updateCurrent(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	var req UpdateSalaryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, err.Error())
		return
	}
	if err := h.repo.updateCurrent(c.Request.Context(), empNo, req.Amount); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"updated": true})
}

// DELETE /api/v1/employees/:id/salaries/:from_date
func (h *salaryHandler) delete(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	if err := h.repo.delete(c.Request.Context(), empNo, c.Param("from_date")); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"deleted": true})
}
