package database

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func SetRulesAt(path string, security map[string]interface{}) error {
	metastore := db.Sub(PATH_METADATA)
	batch := metastore.NewBatch()

	if iread, ok := security["_read"]; ok {
		batch.Put([]byte(path+"/_read"), []byte(iread.(string)))
	}
	if iwrite, ok := security["_write"]; ok {
		batch.Put([]byte(path+"/_write"), []byte(iwrite.(string)))
	}
	if iadmin, ok := security["_admin"]; ok {
		batch.Put([]byte(path+"/_admin"), []byte(iadmin.(string)))
	}

	return metastore.Write(batch, nil)
}

func GetWriteRuleAt(path string) string { return getRuleAt(path, "write") }
func GetReadRuleAt(path string) string  { return getRuleAt(path, "read") }
func GetAdminRuleAt(path string) string { return getRuleAt(path, "admin") }

func getRuleAt(path string, kind string) string {
	metastore := db.Sub(PATH_METADATA)
	val, err := metastore.Get([]byte(path+"/_"+kind), nil)
	if err != nil {
		return ""
	}
	return string(val)
}

func WriteAllowedAt(path string, user string) bool { return operationAllowedAt(path, user, "write") }
func ReadAllowedAt(path string, user string) bool  { return operationAllowedAt(path, user, "read") }
func AdminAllowedAt(path string, user string) bool { return operationAllowedAt(path, user, "admin") }

func operationAllowedAt(path string, user string, kind string) bool {
	if user == "" {
		return false
	}

	keys := SplitKeys(path)
	for i := len(keys); i >= 1; i-- {
		var subpath string
		if len(keys[:i]) <= 1 {
			subpath = "/"
		} else {
			subpath = JoinKeys(keys[:i])
		}
		rule := getRuleAt(subpath, kind)

		for _, nameInRule := range strings.Split(string(rule), ",") {
			nameInRule = strings.TrimSpace(nameInRule)
			if nameInRule == user || nameInRule == "*" {
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

	if name == "" {
		return errors.New("empty user name.")
	}

	nameb := []byte(name)
	passwordb := []byte(password)

	// check if user exists
	if _, err := users.Get(nameb, nil); err == nil {
		return errors.New("user already exists")
	}

	// generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordb, 12)
	if err != nil {
		return err
	}

	// finally save.
	return users.Put(nameb, hashedPassword, nil)
}
