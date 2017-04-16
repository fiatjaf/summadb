package database

import (
	"errors"
	"strconv"
	"strings"

	"github.com/fiatjaf/levelup"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/utils"
)

// the most simple function in the world, a helper to the _rev for a given path.
func (db *SummaDB) Rev(p types.Path) (string, error) {
	// check if the path is valid
	if !p.WriteValid() {
		return "", errors.New("cannot get rev for invalid path: " + p.Join())
	}

	return db.Get(p.Child("_rev").Join())
}

// --- helper functions not related to the above method:

const suffixLength = 5

func bumpRev(rev string) string {
	v, suffix := revNumber(rev)
	v++

	// put it on the minimum size (less 1)
	for len(suffix) < (suffixLength - 1) {
		suffix = utils.RandomString(1) + suffix
	}

	// open space to new char
	suffix = suffix[:suffixLength-1]

	return strconv.Itoa(v) + "-" + utils.RandomString(1) + suffix
}

func revNumber(rev string) (int, string) {
	spl := strings.Split(rev, "-")
	v, _ := strconv.Atoi(spl[0])
	if len(spl) == 2 {
		return v, spl[1]
	}
	return v, utils.RandomString(suffixLength)
}

// revFromParents() takes two revs and produce a third one.
// every time you provide the same two revs, independent of the order,
// revFromParents() should provide you with the same child rev.
// Intended to be used in replication.
// It expects r1 and r2 to have the same prefix number.
func revFromParents(r1 string, r2 string) string {
	n, suffix1 := revNumber(r1)
	_, suffix2 := revNumber(r2)

	var distance int
	for distance = suffixLength - 1; distance >= 0; distance-- {
		if suffix1[distance] != suffix2[distance] {
			break
		}
	}
	distance += 1

	// increases the prefix number according to the suffix distance,
	// because if two databases have diverged a lot, then later converged,
	// their revs should be more valuable than those of other still
	// diverged databases (but this shouldn't make a lot of difference).
	n += distance

	var suffix string
	for i := 0; i < distance; i++ {
		suffix += utils.LetterByIndex(int(suffix1[i] + suffix2[i]))
	}
	suffix += suffix1[distance:]

	return strconv.Itoa(n) + "-" + suffix
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
