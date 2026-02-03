package nested_structs

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

var (
	// Level 1 - Business
	CompanyName   = "TechCorp Inc"
	Industry      = "Technology"
	EmployeeCount = int64(500)
	AnnualRevenue = int64(50000000)

	// Level 2 - CEO
	CEOName      = "Jane Smith"
	CEOEmail     = "jane.smith@techcorp.com"
	YearsWithOrg = int64(10)
	CEOSalary    = int64(500000)

	// Level 3 - Role
	RoleTitle      = "Chief Executive Officer"
	RoleDepartment = "Executive"
	YearsInRole    = int64(5)

	// Updated values
	UpdatedCompanyName   = "TechCorp International"
	UpdatedCEOName       = "Jane Smith-Johnson"
	UpdatedRoleTitle     = "CEO & Chairman"
	UpdatedAnnualRevenue = int64(75000000)
	UpdatedYearsInRole   = int64(6)
)

type Requirement struct {
	createdID keystone.ID
}

func (r *Requirement) Name() string {
	return "Nested Structs CRU"
}

func (r *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Business{})
	return nil
}

func (r *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		r.create(actor),
		r.read(actor),
		r.updateLevel1(actor),
		r.updateLevel2(actor),
		r.updateLevel3(actor),
		r.updateMultipleLevels(actor),
	}
}

func (r *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{
		BaseEntity:    keystone.BaseEntity{},
		CompanyName:   CompanyName,
		Industry:      Industry,
		EmployeeCount: EmployeeCount,
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
		AnnualRevenue: AnnualRevenue,
	}

	createErr := actor.Mutate(context.Background(), business, keystone.WithMutationComment("Create business with nested CEO and Role"))
	if createErr == nil {
		r.createdID = business.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (r *Requirement) read(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(business, r.createdID), business, keystone.WithProperties())

	if getErr == nil {
		// Verify Level 1 (Business)
		if business.CompanyName != CompanyName {
			getErr = errors.New("company name mismatch")
		} else if business.Industry != Industry {
			getErr = errors.New("industry mismatch")
		} else if business.EmployeeCount != EmployeeCount {
			getErr = errors.New("employee count mismatch")
		} else if business.AnnualRevenue != AnnualRevenue {
			getErr = errors.New("annual revenue mismatch")
		} else if business.CEO.Name != CEOName {
			// Verify Level 2 (CEO)
			getErr = errors.New("CEO name mismatch")
		} else if business.CEO.Email != CEOEmail {
			getErr = errors.New("CEO email mismatch")
		} else if business.CEO.YearsWithOrg != YearsWithOrg {
			getErr = errors.New("CEO years with org mismatch")
		} else if business.CEO.Salary != CEOSalary {
			getErr = errors.New("CEO salary mismatch")
		} else if business.CEO.CurrentRole.Title != RoleTitle {
			// Verify Level 3 (Role)
			getErr = errors.New("role title mismatch")
		} else if business.CEO.CurrentRole.Department != RoleDepartment {
			getErr = errors.New("role department mismatch")
		} else if business.CEO.CurrentRole.YearsInRole != YearsInRole {
			getErr = errors.New("role years in role mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}

func (r *Requirement) updateLevel1(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	business.SetKeystoneID(r.createdID)
	business.CompanyName = UpdatedCompanyName

	updateErr := actor.Mutate(context.Background(), business,
		keystone.WithMutationComment("Update level 1 - company name"),
		keystone.MutateProperties("company_name"))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(business, r.createdID), business, keystone.WithProperties())
		if updateErr == nil {
			if business.CompanyName != UpdatedCompanyName {
				updateErr = errors.New("level 1 update failed: company name mismatch")
			}
		}
	}

	return requirements.TestResult{
		Name:  "Update Level 1 (Business)",
		Error: updateErr,
	}
}

func (r *Requirement) updateLevel2(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	business.SetKeystoneID(r.createdID)
	business.CEO.Name = UpdatedCEOName

	updateErr := actor.Mutate(context.Background(), business,
		keystone.WithMutationComment("Update level 2 - CEO name"),
		keystone.MutateProperties("ceo"))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(business, r.createdID), business, keystone.WithProperties())
		if updateErr == nil {
			if business.CEO.Name != UpdatedCEOName {
				updateErr = errors.New("level 2 update failed: CEO name mismatch")
			}
		}
	}

	return requirements.TestResult{
		Name:  "Update Level 2 (CEO)",
		Error: updateErr,
	}
}

func (r *Requirement) updateLevel3(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	business.SetKeystoneID(r.createdID)
	business.CEO.CurrentRole.Title = UpdatedRoleTitle
	business.CEO.CurrentRole.YearsInRole = UpdatedYearsInRole

	updateErr := actor.Mutate(context.Background(), business,
		keystone.WithMutationComment("Update level 3 - Role title and years"),
		keystone.MutateProperties("ceo"))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(business, r.createdID), business, keystone.WithProperties())
		if updateErr == nil {
			if business.CEO.CurrentRole.Title != UpdatedRoleTitle {
				updateErr = errors.New("level 3 update failed: role title mismatch, got: " + business.CEO.CurrentRole.Title)
			} else if business.CEO.CurrentRole.YearsInRole != UpdatedYearsInRole {
				updateErr = errors.New("level 3 update failed: years in role mismatch")
			}
		}
	}

	return requirements.TestResult{
		Name:  "Update Level 3 (Role)",
		Error: updateErr,
	}
}

func (r *Requirement) updateMultipleLevels(actor *keystone.Actor) requirements.TestResult {
	business := &models.Business{}
	business.SetKeystoneID(r.createdID)
	business.AnnualRevenue = UpdatedAnnualRevenue
	business.CEO.Salary = CEOSalary + 50000
	business.CEO.CurrentRole.YearsInRole = UpdatedYearsInRole + 1

	updateErr := actor.Mutate(context.Background(), business,
		keystone.WithMutationComment("Update multiple levels simultaneously"),
		keystone.MutateProperties("annual_revenue", "ceo"))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(business, r.createdID), business, keystone.WithProperties())
		if updateErr == nil {
			if business.AnnualRevenue != UpdatedAnnualRevenue {
				updateErr = errors.New("multi-level update failed: annual revenue mismatch")
			} else if business.CEO.Salary != CEOSalary+50000 {
				updateErr = errors.New("multi-level update failed: CEO salary mismatch")
			} else if business.CEO.CurrentRole.YearsInRole != UpdatedYearsInRole+1 {
				updateErr = errors.New("multi-level update failed: years in role mismatch")
			}
		}
	}

	return requirements.TestResult{
		Name:  "Update Multiple Levels",
		Error: updateErr,
	}
}
