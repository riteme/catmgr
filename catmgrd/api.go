package main

import "fmt"
import "time"
import "errors"
import "strings"
import "database/sql"
import "crypto/sha1"

var (
	day   = time.Hour * 24
	week  = day * 7
	month = day * 30
	year  = day * 365
)

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
	BookID         int
	Title          string
	Author         string
	ISBN           string
	AvailableCount int
	Description    string
	Comment        string
}

type BookInfo struct {
	Title       *string
	Author      *string
	ISBN        *string
	Description *string
	Comment     *string
}

type Record struct {
	RecordID   int
	UserID     int
	BookID     int
	Returned   bool
	ReturnDate time.Time
	BorrowDate time.Time
	DueDate    time.Time
	FinalDate  time.Time
}

var ErrInvalidUserID = errors.New("Invalid user ID")
var ErrInvalidPassword = errors.New("Invalid password")
var ErrPermissionDenied = errors.New("Permission denied")

// `AuthUser` check `login` information against table User
// in database `db`, which stores the sha1 hashes of passwords.
// Requested permissions `req` are encapsulated in Permission struct.
// Auth success if no error returned.
//
// May return `ErrInvalidUserID`, `ErrInvalidPassword` or
// `ErrPermissionDenied`.
func AuthUser(db *sql.DB, user_id int, password string, req Permission) error {
	hash_bytes := sha1.Sum([]byte(password))
	hash := fmt.Sprintf("%x", hash_bytes)

	var token string
	var perm Permission
	query := `SELECT token, can_update, can_adduser, can_borrow, can_inspect
		FROM User JOIN UserType USING (type_id)
		WHERE user_id=?`
	err := db.QueryRow(query, user_id).
		Scan(&token, &perm.Update, &perm.AddUser, &perm.Borrow, &perm.Inspect)
	if err == sql.ErrNoRows {
		return ErrInvalidUserID
	}
	if err != nil {
		return err
	}

	if hash != token {
		return ErrInvalidPassword
	}

	if req.mask()&perm.mask() != req.mask() {
		return ErrPermissionDenied
	}

	return nil
}

var ErrInvalidUserType = errors.New("Invalid user type name/ID")

// `GetUserTypeID` returns `type_id` of `type_name` defined in
// UserType table.
//
// Returns `ErrInvalidUserType` when `type_name` is not found
// in table UserType.
func GetUserTypeID(db *sql.DB, type_name string) (int, error) {
	var type_id int
	query := `SELECT type_id FROM UserType WHERE type_name=?`
	err := db.QueryRow(query, type_name).Scan(&type_id)
	if err == sql.ErrNoRows {
		return -1, ErrInvalidUserType
	}
	if err != nil {
		return -1, err
	}

	return type_id, nil
}

// `AddUser` simply insert a new user record into User table.
//
// Returns the ID of newly added user.
func AddUser(db *sql.DB, type_id int, username string, password string) (int, error) {
	token := fmt.Sprintf("%x", sha1.Sum([]byte(password)))
	result, err := db.Exec(
		"INSERT INTO User (type_id, name, token) VALUES (?, ?, ?)",
		type_id, username, token,
	)
	if err != nil {
		return -1, err
	}

	user_id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(user_id), nil
}

var selectBook = `
SELECT
	book_id,
	COALESCE(title, '(no title)'),
	COALESCE(author, '(no author)'),
	COALESCE(isbn, '(no isbn)'),
	available_count,
	COALESCE(description, '(no description)'),
	COALESCE(comment, '(no comment)')
FROM Book
WHERE `

var ErrBookNotFound = errors.New("Book not found")

func scanBook(row *sql.Row, book *Book) error {
	err := row.Scan(
		&book.BookID, &book.Title,
		&book.Author, &book.ISBN,
		&book.AvailableCount,
		&book.Description,
		&book.Comment,
	)
	if err == sql.ErrNoRows {
		return ErrBookNotFound
	}

	return err
}

