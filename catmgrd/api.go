package main

import "fmt"
import "errors"
import "time"
import "database/sql"
import "crypto/sha1"

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

var ErrInvalidUserID = errors.New("Invalid user id")
var ErrInvalidPassword = errors.New("Invalid password")
var ErrPermissionDenied = errors.New("Permission denied")

// AuthUser check `login` information against table User
// in database `db`, which stores the sha1 hashes of passwords.
// Requested permissions `req` are encapsulated in Permission struct.
// Auth success if no error returned.
func AuthUser(db *sql.DB, user_id int64, password string, req Permission) error {
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

// GetUserTypeID returns `type_id` of `type_name` defined in
// UserType table.
func GetUserTypeID(db *sql.DB, type_name string) (int, error) {
	var type_id int
	query := `SELECT type_id FROM UserType WHERE type_name=?`
	err := db.QueryRow(query, type_name).Scan(&type_id)
	if err == sql.ErrNoRows {
		return 0, ErrInvalidUserType
	}
	if err != nil {
		return 0, err
	}

	return type_id, nil
}

// AddUser simply insert a new user record into User table.
func AddUser(db *sql.DB, type_id int, username string, password string) (int64, error) {
	token := fmt.Sprintf("%x", sha1.Sum([]byte(password)))
	result, err := db.Exec(
		"INSERT INTO User (type_id, name, token) VALUES (?, ?, ?)",
		type_id, username, token,
	)
	if err != nil {
		return 0, err
	}

	user_id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return user_id, nil
}

var selectBook = `
SELECT
	book_id, title, author, isbn,
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

// CheckoutBook obtains book information with id `book_id`.
func CheckoutBook(db *sql.DB, book_id int) (Book, error) {
	var book Book
	row := db.QueryRow(selectBook+"book_id=?", book_id)

	err := scanBook(row, &book)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

// CheckoutISBN obtains book information with `isbn`.
func CheckoutISBN(db *sql.DB, isbn string) (Book, error) {
	var book Book
	row := db.QueryRow(selectBook+"isbn=?", isbn)

	err := scanBook(row, &book)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

var month = time.Hour * 24 * 30

var ErrNoAvailableBook = errors.New("No available book")
var ErrInvalidBookID = errors.New("Invalid book id")

// BorrowBook attempts to borrow a book with `book_id` and leave a record.
func BorrowBook(db *sql.DB, user_id, book_id int) (int64, error) {
	now := time.Now()
	due := now.Add(month)
	final := now.Add(3 * month)

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var exist int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM Book WHERE book_id = ?", book_id).
		Scan(&exist)
	if err != nil {
		return 0, err
	}
	if exist == 0 {
		return 0, ErrInvalidBookID
	}

	// try decreasing available count
	result, err := tx.Exec(`
		UPDATE Book
		SET
			available_count = available_count - 1
		WHERE
			book_id = ? AND
			available_count > 0`, book_id)
	if err != nil {
		return 0, err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if cnt == 0 {
		return 0, ErrNoAvailableBook
	}

	result, err = tx.Exec(`
		INSERT INTO Record
			(user_id, book_id, borrow_date, deadline, final_deadline)
		VALUES (?, ?, ?, ?, ?)`,
		user_id, book_id, now, due, final)
	if err != nil {
		return 0, err
	}

	record_id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	tx.Commit()
	return record_id, nil
}
