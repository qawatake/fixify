# fixify

[![Go Reference](https://pkg.go.dev/badge/github.com/qawatake/fixify.svg)](https://pkg.go.dev/github.com/qawatake/fixify)
[![test](https://github.com/qawatake/fixify/actions/workflows/test.yaml/badge.svg)](https://github.com/qawatake/fixify/actions/workflows/test.yaml)

`fixify` is a Go library that helps you to write test fixtures in a declarative way.

```go
func TestRun(t *testing.T) {
	// specify how to connect models in a declarative way.
	f := fixify.New(t,
		Company().With(
			Department("finance").With(
				Employee(),
				Employee(),
			),
			Department("sales").With(
				Employee(),
				Employee(),
				Employee(),
			),
		),
	)
	// Apply resolves the dependencies and applies a visitor to each model.
	f.Apply(setter)
	// finally, run the test!
}

// Department is a fixture for model.Department.
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
```

For more examples, please refer to the [godoc].

## References

- [Goでテストのフィクスチャをいい感じに書く](https://engineering.mercari.com/blog/entry/20220411-42fc0ba69c/)

<!-- links -->

[godoc]: https://pkg.go.dev/github.com/qawatake/fixify