// `CheckoutBook` obtains book information with id `book_id`.
//
// Book information is stored in struct `Book`. When no book
// matches `book_id`, an `ErrBookNotFound` is returned.
func CheckoutBook(db *sql.DB, book_id int) (Book, error) {
	var book Book
	row := db.QueryRow(selectBook+"book_id=?", book_id)

	err := scanBook(row, &book)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

// `CheckoutISBN` obtains book information with `isbn`. Similar to
// `CheckoutBook`.
//
// Book information is stored in struct `Book`. When no book
// matches `book_id`, an `ErrBookNotFound` is returned.
func CheckoutISBN(db *sql.DB, isbn string) (Book, error) {
	var book Book
	row := db.QueryRow(selectBook+"isbn=?", isbn)

	err := scanBook(row, &book)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

var ErrInvalidRecordID = errors.New("Invalid record ID")

// CheckoutRecord retrieves record with `record_id`.
//
// Record information is stored in struct `Record`. When no record
// matches `record_id`, an `ErrInvalidRecordID` is returned.
func CheckoutRecord(db *sql.DB, record_id int) (Record, error) {
	var r Record
	var return_date sql.NullTime
	err := db.QueryRow(`
		SELECT
			record_id, user_id, book_id,
			return_date, borrow_date,
			deadline, final_deadline
		FROM Record
		WHERE record_id = ?`, record_id).
		Scan(
			&r.RecordID, &r.UserID, &r.BookID,
			&return_date, &r.BorrowDate,
			&r.DueDate, &r.FinalDate,
		)
	if err == sql.ErrNoRows {
		return Record{}, ErrInvalidRecordID
	}
	if err != nil {
		return Record{}, err
	}

	if return_date.Valid {
		r.Returned = true
		r.ReturnDate = return_date.Time
	}

	return r, err
}

var ErrNoAvailableBook = errors.New("No available book")
var ErrInvalidBookID = errors.New("Invalid book ID")
var ErrSuspendedUser = errors.New("User suspended: you have more than 3 overdue books")

// BorrowBook attempts to borrow a book with `book_id` and add a record.
//
// The ID of newly added record is returned when success.
// If there is no available book, `ErrNoAvailableBook` is returned.
// If no book has `book_id`, `ErrInvalidBookID` is returned.
// If the user with `user_id` has more than 3 overdue book records,
// `BorrowBook` rejects this request.
func BorrowBook(db *sql.DB, user_id, book_id int) (int, error) {
	var tmp int
	err := db.QueryRow(
		"SELECT book_id FROM Book WHERE book_id = ?", book_id).
		Scan(&tmp)
	if err == sql.ErrNoRows {
		return -1, ErrInvalidBookID
	}
	if err != nil {
		return -1, err
	}

	now := time.Now()
	var overdue_count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM Record
		WHERE
			user_id = ? AND
			return_date IS NULL AND
			deadline < ?`,
		user_id, now).
		Scan(&overdue_count)
	if err != nil {
		return -1, err
	}
	if overdue_count > 3 {
		return -1, ErrSuspendedUser
	}

	due := now.Add(month)
	final := now.Add(3 * month)

	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	// try decreasing available count
	result, err := tx.Exec(`
		UPDATE Book
		SET
			available_count = available_count - 1
		WHERE
			book_id = ? AND
			available_count > 0`, book_id)
	if err != nil {
		return -1, err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}
	if cnt == 0 {
		return -1, ErrNoAvailableBook
	}

	result, err = tx.Exec(`
		INSERT INTO Record
			(user_id, book_id, borrow_date, deadline, final_deadline)
		VALUES (?, ?, ?, ?, ?)`,
		user_id, book_id, now, due, final)
	if err != nil {
		return -1, err
	}

	record_id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	tx.Commit()
	return int(record_id), nil
}

var ErrAlreadyReturned = errors.New("This book has been returned")
var ErrOverdue = errors.New("Cannot extend deadline for overdue records")
var ErrNotExtensible = errors.New("Do extend deadline in the last week")
var ErrFinalDeadline = errors.New("Must return book in three months")

// `ExtendDeadline` tries to extend deadline of a specific record
// with `record_id` for a month. Deadlines are not allowed to be later than
// final deadlines, in which case an `ErrFinalDeadline` will be returned.
//
// `ErrInvalidRecordID` occurs when no record matches `record_id`.
//
// Attempting to extend deadline for returned or overdue records is invalid, which
// will result in `ErrAlreadyReturned` and `ErrOverdue` respectively.
// Extending deadline is limited within the last week before deadline,
// and an `ErrNotExtensible` will be returned otherwise.
//
// NOTE: this function does not check `user_id`. Anyone who knows
// `record_id` can do this.
func ExtendDeadline(db *sql.DB, record_id int) error {
	var is_null int
	var due, final time.Time
	err := db.QueryRow(
		`SELECT
			ISNULL(return_date),
			deadline,
			final_deadline
		FROM Record WHERE record_id=?`,
		record_id).
		Scan(&is_null, &due, &final)
	if err == sql.ErrNoRows {
		return ErrInvalidRecordID
	}
	if err != nil {
		return err
	}

	if is_null == 0 {
		return ErrAlreadyReturned
	}

	now := time.Now()
	if due.Before(now) {
		return ErrOverdue
	}

	next_week := now.Add(week)
	if next_week.Before(due) {
		return ErrNotExtensible
	}

	new_due := due.Add(month)
	if final.Before(new_due) {
		return ErrFinalDeadline
	}

	_, err = db.Exec(`
		UPDATE Record
		SET deadline = ?
		WHERE record_id = ?`, new_due, record_id)
	if err != nil {
		return err
	}
	return nil
}

// `ReturnBook` returns book for record with `record_id`.
//
// If no record matches `record_id`, an `ErrInvaildRecordID` is returned.
// If the record is marked as "returned", an `ErrAlreadyReturned` is returned.
//
// NOTE: this function does not check `user_id`.
func ReturnBook(db *sql.DB, record_id int) error {
	var is_null int
	err := db.QueryRow("SELECT ISNULL(return_date) FROM Record WHERE record_id=?", record_id).
		Scan(&is_null)
	if err == sql.ErrNoRows {
		return ErrInvalidRecordID
	}
	if err != nil {
		return err
	}
	if is_null == 0 {
		return ErrAlreadyReturned
	}

	var book_id int
	err = db.QueryRow("SELECT book_id FROM Record WHERE record_id=?", record_id).
		Scan(&book_id)
	if err != nil { // must be valid record ID
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec("UPDATE Record SET return_date=? WHERE record_id=?", now, record_id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE Book
		SET
			available_count = available_count + 1
		WHERE book_id = ?`, book_id)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// `NewBook` simply insert a new book record into table Book.
// More information can be added by `UpdateBook`.
// Book's available count is initially 0.
func NewBook(db *sql.DB) (int, error) {
	result, err := db.Exec("INSERT INTO Book SET available_count = 0")
	if err != nil {
		return -1, err
	}

	book_id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(book_id), nil
}

func UpdateBook(db *sql.DB, book_id, delta_cnt int, info BookInfo) error {
	var buf strings.Builder
	buf.WriteString("UPDATE Book SET ")

	var args []interface{}
	if info.Author != nil {
		buf.WriteString("author=?,")
		args = append(args, info.Author)
	}
	if info.Comment != nil {
		buf.WriteString("comment=?,")
		args = append(args, info.Comment)
	}
	if info.Description != nil {
		buf.WriteString("description=?,")
		args = append(args, info.Description)
	}
	if info.ISBN != nil {
		buf.WriteString("isbn=?,")
		args = append(args, info.ISBN)
	}
	if info.Title != nil {
		buf.WriteString("title=?,")
		args = append(args, info.Title)
	}

	buf.WriteString("available_count=available_count+(?) ")
	buf.WriteString("WHERE book_id=?")
	args = append(args, delta_cnt)
	args = append(args, book_id)

	_, err := db.Exec(buf.String(), args...)
	return err
}

// TODO:
// * search by title/author
// * query borrow history with filtering
