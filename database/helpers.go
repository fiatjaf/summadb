package database

import (
	"crypto/rand"
	"fmt"
	"github.com/fiatjaf/sublevel"
	"log"
	"strconv"
	"strings"
)

func RevIsOk(rev string) bool {
	split := strings.Split(rev, "-")
	if len(split) != 2 || len(split[1]) <= 0 {
		return false
	}
	_, err := strconv.Atoi(split[0])
	if err != nil {
		return false
	}
	return true
}

func NewRev(oldrev string) string {
	n, _ := strconv.Atoi(strings.Split(oldrev, "-")[0])

	random := make([]byte, 12)
	_, err := rand.Read(random)
	if err != nil {
		log.Fatal("Couldn't read random bytes: ", err)
	}
	return fmt.Sprintf("%d-%x", (n + 1), random)
}

func bumpRevsInBatch(db *sublevel.Sublevel, batch *sublevel.SubBatch, modifiedKey string) {
	keyParts := strings.Split(modifiedKey, "/")
	for i := range keyParts {
		parentKey := strings.Join(keyParts[:i], "/")
		oldrev, err := db.Get([]byte(parentKey), nil)
		if err != nil {
			oldrev = []byte("0-00000")
		}

		batch.Put([]byte(parentKey+"/_rev"), []byte(NewRev(string(oldrev))))
	}
}
