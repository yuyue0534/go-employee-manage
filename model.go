package main

import "time"

// ─── Employee ────────────────────────────────────────────────

type Employee struct {
	EmpNo     int       `json:"emp_no"`
	BirthDate time.Time `json:"birth_date"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Gender    string    `json:"gender"`
	HireDate  time.Time `json:"hire_date"`
}

type CreateEmployeeReq struct {
	BirthDate string `json:"birth_date" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Gender    string `json:"gender"     binding:"required,oneof=M F"`
	HireDate  string `json:"hire_date"  binding:"required"`
}

type UpdateEmployeeReq struct {
	BirthDate string `json:"birth_date"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender" binding:"omitempty,oneof=M F"`
	HireDate  string `json:"hire_date"`
}

type EmployeeDetail struct {
	Employee
	Department    *Department `json:"department,omitempty"`
	CurrentTitle  *string     `json:"current_title,omitempty"`
	CurrentSalary *int        `json:"current_salary,omitempty"`
}

// ─── Department ──────────────────────────────────────────────

type Department struct {
	DeptNo   string `json:"dept_no"`
	DeptName string `json:"dept_name"`
}

type CreateDepartmentReq struct {
	DeptNo   string `json:"dept_no"   binding:"required"`
	DeptName string `json:"dept_name" binding:"required"`
}

type UpdateDepartmentReq struct {
	DeptName string `json:"dept_name" binding:"required"`
}

type AssignManagerReq struct {
	EmpNo    int    `json:"emp_no"    binding:"required"`
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date"   binding:"required"`
}

// ─── Salary ──────────────────────────────────────────────────

type Salary struct {
	EmpNo    int       `json:"emp_no"`
	Amount   int       `json:"amount"`
	FromDate time.Time `json:"from_date"`
	ToDate   time.Time `json:"to_date"`
}

type CreateSalaryReq struct {
	Amount   int    `json:"amount"    binding:"required,gt=0"`
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date"   binding:"required"`
}

type UpdateSalaryReq struct {
	Amount int `json:"amount" binding:"required,gt=0"`
}

// ─── Title ───────────────────────────────────────────────────

type Title struct {
	EmpNo    int        `json:"emp_no"`
	Title    string     `json:"title"`
	FromDate time.Time  `json:"from_date"`
	ToDate   *time.Time `json:"to_date,omitempty"`
}

type AssignTitleReq struct {
	Title    string `json:"title"     binding:"required"`
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date"`
}
