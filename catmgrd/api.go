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

// AuthUser check `user_id` and `password` against table User
// in database `db`, which stores the sha1 hashes of passwords.
// Requested permissions `req` are encapsulated in Permission struct.
// Auth success if no error returned.
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
