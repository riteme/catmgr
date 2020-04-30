package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	addr := flag.String("listen", ":10777", "address for catmgrd server listening to")
	flag.Parse()

	config, err := LoadMySQLConfig("catmgrd.json")
	if err != nil {
		panic(err)
	}

	db, err = ConnectMySQL(config)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/new", handleNew)
	mux.HandleFunc("/update", handleUpdate)
	mux.HandleFunc("/adduser", handleAddUser)
	mux.HandleFunc("/show", handleShow)
	mux.HandleFunc("/list", handleList)
	mux.HandleFunc("/borrow", handleBorrow)
	mux.HandleFunc("/extend", handleExtend)
	mux.HandleFunc("/return", handleReturn)

	log.Println("start catmgrd")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

var MErrDecodePayload = NewMError("failed to decode payload")
var MErrAuthUser = NewMError("error occurred during authentication")

func handleRoot(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/\"")
	SendJSON(resp, MHello{"ok", "Hello, world!"})
}

func handleNew(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/new\"")

	var params struct {
		UserID   int    `json:"user_id"`
		Password string `json:"password"`
	}
	if !DecodePayload(resp, req, &params) ||
		!AuthRequest(resp, req, params.UserID, params.Password, Permission{Update: true}) {
		return
	}

	book_id, err := NewBook(db)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("error occurred during adding a book"))
	} else {
		log.Printf(req.RemoteAddr, "new book: %d", book_id)
		SendJSON(resp, MNewBook{"ok", book_id})
	}
}

func handleUpdate(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/update\"")

	var params struct {
		UserID      int     `json:"user_id"`
		Password    string  `json:"password"`
		BookID      int     `json:"book_id"`
		Diff        *int    `json:"diff"`
		Title       *string `json:"title"`
		Author      *string `json:"author"`
		ISBN        *string `json:"isbn"`
		Description *string `json:"description"`
		Comment     *string `json:"comment"`
	}
	if !DecodePayload(resp, req, &params) ||
		!AuthRequest(resp, req, params.UserID, params.Password, Permission{Update: true}) {
		return
	}

	diff := 0
	if params.Diff != nil {
		diff = *params.Diff
	}
	info := BookInfo{
		Title:       params.Title,
		Author:      params.Author,
		ISBN:        params.ISBN,
		Description: params.Description,
		Comment:     params.Comment,
	}

	err := UpdateBook(db, params.BookID, diff, info)
	if err == ErrInvalidBookID {
		log.Println(err)
		SendJSON(resp, err)
	} else if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("failed to update book information"))
	} else {
		log.Printf("update book: %d", params.BookID)
		SendJSON(resp, MBook{"ok", params.BookID})
	}
}

func handleAddUser(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/adduser\"")

	var params struct {
		UserID      int     `json:"user_id"`
		Password    string  `json:"password"`
		NewUserType *string `json:"new_user_type"`
		NewUsername *string `json:"new_username"`
		NewPassword *string `json:"new_password"`
	}
	if !DecodePayload(resp, req, &params) {
		return
	}
	if params.NewUserType == nil {
		SendJSON(resp, NewMError("missing field: new_user_type"))
		return
	}
	if params.NewUsername == nil {
		SendJSON(resp, NewMError("missing field: new_username"))
		return
	}
	if params.NewPassword == nil {
		SendJSON(resp, NewMError("missing field: new_password"))
		return
	}
	if !AuthRequest(resp, req, params.UserID, params.Password, Permission{AddUser: true}) {
		return
	}

	type_id, err := GetUserTypeID(db, *params.NewUserType)
	if err == ErrInvalidUserType {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("invalid user type"))
		return
	}
	if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("error occurred during examining user type"))
		return
	}

	user_id, err := AddUser(db, type_id, *params.NewUsername, *params.NewPassword)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("error occurred during adding user"))
	} else {
		log.Printf("user added: %d", user_id)
		SendJSON(resp, MAddUser{"ok", user_id})
	}
}

