# fixify

`fixify` is a Go library that helps you to write test fixtures in a declarative way.

```go
func TestRun(t *testing.T) {
	// specify how to connect models in the declarative way.
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
	// Iterate applies visitor function to each model and then call the connector functions.
	f.Iterate(setter)
	// retrieve all models.
	allModels := f.All()
	// assertion here
	// ...
}
```

## References

- [Goでテストのフィクスチャをいい感じに書く](https://engineering.mercari.com/blog/entry/20220411-42fc0ba69c/)
