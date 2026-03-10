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

type titleRepo struct{ db *pgxpool.Pool }

func (r *titleRepo) listByEmployee(ctx context.Context, empNo int) ([]*Title, error) {
	rows, err := r.db.Query(ctx,
		`SELECT emp_no, title, from_date, to_date FROM emp_title
		 WHERE emp_no = $1 ORDER BY from_date DESC`, empNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Title
	for rows.Next() {
		t := &Title{}
		if err := rows.Scan(&t.EmpNo, &t.Title, &t.FromDate, &t.ToDate); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *titleRepo) create(ctx context.Context, empNo int, req *AssignTitleReq) error {
	var toDate *string
	if req.ToDate != "" {
		toDate = &req.ToDate
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO emp_title (emp_no, title, from_date, to_date) VALUES ($1,$2,$3,$4)`,
		empNo, req.Title, req.FromDate, toDate,
	)
	return err
}

func (r *titleRepo) delete(ctx context.Context, empNo int, title, fromDate string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM emp_title WHERE emp_no = $1 AND title = $2 AND from_date = $3",
		empNo, title, fromDate,
	)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Handlers
// ═══════════════════════════════════════════════════════════════

type titleHandler struct{ repo *titleRepo }

// GET /api/v1/employees/:id/titles
func (h *titleHandler) list(c *gin.Context) {
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
		list = []*Title{}
	}
	respOK(c, list)
}

// POST /api/v1/employees/:id/titles
func (h *titleHandler) create(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	var req AssignTitleReq
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

// DELETE /api/v1/employees/:id/titles/:title/:from_date
func (h *titleHandler) delete(c *gin.Context) {
	empNo, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respBadRequest(c, "invalid employee id")
		return
	}
	if err := h.repo.delete(c.Request.Context(), empNo, c.Param("title"), c.Param("from_date")); err != nil {
		respError(c, err.Error())
		return
	}
	respOK(c, gin.H{"deleted": true})
}
