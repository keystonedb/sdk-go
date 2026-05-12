package nested_property_loading

import (
	"context"
	"errors"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

var (
	CompanyName   = "NotationCorp Ltd"
	Industry      = "Research"
	EmployeeCount = int64(42)
	AnnualRevenue = int64(1234567)

	CEOName      = "Ada Lovelace"
	CEOEmail     = "ada@notationcorp.com"
	YearsWithOrg = int64(7)
	CEOSalary    = int64(250000)

	RoleTitle      = "Chief Executive Officer"
	RoleDepartment = "Executive"
	YearsInRole    = int64(3)
)

type Requirement struct {
	createdID keystone.ID
}

func (r *Requirement) Name() string {
	return "Nested Property Loading Notation"
}

func (r *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Business{})
	return nil
}

func (r *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(r.create(actor))
	report(r.loadWholeNestedStruct(actor))
	report(r.loadSingleFieldOnNestedStruct(actor))
	report(r.loadWholeDeeplyNestedStruct(actor))
	report(r.loadSingleFieldOnDeeplyNestedStruct(actor))
	report(r.loadMixedTopLevelAndNestedPaths(actor))
	report(r.selectivityUnrequestedFieldsRemainZero(actor))
}

func (r *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{
		CompanyName:   CompanyName,
		Industry:      Industry,
		EmployeeCount: EmployeeCount,
		AnnualRevenue: AnnualRevenue,
		CEO: models.CEO{
			Name:         CEOName,
			Email:        CEOEmail,
			YearsWithOrg: YearsWithOrg,
			Salary:       CEOSalary,
			CurrentRole: models.Role{
				Title:       RoleTitle,
				Department:  RoleDepartment,
				YearsInRole: YearsInRole,
			},
		},
	}

	err := actor.Mutate(context.Background(), business,
		keystone.WithMutationComment("Create business for nested property loading"))
	if err == nil {
		r.createdID = business.GetKeystoneID()
	}

	return requirements.TestResult{Name: "Create", Error: err}
}

// loadWholeNestedStruct verifies WithProperties("ceo") hydrates every field on
// the level-2 struct, including its level-3 child.
func (r *Requirement) loadWholeNestedStruct(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("ceo."))

	if err == nil {
		if business.CEO.Name != CEOName {
			err = fmt.Errorf("expected CEO.Name=%q, got %q", CEOName, business.CEO.Name)
		} else if business.CEO.Email != CEOEmail {
			err = fmt.Errorf("expected CEO.Email=%q, got %q", CEOEmail, business.CEO.Email)
		} else if business.CEO.Salary != CEOSalary {
			err = fmt.Errorf("expected CEO.Salary=%d, got %d", CEOSalary, business.CEO.Salary)
		} else if business.CEO.CurrentRole.Title != RoleTitle {
			err = fmt.Errorf("expected nested Role.Title=%q, got %q", RoleTitle, business.CEO.CurrentRole.Title)
		} else if business.CEO.CurrentRole.YearsInRole != YearsInRole {
			err = fmt.Errorf("expected nested Role.YearsInRole=%d, got %d", YearsInRole, business.CEO.CurrentRole.YearsInRole)
		}
	}

	return requirements.TestResult{Name: `WithProperties("ceo") loads whole nested struct`, Error: err}
}

// loadSingleFieldOnNestedStruct verifies WithProperties("ceo.name") hydrates
// only that field on the nested struct and leaves siblings zero.
func (r *Requirement) loadSingleFieldOnNestedStruct(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("ceo.name"))

	if err == nil {
		if business.CEO.Name != CEOName {
			err = fmt.Errorf("expected CEO.Name=%q, got %q", CEOName, business.CEO.Name)
		} else if business.CEO.Email != "" {
			err = fmt.Errorf("expected CEO.Email to be zero, got %q", business.CEO.Email)
		} else if business.CEO.Salary != 0 {
			err = fmt.Errorf("expected CEO.Salary to be zero, got %d", business.CEO.Salary)
		} else if business.CEO.CurrentRole.Title != "" {
			err = fmt.Errorf("expected Role.Title to be zero, got %q", business.CEO.CurrentRole.Title)
		}
	}

	return requirements.TestResult{Name: `WithProperties("ceo.name") loads only that field`, Error: err}
}

