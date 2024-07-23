package fixify_test

import (
	"testing"

	"github.com/qawatake/fixify"
	"github.com/qawatake/fixify/model"
)

func Book() *fixify.ModelConnectorImpl[*model.Book] {
	book := new(model.Book)
	return fixify.NewModelConnector(book,
		fixify.ConnectParentFunc(func(t testing.TB, childModel *model.Book, parentModel *model.Library) {
			childModel.LibraryID = parentModel.ID
		}),
	)
}

func Page() *fixify.ModelConnectorImpl[*model.Page] {
	page := new(model.Page)
	return fixify.NewModelConnector(page,
		fixify.ConnectParentFunc(func(t testing.TB, childModel *model.Page, parentModel *model.Book) {
			childModel.BookID = parentModel.ID
		}),
	)
}

func Library() *fixify.ModelConnectorImpl[*model.Library] {
	return fixify.NewModelConnector(new(model.Library))
}
