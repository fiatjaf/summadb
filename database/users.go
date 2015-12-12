package database

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func SetWriteRuleAt(path string, val string) error {
	metastore := db.Sub(PATH_METADATA)
	return metastore.Put([]byte(path+"/_write"), []byte(val), nil)
}

func SetReadRuleAt(path string, val string) error {
	metastore := db.Sub(PATH_METADATA)
	return metastore.Put([]byte(path+"/_write"), []byte(val), nil)
}

func GetWriteRuleAt(path string) string {
	metastore := db.Sub(PATH_METADATA)
	val, err := metastore.Get([]byte(path+"/_write"), nil)
	if err != nil {
		return ""
	}
	return string(val)
}

func GetReadRuleAt(path string) string {
	metastore := db.Sub(PATH_METADATA)
	val, err := metastore.Get([]byte(path+"/_write"), nil)
	if err != nil {
		return ""
	}
	return string(val)
}

func ReadAllowedAt(path string, user string) bool {
	keys := SplitKeys(path)
	for i := len(keys) - 1; i >= 1; i-- {
		subpath := JoinKeys(keys[:i])
		rule := GetReadRuleAt(subpath)

		for _, name := range strings.Split(string(rule), ",") {
			if name == user || name == "*" {
				return true
			}
		}
	}
	return false
}

func WriteAllowedAt(path string, user string) bool {
	keys := SplitKeys(path)
	for i := len(keys) - 1; i >= 1; i-- {
		subpath := JoinKeys(keys[:i])
		rule := GetWriteRuleAt(subpath)

		for _, name := range strings.Split(string(rule), ",") {
			if name == user || name == "*" {
				return true
			}
		}
	}
	return false
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
