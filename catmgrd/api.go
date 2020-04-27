package main

import "fmt"
import "errors"
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
		return errors.New(fmt.Sprintf("Invalid user id: %d", user_id))
	}
	if err != nil {
		return err
	}

	if hash != token {
		return errors.New("Incorrect password")
	}

	if req.mask()&perm.mask() != req.mask() {
		return errors.New("Permission denied")
	}

	return nil
}

// GetUserTypeID returns `type_id` of `type_name` defined in
// UserType table.
func GetUserTypeID(db *sql.DB, type_name string) (int, error) {
	var type_id int
	query := `SELECT type_id FROM UserType WHERE type_name=?`
	err := db.QueryRow(query, type_name).Scan(&type_id)
	if err == sql.ErrNoRows {
		return 0, errors.New(fmt.Sprintf("No user type named \"%s\"", type_name))
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

func scanBook(row *sql.Row, book *Book) error {
	err := row.Scan(
		&book.BookID, &book.Title,
		&book.Author, &book.ISBN,
		&book.AvailableCount,
		&book.Description,
		&book.Comment,
	)
	if err == sql.ErrNoRows {
		return errors.New("Book not found")
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

// SearchByISBN obtains book information with `isbn`.
func SearchByISBN(db *sql.DB, isbn string) (Book, error) {
	var book Book
	row := db.QueryRow(selectBook+"isbn=?", isbn)

	err := scanBook(row, &book)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}
