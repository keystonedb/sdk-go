package models

import "github.com/keystonedb/sdk-go/keystone"

// Role is the deepest level (level 3)
type Role struct {
	Title       string `keystone:"title"`
	Department  string `keystone:"department"`
	YearsInRole int64  `keystone:"years_in_role"`
}

// CEO is the middle level (level 2) containing a Role
type CEO struct {
	Name         string `keystone:"name"`
	Email        string `keystone:"email"`
	YearsWithOrg int64  `keystone:"years_with_org"`
	CurrentRole  Role   `keystone:"current_role"`
	Salary       int64  `keystone:"salary"`
}

// Business is the top level (level 1) containing a CEO
type Business struct {
	keystone.BaseEntity
	CompanyName   string `keystone:"company_name,indexed"`
	Industry      string `keystone:"industry"`
	EmployeeCount int64  `keystone:"employee_count"`
	CEO           CEO    `keystone:"ceo"`
	AnnualRevenue int64  `keystone:"annual_revenue"`
}
