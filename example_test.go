package fixify_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/qawatake/fixify"
	"github.com/qawatake/fixify/internal/example/model"
)

func Example() {
	// t is passed from the test function.
	t := &testing.T{}
	// specify how to connect models in a declarative way.
	f := fixify.New(t,
		Company().With(
			Department("finance").With(
				Employee(),
			),
			Department("sales"),
		),
	)
	// Apply applies visitor function to each model and then call the connector functions.
	f.Apply(func(v any) error {
		switch v := v.(type) {
		case *model.Company:
			v.ID = 1
		case *model.Department:
			if v.Name == "finance" {
				v.ID = 2
			} else {
				v.ID = 3
			}
		case *model.Employee:
			v.ID = 4
		}
		return nil
	})
	allModels := f.All()
	for _, company := range filter[*model.Company](allModels) {
		fmt.Printf("CompanyID: %d\n", company.ID)
	}
	for _, department := range sortDepartments(filter[*model.Department](allModels)) {
		fmt.Printf("DepartmentID: %d Name: %s CompanyID: %d\n", department.ID, department.Name, department.CompanyID)
	}
	for _, employee := range filter[*model.Employee](allModels) {
		fmt.Printf("EmployeeID: %d DepartmentID: %d\n", employee.ID, employee.DepartmentID)
	}
	// Output:
	// CompanyID: 1
	// DepartmentID: 2 Name: finance CompanyID: 1
	// DepartmentID: 3 Name: sales CompanyID: 1
	// EmployeeID: 4 DepartmentID: 2
}

// Company represents a fixture for the company model.
func Company() *fixify.Model[model.Company] {
	// Company is the root model, so it does not need a connector function.
	return fixify.NewModel(new(model.Company))
}

// Department represents a fixture for the department model.
func Department(name string) *fixify.Model[model.Department] {
	d := &model.Department{
		Name: name,
	}
	return fixify.NewModel(d,
		// specify how to connect a department to a company.
		fixify.ConnectorFunc(func(_ testing.TB, department *model.Department, company *model.Company) {
			department.CompanyID = company.ID
		}),
	)
}

// Employee represents a fixture for the employee model.
func Employee() *fixify.Model[model.Employee] {
	return fixify.NewModel(new(model.Employee),
		// specify how to connect an employee to a department.
		fixify.ConnectorFunc(func(_ testing.TB, employee *model.Employee, department *model.Department) {
			employee.DepartmentID = department.ID
		}),
	)
}

func filter[T any](models []any) []T {
	filtered := make([]T, 0, len(models))
	for _, v := range models {
		if v, ok := v.(T); ok {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func sortDepartments(departments []*model.Department) []*model.Department {
	slices.SortFunc(departments, func(a, b *model.Department) int {
		return strings.Compare(a.Name, b.Name)
	})
	return departments
}
