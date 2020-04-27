package main

import "os"
import "fmt"
import "testing"
import "encoding/json"
import "database/sql"

import _ "github.com/go-sql-driver/mysql"

var db *sql.DB

func TestMain(m *testing.M) {
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
		userId   int
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
		err := AuthUser(db, entry.userId, entry.password, entry.req)
		ok := err == nil
		if ok != entry.ok {
			t.Error(err)
			t.Errorf("fail: %+v", entry)
		}
	}
}
