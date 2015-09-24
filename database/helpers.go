package database

import (
	"crypto/md5"
	"fmt"
	"io"
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

func NewRev(oldrev string, value []byte) string {
	n, _ := strconv.Atoi(strings.Split(oldrev, "-")[0])

	h := md5.New()
	io.WriteString(h, oldrev)
	h.Write(value)
	return fmt.Sprintf("%d-%x", (n + 1), h.Sum(nil))
}
