package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type access string

const (
	READ_ACCESS  access = "READ"
	WRITE_ACCESS        = "WRITE"
	ADMIN_ACCESS        = "ADMIN"
)

func SetRuleAt(path string, kind access, val string) {

}

func GetRuleAt(path string, kind access) {

}

func ReadAllowedAt(path string, user string, kind access) bool {

}

func WriteAllowedAt(path string, user string, kind access) bool {

}

/* returns true if the given name/password combination is
   valid for an existing user, false otherwise. */
func ValidUser(name string, password string) bool {
	users := db.Sub(USER_STORE)

	if hashedPassword, err := users.Get([]byte(name), nil); err == nil {
		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		if err == nil {
			return true
		}
	}
	return false
}

/* saves a new combination of name/password in the database, if
   it doesn't exists. returns nil on success, an error otherwise. */
func SaveUser(name string, password string) error {
	users := db.Sub(USER_STORE)

	nameb := []byte(name)
	passwordb := []byte(password)

	// check if user exists
	if _, err := users.Get(nameb, nil); err != nil {
		return errors.New("user already exists")
	}

	// generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordb, 23)
	if err != nil {
		return err
	}

	// finally save.
	return users.Put(nameb, hashedPassword, nil)
}