func handleShow(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/show\"")

	var params struct {
		Section string `json:"section"`
		Keyword string `json:"keyword"`
	}
	if !DecodePayload(resp, req, &params) {
		return
	}

	books := make([]Book, 1)
	var err error
	switch params.Section {
	case "book_id":
		book_id, parse_err := strconv.ParseInt(params.Keyword, 10, 32)
		if parse_err != nil {
			log.Println(req.RemoteAddr, err)
			SendJSON(resp, NewMError("invalid book ID"))
			return
		}
		books[0], err = CheckoutBook(db, int(book_id))
	case "isbn":
		books[0], err = CheckoutISBN(db, params.Keyword)
	case "title":
		books, err = SearchBookByTitle(db, params.Keyword)
	case "author":
		books, err = SearchBookByAuthor(db, params.Keyword)
	default:
		SendJSON(resp, NewMError(fmt.Sprintf("unknown section name: %#v", params.Section)))
		return
	}

	if err == ErrBookNotFound {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, err)
	} else if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("an error occurred during retrieving book information"))
	} else {
		SendJSON(resp, MBookList{"ok", books})
	}
}

func handleList(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/list\"")

	var params struct {
		UserID   int    `json:"user_id"`
		Password string `json:"password"`
		TargetID int    `json:"target_id"`
		Filter   string `json:"filter"`
		Limit    *int   `json:"limit"`
	}
	if !DecodePayload(resp, req, &params) {
		return
	}

	limit := 100
	if params.Limit != nil {
		limit = *params.Limit
	}

	perms := Permission{Inspect: params.UserID != params.TargetID}
	if !AuthRequest(resp, req, params.UserID, params.Password, perms) {
		return
	}

	today := time.Now()
	filter := ""
	args := []interface{}{}
	switch params.Filter {
	case "all":
		filter = ""
	case "not-returned":
		filter = "return_date IS NULL"
	case "overdue":
		filter = "return_date IS NOT NULL AND deadline < ?"
		args = []interface{}{today}
	default:
		SendJSON(resp, NewMError("invalid filter type"))
		return
	}

	list, err := CheckoutHistory(db, params.TargetID, limit, filter, args...)
	if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("an error occurred during retrieving borrow history"))
	} else {
		SendJSON(resp, MRecordList{"ok", list})
	}
}

func handleBorrow(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/borrow\"")

	var params struct {
		UserID   int    `json:"user_id"`
		Password string `json:"password"`
		BookID   int    `json:"book_id"`
	}
	if !DecodePayload(resp, req, &params) ||
		!AuthRequest(resp, req, params.UserID, params.Password, Permission{Borrow: true}) {
		return
	}

	record_id, err := BorrowBook(db, params.UserID, params.BookID)
	if err == ErrInvalidBookID || err == ErrSuspendedUser ||
		err == ErrNoAvailableBook {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, err)
	} else if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("an error occurred during borrowing book"))
	} else {
		log.Printf("new record: %d", record_id)
		SendJSON(resp, MRecord{"ok", record_id})
	}
}

func handleExtend(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/extend\"")

	var params struct {
		UserID   int    `json:"user_id"`
		Password string `json:"password"`
		RecordID int    `json:"record_id"`
	}
	if !DecodePayload(resp, req, &params) ||
		!AuthRequest(resp, req, params.UserID, params.Password, Permission{}) ||
		!CheckRecordID(resp, req, params.RecordID, params.UserID) {
		return
	}

	err := ExtendDeadline(db, params.RecordID)
	if err == ErrAlreadyReturned || err == ErrOverdue ||
		err == ErrNotExtensible || err == ErrFinalDeadline {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, err)
	} else if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("an error occurred during extending deadline"))
	} else {
		log.Printf("extend deadline: %d", params.RecordID)
		SendJSON(resp, MRecord{"ok", params.RecordID})
	}
}

func handleReturn(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, "access \"/return\"")

	var params struct {
		UserID   int    `json:"user_id"`
		Password string `json:"password"`
		RecordID int    `json:"record_id"`
	}
	if !DecodePayload(resp, req, &params) ||
		!AuthRequest(resp, req, params.UserID, params.Password, Permission{}) ||
		!CheckRecordID(resp, req, params.RecordID, params.UserID) {
		return
	}

	err := ReturnBook(db, params.RecordID)
	if err == ErrInvalidRecordID || err == ErrAlreadyReturned {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, err)
	} else if err != nil {
		log.Println(req.RemoteAddr, err)
		SendJSON(resp, NewMError("an error occurred during returning book"))
	} else {
		log.Printf("return book: %d", params.RecordID)
		SendJSON(resp, MRecord{"ok", params.RecordID})
	}
}
