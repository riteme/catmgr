package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	fp, err := os.Open("test_config.json")
	if err != nil {
		panic(err)
	}

	var config struct {
		Username string
		Password string
		Protocol string
		Address  string
		Port     int
		Database string
	}
	err = json.NewDecoder(fp).Decode(&config)
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true",
		config.Username, config.Password,
		config.Protocol, config.Address,
		config.Port, config.Database)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var return_code int
	defer func() {
		db.Close()
		os.Exit(return_code)
	}()

	return_code = m.Run()
}

func TestAuthUser(t *testing.T) {
	tb := []struct {
		userID   int
		password string
		req      Permission
		err      error
	}{
		{1, "root", Permission{true, true, true, true}, nil},
		{2, "admin", Permission{true, true, false, true}, nil},
		{233, "admin", Permission{true, true, false, true}, ErrInvalidUserID},
		{2, "admin", Permission{true, true, true, true}, ErrPermissionDenied},
		{2, "admin", Permission{false, false, false, true}, nil},
		{2, "admin", Permission{false, false, false, false}, nil},
		{3, "123456", Permission{false, false, true, false}, nil},
		{3, "1234567", Permission{false, false, true, false}, ErrInvalidPassword},
		{4, "123456", Permission{false, false, true, false}, nil},
		{5, "654321", Permission{false, false, true, false}, ErrPermissionDenied},
		{5, "654321", Permission{false, false, false, false}, nil},
		{5, "", Permission{false, false, false, false}, ErrInvalidPassword},
		{19260817, "", Permission{false, false, false, false}, ErrInvalidUserID},
		{0, "", Permission{false, false, false, false}, ErrInvalidUserID},
		{-1, "", Permission{false, false, false, false}, ErrInvalidUserID},
	}

	for _, e := range tb {
		err := AuthUser(db, e.userID, e.password, e.req)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		}
	}
}

func TestGetUserTypeID(t *testing.T) {
	tb := []struct {
		type_name string
		type_id   int
		err       error
	}{
		{"root", 1, nil},
		{"admin", 2, nil},
		{"student", 3, nil},
		{"guest", 4, nil},
		{"trump", -1, ErrInvalidUserType},
	}

	for _, e := range tb {
		got, err := GetUserTypeID(db, e.type_name)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		} else if got != e.type_id {
			t.Errorf("type_name = %#v; expected: %d, got: %d",
				e.type_name, e.type_id, got)
		}
	}
}

func randString(length int) string {
	var buf strings.Builder
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < length; i++ {
		buf.WriteByte(charset[rand.Intn(len(charset))])
	}
	return buf.String()
}

func TestAddUser(t *testing.T) {
	type_id := rand.Intn(4) + 1
	username := randString(8)
	password := randString(16)
	t.Logf("type_id = %d, username = %#v, password = %#v",
		type_id, username, password)

	user_id, err := AddUser(db, type_id, username, password)
	if err != nil {
		t.Error(err)
	} else {
		err := AuthUser(db, user_id, password, Permission{})
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCheckoutBook(t *testing.T) {
	tb := []struct {
		book_id int
		isbn    string
		title   string
		err     error
	}{
		{1, "978-981-13-2971-5", "Monte Carlo Methods", nil},
		{5, "978-3-030-33836-7", "Database Design and Implementation", nil},
		{13, "978-3-662-53622-3", "Graph Theory", nil},
		{26, "978-1-4939-2865-1", "Encyclopedia of Algorithms", nil},
		{-1, "", "", ErrBookNotFound},
	}

	for _, e := range tb {
		book, err := CheckoutBook(db, e.book_id)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		} else if err == nil {
			if book.Title != e.title {
				t.Errorf("expected: %#v, got: %#v", e.title, book.Title)
			} else if book.ISBN != e.isbn {
				t.Errorf("expected: %#v, got: %#v", e.isbn, book.ISBN)
			}
		}
	}
}

func TestCheckoutISBN(t *testing.T) {
	tb := []struct {
		isbn  string
		title string
		err   error
	}{
		{"978-981-13-2971-5", "Monte Carlo Methods", nil},
		{"978-3-030-33836-7", "Database Design and Implementation", nil},
		{"978-3-662-53622-3", "Graph Theory", nil},
		{"978-1-4939-2865-1", "Encyclopedia of Algorithms", nil},
		{"978-1-4939-2865-12", "Encyclopedia of Algorithms", ErrBookNotFound},
	}

	for _, e := range tb {
		book, err := CheckoutISBN(db, e.isbn)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		} else if err == nil {
			if book.Title != e.title {
				t.Errorf("expected: %#v, got: %#v", e.title, book.Title)
			}
		}
	}
}

