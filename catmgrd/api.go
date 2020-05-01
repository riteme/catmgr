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

var ErrInvalidUser = errors.New("invalid username/user ID")
var ErrInvalidPassword = errors.New("invalid password")
var ErrPermissionDenied = errors.New("permission denied")

// `AuthUser` check `login` information against table User
// in database `db`, which stores the sha1 hashes of passwords.
// Requested permissions `req` are encapsulated in Permission struct.
// Auth success if no error returned.
//
// May return `ErrInvalidUser`, `ErrInvalidPassword` or
// `ErrPermissionDenied`.
func AuthUser(db *sql.DB, user interface{}, password string, req Permission) error {
	hash_bytes := sha1.Sum([]byte(password))
	hash := fmt.Sprintf("%x", hash_bytes)

	var token string
	var perm Permission
	query := `SELECT token, can_update, can_adduser, can_borrow, can_inspect
		FROM User JOIN UserType USING (type_id)
		WHERE `

	var row *sql.Row
	switch v := user.(type) {
	case int:
		row = db.QueryRow(query+"user_id = ?", v)
	case float64:
		user_id := int(v)
		row = db.QueryRow(query+"user_id = ?", user_id)
	case string:
		row = db.QueryRow(query+"name = ?", v)
	default:
		return ErrInvalidUser
	}

	err := row.Scan(&token, &perm.Update, &perm.AddUser, &perm.Borrow, &perm.Inspect)
	if err == sql.ErrNoRows {
		return ErrInvalidUser
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

func GetUserID(db *sql.DB, name string) (int, error) {
	var user_id int
	query := "SELECT user_id FROM User WHERE name=?"
	err := db.QueryRow(query, name).Scan(&user_id)
	if err == sql.ErrNoRows {
		return -1, ErrInvalidUser
	}
	if err != nil {
		return -1, err
	}
	return user_id, nil
}

var ErrInvalidUserType = errors.New("invalid user type name/ID")

// `GetUserTypeID` returns `type_id` of `type_name` defined in
// UserType table.
//
// Returns `ErrInvalidUserType` when `type_name` is not found
// in table UserType.
func GetUserTypeID(db *sql.DB, type_name string) (int, error) {
	var type_id int
	query := "SELECT type_id FROM UserType WHERE type_name=?"
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

var ErrBookNotFound = errors.New("book not found")

func scanBook(row RowScanner) (Book, error) {
	var book Book
	err := row.Scan(
		&book.BookID, &book.Title,
		&book.Author, &book.ISBN,
		&book.AvailableCount,
		&book.Description,
		&book.Comment,
	)
	if err == sql.ErrNoRows {
		return Book{}, ErrBookNotFound
	}

	return book, err
}

// `CheckoutBook` obtains book information with id `book_id`.
//
// Book information is stored in struct `Book`. When no book
// matches `book_id`, an `ErrBookNotFound` is returned.
func CheckoutBook(db *sql.DB, book_id int) (Book, error) {
	row := db.QueryRow(selectBook+"book_id=?", book_id)
	book, err := scanBook(row)
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
	row := db.QueryRow(selectBook+"isbn=?", isbn)
	book, err := scanBook(row)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

var selectRecord = `
SELECT
	record_id, user_id, book_id,
	return_date, borrow_date,
	deadline, final_deadline
FROM Record
WHERE `

func scanRecord(row RowScanner) (Record, error) {
	var return_date sql.NullTime
	var r Record
	err := row.Scan(
		&r.RecordID, &r.UserID, &r.BookID,
		&return_date, &r.BorrowDate,
		&r.DueDate, &r.FinalDate,
	)
	if err != nil {
		return Record{}, err
	}

	if return_date.Valid {
		r.Returned = true
		r.ReturnDate = return_date.Time
	}

	return r, nil
}

var ErrInvalidRecordID = errors.New("invalid record ID")

// CheckoutRecord retrieves record with `record_id`.
//
// Record information is stored in struct `Record`. When no record
// matches `record_id`, an `ErrInvalidRecordID` is returned.
func CheckoutRecord(db *sql.DB, record_id int) (Record, error) {
	row := db.QueryRow(selectRecord+"record_id = ?", record_id)
	r, err := scanRecord(row)
	if err == sql.ErrNoRows {
		return Record{}, ErrInvalidRecordID
	}
	if err != nil {
		return Record{}, err
	}

	return r, err
}

var ErrNoAvailableBook = errors.New("no available book")
var ErrInvalidBookID = errors.New("invalid book ID")
var ErrSuspendedUser = errors.New("user suspended: you have more than 3 overdue books")

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

var ErrAlreadyReturned = errors.New("this book has been returned")
var ErrOverdue = errors.New("cannot extend deadline for overdue records")
var ErrNotExtensible = errors.New("do extend deadline in the last week")
var ErrFinalDeadline = errors.New("must return book in three months")

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
	var tmp int
	err := db.QueryRow("SELECT book_id FROM Book WHERE book_id=?", book_id).
		Scan(&tmp)
	if err == sql.ErrNoRows {
		return ErrInvalidBookID
	}
	if err != nil {
		return err
	}

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

	_, err = db.Exec(buf.String(), args...)
	return err
}

// `SearchBookByTitle` returns all books whose title contain `keyword`.
func SearchBookByTitle(db *sql.DB, keyword string) ([]Book, error) {
	list := []Book{}
	rows, err := db.Query(selectBook+"title LIKE ?", fmt.Sprintf("%%%s%%", keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		book, err := scanBook(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, book)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}

// `SearchBookByAuthor` returns all book whose author names contain `keyword`.
func SearchBookByAuthor(db *sql.DB, keyword string) ([]Book, error) {
	list := []Book{}
	rows, err := db.Query(selectBook+"author LIKE ?", fmt.Sprintf("%%%s%%", keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		book, err := scanBook(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, book)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}

// `ChechoutHistory` list borrow history of user with `user_id`.
// The max number of records can be controlled by `limit` argument.
// `filter` is used in WHERE clause in SQL statement, and `args` is the
// placeholders for prepared statements.
func CheckoutHistory(db *sql.DB, user_id int, limit int, filter string, args ...interface{}) ([]Record, error) {
	if len(filter) != 0 {
		filter = filter + " AND user_id = ?"
	} else {
		filter = "user_id = ?"
	}
	query := selectRecord + filter + " ORDER BY borrow_date DESC, record_id DESC LIMIT ?"
	// fmt.Println(query)

	args = append(args, user_id)
	args = append(args, limit)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Record{}
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, r)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}