// loadWholeDeeplyNestedStruct verifies WithProperties("ceo.current_role")
// hydrates every field of the level-3 struct via dot-path.
func (r *Requirement) loadWholeDeeplyNestedStruct(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("ceo.current_role."))

	if err == nil {
		if business.CEO.CurrentRole.Title != RoleTitle {
			err = fmt.Errorf("expected Role.Title=%q, got %q", RoleTitle, business.CEO.CurrentRole.Title)
		} else if business.CEO.CurrentRole.Department != RoleDepartment {
			err = fmt.Errorf("expected Role.Department=%q, got %q", RoleDepartment, business.CEO.CurrentRole.Department)
		} else if business.CEO.CurrentRole.YearsInRole != YearsInRole {
			err = fmt.Errorf("expected Role.YearsInRole=%d, got %d", YearsInRole, business.CEO.CurrentRole.YearsInRole)
		} else if business.CEO.Name != "" {
			err = fmt.Errorf("expected CEO.Name to be zero (not requested), got %q", business.CEO.Name)
		}
	}

	return requirements.TestResult{Name: `WithProperties("ceo.current_role") loads whole deeply nested struct`, Error: err}
}

// loadSingleFieldOnDeeplyNestedStruct verifies WithProperties("ceo.current_role.title")
// targets one field three levels deep via dot-path.
func (r *Requirement) loadSingleFieldOnDeeplyNestedStruct(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("ceo.current_role.title"))

	if err == nil {
		if business.CEO.CurrentRole.Title != RoleTitle {
			err = fmt.Errorf("expected Role.Title=%q, got %q", RoleTitle, business.CEO.CurrentRole.Title)
		} else if business.CEO.CurrentRole.Department != "" {
			err = fmt.Errorf("expected Role.Department to be zero, got %q", business.CEO.CurrentRole.Department)
		} else if business.CEO.CurrentRole.YearsInRole != 0 {
			err = fmt.Errorf("expected Role.YearsInRole to be zero, got %d", business.CEO.CurrentRole.YearsInRole)
		} else if business.CEO.Name != "" {
			err = fmt.Errorf("expected CEO.Name to be zero, got %q", business.CEO.Name)
		}
	}

	return requirements.TestResult{Name: `WithProperties("ceo.current_role.title") loads only that field`, Error: err}
}

// loadMixedTopLevelAndNestedPaths verifies a single WithProperties call can
// mix top-level fields with cherry-picked nested fields.
func (r *Requirement) loadMixedTopLevelAndNestedPaths(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("company_name", "ceo.email", "ceo.current_role.years_in_role"))

	if err == nil {
		if business.CompanyName != CompanyName {
			err = fmt.Errorf("expected CompanyName=%q, got %q", CompanyName, business.CompanyName)
		} else if business.CEO.Email != CEOEmail {
			err = fmt.Errorf("expected CEO.Email=%q, got %q", CEOEmail, business.CEO.Email)
		} else if business.CEO.CurrentRole.YearsInRole != YearsInRole {
			err = fmt.Errorf("expected Role.YearsInRole=%d, got %d", YearsInRole, business.CEO.CurrentRole.YearsInRole)
		} else if business.Industry != "" {
			err = fmt.Errorf("expected Industry to be zero, got %q", business.Industry)
		} else if business.CEO.Name != "" {
			err = fmt.Errorf("expected CEO.Name to be zero, got %q", business.CEO.Name)
		} else if business.CEO.CurrentRole.Title != "" {
			err = fmt.Errorf("expected Role.Title to be zero, got %q", business.CEO.CurrentRole.Title)
		}
	}

	return requirements.TestResult{Name: `WithProperties mixes top-level and dot-path entries`, Error: err}
}

// selectivityUnrequestedFieldsRemainZero is an explicit guard that loading a
// deep field does not bleed sibling fields in from the same nested struct.
func (r *Requirement) selectivityUnrequestedFieldsRemainZero(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	err := actor.Get(context.Background(),
		keystone.ByEntityID(business, r.createdID), business,
		keystone.WithProperties("ceo.current_role.department"))

	if err == nil {
		if business.CEO.CurrentRole.Department != RoleDepartment {
			err = fmt.Errorf("expected Role.Department=%q, got %q", RoleDepartment, business.CEO.CurrentRole.Department)
		} else if business.CEO.CurrentRole.Title != "" {
			err = errors.New("Role.Title leaked despite requesting only ceo.current_role.department")
		} else if business.CEO.CurrentRole.YearsInRole != 0 {
			err = errors.New("Role.YearsInRole leaked despite requesting only ceo.current_role.department")
		} else if business.CompanyName != "" {
			err = errors.New("CompanyName leaked despite requesting only a deep nested field")
		}
	}

	return requirements.TestResult{Name: `Unrequested fields remain zero (selectivity guard)`, Error: err}
}
