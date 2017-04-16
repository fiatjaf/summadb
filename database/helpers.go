package database

import (
	"errors"
	"strconv"
	"strings"

	"github.com/fiatjaf/levelup"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/utils"
)

func bumpRev(rev string) string {
	v := RevNumber(rev)
	v++
	return strconv.Itoa(v) + "-" + utils.RandomString(4)
}

func RevNumber(rev string) int {
	spl := strings.Split(rev, "-")
	v, _ := strconv.Atoi(spl[0])
	return v
}

func (db *SummaDB) checkRev(providedrev string, p types.Path) error {
	currentrev, err := db.Get(p.Child("_rev").Join())
	if err == levelup.NotFound && providedrev == "" {
		return nil
	}
	if err != nil {
		log.Error("failed to fetch rev for checking.",
			"path", p,
			"provided", providedrev)
		return err
	}
	if currentrev == providedrev {
		return nil
	}
	return errors.New(
		"mismatched revs at " + p.Join() + ". current: " + currentrev + "; provided: " + providedrev)
}

type mapfupdated struct {
	path types.Path
	mapf string
}
