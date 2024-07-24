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

type Company struct {
	ID int64
}

type Department struct {
	ID        int64
	CompanyID int64
	Name      string
}

type Employee struct {
	ID           int64
	DepartmentID int64
}

type Classroom struct {
	ID int64
}

type Student struct {
	ID int64
}

type Enrollment struct {
	ID          int64
	StudentID   int64
	ClassroomID int64
}

type User struct {
	ID   int64
	Name string
}

type Follow struct {
	ID         int64
	FollowerID int64
	FolloweeID int64
}
