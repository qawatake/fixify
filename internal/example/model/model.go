package model

type Library struct {
	ID   int64
	Name string
}

type Book struct {
	ID        int64
	Name      string
	LibraryID int64
	AuthorID  int64
}

type Page struct {
	ID     int64
	Num    int
	BookID int64
}

type Author struct {
	ID   int64
	Name string
}
