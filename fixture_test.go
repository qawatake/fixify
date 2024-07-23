package fixify_test

import (
	"testing"

	"github.com/qawatake/fixify"
	"github.com/qawatake/fixify/model"
	"github.com/stretchr/testify/assert"
)

func TestModelConnectorImpl_With(t *testing.T) {
	t.Parallel()

	t.Run("normal", func(t *testing.T) {
		Library().With(
			Book().With(
				Page(),
			),
			Book(),
		)
		Library().With(
			Book(),
		)
	})

	t.Run("cyclic", func(t *testing.T) {
		library := Library()
		book := Book()
		book.With(library)
		assert.Panics(t, func() {
			library.With(book)
		})
	})
}

func TestModelConnectorImpl_Bind(t *testing.T) {
	var library *fixify.ModelConnectorImpl[*model.Library]
	f := fixify.New(t,
		Library().With(
			Book(),
		).Bind(&library),
		library.With(
			Book(),
		),
	)
	f.Iterate(func(v any) error {
		switch v := v.(type) {
		case *model.Library:
			v.ID = 1
		case *model.Book:
			v.ID = 2
		}
		return nil
	})
	{
		got := filter[*model.Library](f.All())
		assert.ElementsMatch(t, []*model.Library{{ID: 1}}, got)
	}
	{
		got := filter[*model.Book](f.All())
		assert.ElementsMatch(t, []*model.Book{{ID: 2, LibraryID: 1}, {ID: 2, LibraryID: 1}}, got)
	}
}

// func TestModelConnector_Children(t *testing.T) {
// 	t.Parallel()

// 	t.Run("no children", func(t *testing.T) {
// 		library := Library()
// 		assert.Empty(t, library.Children())
// 	})

// 	t.Run("normal", func(t *testing.T) {
// 		library := Library().With(
// 			Book().With(
// 				Page(),
// 			),
// 			Book(),
// 		)
// 		assert.Len(t, library.Children(), 2)
// 		assert.Len(t, extractModels[*model.Book](library.Children()), 2)
// 	})
// }

// func TestModelConnector_Descendants(t *testing.T) {
// 	t.Parallel()

// 	t.Run("no descendants", func(t *testing.T) {
// 		library := Library()
// 		assert.Empty(t, library.Descendants())
// 	})

// 	t.Run("normal", func(t *testing.T) {
// 		library := Library().With(
// 			Book().With(
// 				Page(),
// 			),
// 			Book(),
// 		)
// 		assert.Len(t, library.Descendants(), 3)
// 		assert.Len(t, extractModels[*model.Book](library.Descendants()), 2)
// 		assert.Len(t, extractModels[*model.Page](library.Descendants()), 1)
// 	})
// }

func TestBuild_and_Fixture_All(t *testing.T) {
	t.Run("no connectors", func(t *testing.T) {
		f := fixify.New(t)
		assert.Len(t, f.All(), 0)
	})

	t.Run("same connector", func(t *testing.T) {
		library := Library()
		f := fixify.New(t, library, library)
		assert.Len(t, f.All(), 1)
		assert.Len(t, filter[*model.Library](f.All()), 1)
	})

	t.Run("normal", func(t *testing.T) {
		f := fixify.New(t,
			Library().With(
				Book().With(
					Page(),
				),
				Book(),
			),
			Library().With(
				Book(),
			),
		)
		assert.Len(t, f.All(), 6)
		assert.Len(t, filter[*model.Library](f.All()), 2)
		assert.Len(t, filter[*model.Book](f.All()), 3)
		assert.Len(t, filter[*model.Page](f.All()), 1)
	})
}

func TestFixture_Iterate(t *testing.T) {
	f := fixify.New(t,
		Library().With(
			Book().With(
				Page(),
			),
			Book(),
		),
		Library().With(
			Book(),
		),
	)
	f.Iterate(func(v any) error {
		switch v := v.(type) {
		case *model.Library:
			v.ID = 1
		case *model.Book:
			v.ID = 2
		case *model.Page:
			v.ID = 3
		}
		return nil
	})
	{
		got := filter[*model.Library](f.All())
		assert.ElementsMatch(t, []*model.Library{{ID: 1}, {ID: 1}}, got)
	}
	{
		got := filter[*model.Book](f.All())
		assert.ElementsMatch(t, []*model.Book{{ID: 2, LibraryID: 1}, {ID: 2, LibraryID: 1}, {ID: 2, LibraryID: 1}}, got)
	}
	{
		got := filter[*model.Page](f.All())
		assert.ElementsMatch(t, []*model.Page{{ID: 3, BookID: 2}}, got)
	}
}

func filter[V any](values []any) []V {
	results := make([]V, 0, len(values))
	for _, v := range values {
		if v, ok := v.(V); ok {
			results = append(results, v)
		}
	}
	return results
}
