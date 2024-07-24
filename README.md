# fixify

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
	// Iterate applies visitor function to each model and connect it to its children in the topological order.
	f.Iterate(setter)
	// finally, run the test!
}
```

## References

- [Goでテストのフィクスチャをいい感じに書く](https://engineering.mercari.com/blog/entry/20220411-42fc0ba69c/)
