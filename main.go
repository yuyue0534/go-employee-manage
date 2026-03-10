package main

import (
	"fmt"
	"log"
)

func main() {
	cfg := loadConfig()
	pool := newPool(cfg.DatabaseURL)
	defer pool.Close()

	engine := setupRouter(
		&employeeHandler{repo: &employeeRepo{db: pool}},
		&departmentHandler{repo: &departmentRepo{db: pool}},
		&salaryHandler{repo: &salaryRepo{db: pool}},
		&titleHandler{repo: &titleRepo{db: pool}},
	)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("[server] listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("[server] failed to start: %v", err)
	}
}
