package fixify_test

import (
	"testing"

	"github.com/qawatake/fixify"
	"github.com/qawatake/fixify/model"
)

func Book() *fixify.Model[model.Book] {
	book := new(model.Book)
	return fixify.NewModel(book,
		fixify.ConnectorFunc(func(t testing.TB, childModel *model.Book, parentModel *model.Library) {
			childModel.LibraryID = parentModel.ID
		}),
	)
}

func Page() *fixify.Model[model.Page] {
	page := new(model.Page)
	return fixify.NewModel(page,
		fixify.ConnectorFunc(func(t testing.TB, childModel *model.Page, parentModel *model.Book) {
			childModel.BookID = parentModel.ID
		}),
	)
}

func Library() *fixify.Model[model.Library] {
	return fixify.NewModel(new(model.Library))
}
