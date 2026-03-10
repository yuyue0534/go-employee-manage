package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter(
	emp *employeeHandler,
	dept *departmentHandler,
	sal *salaryHandler,
	title *titleHandler,
) *gin.Engine {
	engine := gin.New()

	engine.Use(
		loggerMiddleware(),
		gin.Recovery(),
		cors.New(cors.Config{
			AllowOriginFunc: func(origin string) bool {
				switch origin {
				case "http://localhost:3000",
					"http://127.0.0.1:3000",
					"http://localhost:5173",
					"http://127.0.0.1:5173",
					"http://localhost:5500",
					"http://127.0.0.1:5500",
					"http://localhost:8080",
					"http://127.0.0.1:8080",
					"null":
					return true
				default:
					return false
				}
			},
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Authorization",
				"X-Requested-With",
				"Accept",
			},
			ExposeHeaders: []string{
				"Content-Length",
				"Content-Type",
			},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")

	e := v1.Group("/employees")
	{
		e.GET("", emp.list)
		e.POST("", emp.create)
		e.GET("/:id", emp.get)
		e.PATCH("/:id", emp.update)
		e.DELETE("/:id", emp.delete)

		e.GET("/:id/salaries", sal.list)
		e.POST("/:id/salaries", sal.create)
		e.PATCH("/:id/salaries/current", sal.updateCurrent)
		e.DELETE("/:id/salaries/:from_date", sal.delete)

		e.GET("/:id/titles", title.list)
		e.POST("/:id/titles", title.create)
		e.DELETE("/:id/titles/:title/:from_date", title.delete)
	}

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
