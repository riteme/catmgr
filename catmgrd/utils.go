package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type MySQLConfig struct {
	Username string
	Password string
	Protocol string
	Address  string
	Port     int
	Database string
}

func LoadMySQLConfig(path string) (MySQLConfig, error) {
	fp, err := os.Open(path)
	if err != nil {
		return MySQLConfig{}, err
	}

	var config MySQLConfig
	err = json.NewDecoder(fp).Decode(&config)
	if err != nil {
		return MySQLConfig{}, err
	}

	return config, nil
}

func ConnectMySQL(config MySQLConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true",
		config.Username, config.Password,
		config.Protocol, config.Address,
		config.Port, config.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SendJSON(resp http.ResponseWriter, v interface{}) {
	resp.Header().Set("Content-Type", "application/json")

	var err error
	switch v.(type) {
	case error:
		err = json.NewEncoder(resp).Encode(MError{"failed", v.(error).Error()})
	default:
		err = json.NewEncoder(resp).Encode(v)
	}

	if err != nil {
		log.Print(err)
		http.Error(resp, "interal error", http.StatusInternalServerError)
	}
}

func DecodePayload(resp http.ResponseWriter, req *http.Request, v interface{}) bool {
	err := json.NewDecoder(req.Body).Decode(v)
	if err != nil {
		log.Println(req.RemoteAddr, "DecodePayload", err)
		SendJSON(resp, MErrDecodePayload)
		return false
	}
	return true
}

func AuthRequest(resp http.ResponseWriter, req *http.Request, user_id int, password string, perm Permission) bool {
	err := AuthUser(db, user_id, password, perm)
	if err == ErrInvalidUserID || err == ErrInvalidPassword || err == ErrPermissionDenied {
		log.Println(req.RemoteAddr, "AuthRequest", err)
		SendJSON(resp, MError{"failed", err.Error()})
		return false
	}
	if err != nil {
		log.Println(req.RemoteAddr, "AuthRequest", err)
		SendJSON(resp, MErrAuthUser)
		return false
	}
	return true
}

func CheckRecordID(resp http.ResponseWriter, req *http.Request, record_id, user_id int) bool {
	r, err := CheckoutRecord(db, record_id)
	if err == ErrInvalidRecordID {
		log.Println(req.RemoteAddr, "CheckRecordID", err)
		SendJSON(resp, err)
		return false
	}
	if err != nil {
		log.Println(req.RemoteAddr, "CheckRecordID", err)
		SendJSON(resp, NewMError("an error occurred during examining record"))
		return false
	}
	if r.UserID != user_id {
		SendJSON(resp, NewMError("your user ID does not match that of the record"))
		return false
	}
	return true
}
