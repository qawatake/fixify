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
	// Iterate applies visitor function to each model and connect it to its children in the topological order.
	f.Iterate(setter)
	// finally, run the test!
}
```

For more examples, please refer to the [godoc].

## References

- [Goでテストのフィクスチャをいい感じに書く](https://engineering.mercari.com/blog/entry/20220411-42fc0ba69c/)

<!-- links -->

[godoc]: https://pkg.go.dev/github.com/qawatake/fixify
