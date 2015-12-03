package database

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

func Random(bytes int) string {
	random := make([]byte, bytes)
	_, err := rand.Read(random)
	if err != nil {
		log.Fatal("Couldn't read random bytes: ", err)
	}
	return fmt.Sprintf("%x", random)
}

func NewRev(oldrev string) string {
	n, _ := strconv.Atoi(strings.Split(oldrev, "-")[0])

	return fmt.Sprintf("%d-%s", (n + 1), Random(5))
}

func GetRev(path string) []byte {
	docs := db.Sub(DOC_STORE)
	oldrev, err := docs.Get([]byte(path+"/_rev"), nil)
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

func EscapeKey(e string) string {
	return url.QueryEscape(e)
}

func UnescapeKey(e string) string {
	v, _ := url.QueryUnescape(e)
	return v
}
