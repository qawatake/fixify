package fixify_test

import (
	"testing"

	"github.com/qawatake/fixify"
	"github.com/qawatake/fixify/internal/example/model"
	"github.com/stretchr/testify/assert"
)

func ExampleNewModel() {
	newFixtureBook := func() *fixify.Model[model.Book] {
		book := new(model.Book)
		return fixify.NewModel(book,
			// specify how to connect a book to a library.
			fixify.ConnectorFunc(func(_ testing.TB, book *model.Book, library *model.Library) {
				book.LibraryID = library.ID
			}),
			// specify how to connect a book to an author.
			fixify.ConnectorFunc(func(_ testing.TB, book *model.Book, author *model.Author) {
				book.AuthorID = author.ID
			}),
		)
	}
	newFixtureBook()
	// Output:
}

func ExampleModel_With() {
	// Company, Department, and Employee are fixtures for the company, department, and employee models.
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
	)
	// Output:
}

func ExampleModel_Bind() {
	// t is passed from the test function.
	t := &testing.T{}
	var enrollment *fixify.Model[model.Enrollment]
	fixify.New(t,
		Student().With(
			Enrollment().Bind(&enrollment),
		),
		Classroom().With(
			enrollment,
		),
	)
	// Output:
}

func TestModel_With(t *testing.T) {
	t.Parallel()
	t.Run("normal", func(t *testing.T) {
		t.Parallel()
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

	t.Run("try to connect to non-parent", func(t *testing.T) {
		t.Parallel()
		book := Book()
		assert.Panics(t, func() {
			book.With(Library())
		})
	})

	t.Run("cyclic", func(t *testing.T) {
		t.Parallel()
		library := Library()
		book := Book()
		library.With(book)
		assert.Panics(t, func() {
			book.With(library)
		})
	})
}

func TestModel_Bind(t *testing.T) {
	t.Parallel()
	var library *fixify.Model[model.Library]
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

func TestNew_and_Fixture_All(t *testing.T) {
	t.Parallel()
	t.Run("no connectors", func(t *testing.T) {
		t.Parallel()
		f := fixify.New(t)
		assert.Empty(t, f.All(), 0)
	})

	t.Run("same connector", func(t *testing.T) {
		t.Parallel()
		library := Library()
		f := fixify.New(t, library, library)
		assert.Len(t, f.All(), 1)
		assert.Len(t, filter[*model.Library](f.All()), 1)
	})

	t.Run("normal", func(t *testing.T) {
		t.Parallel()
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

	t.Run("labeled", func(t *testing.T) {
		t.Parallel()
		follow := Follow()
		f := fixify.New(t,
			User().With(
				follow.Label("follower"),
				follow.Label("followee"),
			),
		)
		assert.Len(t, f.All(), 2)
		assert.Len(t, filter[*model.User](f.All()), 1)
		assert.Len(t, filter[*model.Follow](f.All()), 1)
	})
}

func TestFixture_Iterate(t *testing.T) {
	t.Parallel()
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

// Book represents a fixture for the book model.
func Book() *fixify.Model[model.Book] {
	book := new(model.Book)
	return fixify.NewModel(book,
		// specify how to connect a book to a library.
		fixify.ConnectorFunc(func(_ testing.TB, book *model.Book, library *model.Library) {
			book.LibraryID = library.ID
		}),
	)
}

// Page represents a fixture for the page model.
func Page() *fixify.Model[model.Page] {
	page := new(model.Page)
	return fixify.NewModel(page,
		// specify how to connect a page to a book.
		fixify.ConnectorFunc(func(_ testing.TB, page *model.Page, book *model.Book) {
			page.BookID = book.ID
		}),
	)
}

// Library represents a fixture for the library model.
func Library() *fixify.Model[model.Library] {
	// library is the root model, so it does not need a connector function.
	return fixify.NewModel(new(model.Library))
}

func Student() *fixify.Model[model.Student] {
	return fixify.NewModel(new(model.Student))
}

func Classroom() *fixify.Model[model.Classroom] {
	return fixify.NewModel(new(model.Classroom))
}

func Enrollment() *fixify.Model[model.Enrollment] {
	return fixify.NewModel(new(model.Enrollment),
		fixify.ConnectorFunc(func(_ testing.TB, enrollment *model.Enrollment, student *model.Student) {
			enrollment.StudentID = student.ID
		}),
		fixify.ConnectorFunc(func(_ testing.TB, enrollment *model.Enrollment, classroom *model.Classroom) {
			enrollment.ClassroomID = classroom.ID
		}),
	)
}

func User() *fixify.Model[model.User] {
	return fixify.NewModel(new(model.User))
}

func Follow() *fixify.Model[model.Follow] {
	return fixify.NewModel(new(model.Follow),
		fixify.ConnectorFuncWithLabel("follower", func(_ testing.TB, follow *model.Follow, follower *model.User) {
			follow.FollowerID = follower.ID
		}),
		fixify.ConnectorFuncWithLabel("followee", func(_ testing.TB, follow *model.Follow, followee *model.User) {
			follow.FolloweeID = followee.ID
		}),
	)
}
