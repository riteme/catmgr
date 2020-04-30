package main

import "time"

type MError struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func NewMError(message string) MError {
	return MError{"failed", message}
}

type MHello struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type MNewBook struct {
	Status string `json:"status"`
	BookID int    `json:"book_id"`
}

type MAddUser struct {
	Status string `json:"status"`
	UserID int    `json:"user_id"`
}

type MBookList struct {
	Status  string `json:"status"`
	Results []Book `json:"results"`
}

type MRecordList struct {
	Status  string   `json:"status"`
	Results []Record `json:"results"`
}

type MBook struct {
	Status string `json:"status"`
	BookID int    `json:"book_id"`
}

type MRecord struct {
	Status   string `json:"status"`
	RecordID int    `json:"record_id"`
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type Permission struct {
	Update  bool
	AddUser bool
	Borrow  bool
	Inspect bool
}

func (e Permission) mask() int {
	var ret int
	if e.Update {
		ret |= 1
	}
	if e.AddUser {
		ret |= 2
	}
	if e.Borrow {
		ret |= 4
	}
	if e.Inspect {
		ret |= 8
	}
	return ret
}

type Book struct {
	BookID         int    `json:"book_id"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	ISBN           string `json:"isbn"`
	AvailableCount int    `json:"count"`
	Description    string `json:"description"`
	Comment        string `json:"comment"`
}

// `BookInfo` is used by `UpdateBook`.
// nil values indicate that corresponding fields will not be updated.
type BookInfo struct {
	Title       *string
	Author      *string
	ISBN        *string
	Description *string
	Comment     *string
}

type Record struct {
	RecordID   int       `json:"record_id"`
	UserID     int       `json:"user_id"`
	BookID     int       `json:"book_id"`
	Returned   bool      `json:"returned"`
	ReturnDate time.Time `json:"return_date"`
	BorrowDate time.Time `json:"borrow_date"`
	DueDate    time.Time `json:"deadline"`
	FinalDate  time.Time `json:"final_deadline"`
}
