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

	type TestConfig struct {
		Username string
		Password string
		Protocol string
		Address  string
		Port     int
		Database string
	}

	var config TestConfig
	err = json.NewDecoder(fp).Decode(&config)
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
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
		userID   int64
		password string
		req      Permission
		ok       bool
	}{
		{1, "root", Permission{true, true, true, true}, true},
		{2, "admin", Permission{true, true, false, true}, true},
		{233, "admin", Permission{true, true, false, true}, false},
		{2, "admin", Permission{true, true, true, true}, false},
		{2, "admin", Permission{false, false, false, true}, true},
		{2, "admin", Permission{false, false, false, false}, true},
		{3, "123456", Permission{false, false, true, false}, true},
		{3, "1234567", Permission{false, false, true, false}, false},
		{4, "123456", Permission{false, false, true, false}, true},
		{5, "654321", Permission{false, false, true, false}, false},
		{5, "654321", Permission{false, false, false, false}, true},
		{5, "", Permission{false, false, false, false}, false},
		{19260817, "", Permission{false, false, false, false}, false},
		{0, "", Permission{false, false, false, false}, false},
		{-1, "", Permission{false, false, false, false}, false},
	}

	for _, entry := range tb {
		err := AuthUser(db, entry.userID, entry.password, entry.req)
		ok := err == nil
		if ok != entry.ok {
			t.Error(err)
			t.Errorf("fail: %+v", entry)
		}
	}
}

func TestGetUserTypeID(t *testing.T) {
	tb := []struct {
		type_name string
		type_id   int
	}{
		{"root", 1},
		{"admin", 2},
		{"student", 3},
		{"guest", 4},
	}

	for _, entry := range tb {
		got, err := GetUserTypeID(db, entry.type_name)
		if err != nil {
			t.Error(err)
		} else if got != entry.type_id {
			t.Errorf("type_name = %#v; expected: %d, got: %d",
				entry.type_name, entry.type_id, got)
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
		ok      bool
	}{
		{1, "978-981-13-2971-5", "Monte Carlo Methods", true},
		{5, "978-3-030-33836-7", "Database Design and Implementation", true},
		{13, "978-3-662-53622-3", "Graph Theory", true},
		{26, "978-1-4939-2865-1", "Encyclopedia of Algorithms", true},
		{-1, "978-1-4939-2865-1", "Encyclopedia of Algorithms", false},
	}

	for _, entry := range tb {
		book, err := CheckoutBook(db, entry.book_id)
		ok := err == nil
		if ok != entry.ok {
			t.Error(err)
		} else if ok {
			if book.Title != entry.title {
				t.Errorf("expected: %#v, got: %#v", entry.title, book.Title)
			} else if book.ISBN != entry.isbn {
				t.Errorf("expected: %#v, got: %#v", entry.isbn, book.ISBN)
			}
		}
	}
}

func TestSearchByISBN(t *testing.T) {
	tb := []struct {
		isbn  string
		title string
		ok    bool
	}{
		{"978-981-13-2971-5", "Monte Carlo Methods", true},
		{"978-3-030-33836-7", "Database Design and Implementation", true},
		{"978-3-662-53622-3", "Graph Theory", true},
		{"978-1-4939-2865-1", "Encyclopedia of Algorithms", true},
		{"978-1-4939-2865-12", "Encyclopedia of Algorithms", false},
	}

	for _, entry := range tb {
		book, err := SearchByISBN(db, entry.isbn)
		ok := err == nil
		if ok != entry.ok {
			t.Error(err)
		} else if ok {
			if book.Title != entry.title {
				t.Errorf("expected: %#v, got: %#v", entry.title, book.Title)
			}
		}
	}
}