func parseDate(val string) time.Time {
	layout := "2006-01-02"
	ret, _ := time.Parse(layout, val)
	return ret
}

func TestCheckoutRecord(t *testing.T) {
	e := Record{
		RecordID:   1,
		UserID:     6,
		BookID:     1,
		Returned:   false,
		ReturnDate: time.Time{},
		BorrowDate: parseDate("1926-08-17"),
		DueDate:    parseDate("1926-09-17"),
		FinalDate:  parseDate("2020-02-02"),
	}
	r, err := CheckoutRecord(db, 1)
	if err != nil {
		t.Error(err)
	} else if r != e {
		t.Errorf("expected: %+v, got: %+v", e, r)
	}

	_, err = CheckoutRecord(db, -1)
	if err != ErrInvalidRecordID {
		t.Error(err)
	}
}

func TestBorrowBook(t *testing.T) {
	tb := []struct {
		user_id int
		book_id int
		err     error
	}{
		{3, 5, nil},
		{3, 11, ErrNoAvailableBook},
		{3, -1, ErrInvalidBookID},
		{6, 10, nil},
		{7, 10, ErrSuspendedUser},
		{4, 10, nil},
	}

	for _, e := range tb {
		_, err := BorrowBook(db, e.user_id, e.book_id)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		}
	}
}

type fakeRecord struct {
	user_id int
	book_id int
	ret     *time.Time
	borrow  time.Time
	due     time.Time
	final   time.Time
	err     error
}

func insertFakeRecord(r fakeRecord) (int, error) {
	result, err := db.Exec(`
		INSERT INTO Record
			(user_id, book_id, borrow_date, deadline, final_deadline)
		VALUES (?, ?, ?, ?, ?)`,
		r.user_id, r.book_id, r.borrow, r.due, r.final)
	if err != nil {
		return -1, err
	}

	record_id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	if r.ret != nil {
		_, err = db.Exec(`
			UPDATE Record
			SET
				return_date = ?
			WHERE record_id = ?`,
			r.ret, record_id)
		if err != nil {
			return -1, err
		}
	}

	return int(record_id), nil
}

func TestExtendDeadline(t *testing.T) {
	err := ExtendDeadline(db, -1)
	if err != ErrInvalidRecordID {
		t.Fatalf("expected <invalid record id>, got: %+v", err)
	}

	today := time.Now()
	tb := []fakeRecord{
		{8, 5, nil, today, today.Add(day), today.Add(day + month), nil},
		{8, 5, &today, today, today, today, ErrAlreadyReturned},
		{8, 5, nil, today.Add(-2 * day), today.Add(-day), today.Add(3 * month), ErrOverdue},
		{8, 5, nil, today, today.Add(month), today.Add(month * 2), ErrNotExtensible},
		{8, 5, nil, today, today.Add(day), today.Add(week), ErrFinalDeadline},
	}

	for _, e := range tb {
		record_id, err := insertFakeRecord(e)
		if err != nil {
			t.Fatal(err)
		}

		err = ExtendDeadline(db, record_id)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		}
	}
}

func TestReturnBook(t *testing.T) {
	err := ReturnBook(db, -1)
	if err != ErrInvalidRecordID {
		t.Fatalf("expected <invalid record id>, got: %+v", err)
	}

	today := time.Now()
	tb := []fakeRecord{
		{8, 5, nil, today, today.Add(day), today.Add(day + month), nil},
		{8, 5, &today, today, today, today, ErrAlreadyReturned},
		{8, 5, nil, today.Add(-2 * day), today.Add(-day), today.Add(3 * month), nil},
		{8, 5, nil, today, today.Add(month), today.Add(month * 2), nil},
		{8, 5, nil, today, today.Add(day), today.Add(week), nil},
	}

	for _, e := range tb {
		record_id, err := insertFakeRecord(e)
		if err != nil {
			t.Fatal(err)
		}

		err = ReturnBook(db, record_id)
		if err != e.err {
			t.Errorf("expected: %+v, got: %+v", e.err, err)
		}
	}
}
