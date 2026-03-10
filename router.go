package main

import (
	"github.com/gin-gonic/gin"
)

func setupRouter(
	emp *employeeHandler,
	dept *departmentHandler,
	sal *salaryHandler,
	title *titleHandler,
) *gin.Engine {
	engine := gin.New()
	engine.Use(loggerMiddleware(), gin.Recovery())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")

	// ── Employees ────────────────────────────────────────────
	e := v1.Group("/employees")
	{
		e.GET("", emp.list)
		e.POST("", emp.create)
		e.GET("/:id", emp.get)
		e.PATCH("/:id", emp.update)
		e.DELETE("/:id", emp.delete)

		// Salaries
		e.GET("/:id/salaries", sal.list)
		e.POST("/:id/salaries", sal.create)
		e.PATCH("/:id/salaries/current", sal.updateCurrent)
		e.DELETE("/:id/salaries/:from_date", sal.delete)

		// Titles
		e.GET("/:id/titles", title.list)
		e.POST("/:id/titles", title.create)
		e.DELETE("/:id/titles/:title/:from_date", title.delete)
	}

	// ── Departments ──────────────────────────────────────────
	d := v1.Group("/departments")
	{
		d.GET("", dept.list)
		d.POST("", dept.create)
		d.GET("/:id", dept.get)
		d.PATCH("/:id", dept.update)
		d.DELETE("/:id", dept.delete)
		d.GET("/:id/employees", dept.listEmployees)
		d.GET("/:id/manager", dept.getManager)
		d.PUT("/:id/manager", dept.assignManager)
	}

	return engine
}
