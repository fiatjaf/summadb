package database

import (
	"crypto/rand"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fiatjaf/sublevel"
)

func NewRev(oldrev string) string {
	n, _ := strconv.Atoi(strings.Split(oldrev, "-")[0])

	random := make([]byte, 12)
	_, err := rand.Read(random)
	if err != nil {
		log.Fatal("Couldn't read random bytes: ", err)
	}
	return fmt.Sprintf("%d-%x", (n + 1), random)
}

func GetRev(db *sublevel.Sublevel, path string) []byte {
	oldrev, err := db.Get([]byte(path+"/_rev"), nil)
	if err != nil {
		oldrev = []byte("0-00000")
	}
	return oldrev
}

func SplitKeys(path string) []string {
	return strings.Split(path, "/")
}

func JoinKeys(keys []string) string {
	return strings.Join(keys, "/")
}

/*
   Removes _rev, _changes, _deleted and other special things
   from the end of a path.
*/
func CleanPath(path string) string {
	return strings.Split(path, "/_")[0]
}

///*
//   Returns the base key (or the "database" to which the last key
//   is and id) and the last key (or the "id" of this path relative
//   to the aforementioned database).
//*/
//func BaseAndLastKey(path string) (base string, last string) {
//
//}
