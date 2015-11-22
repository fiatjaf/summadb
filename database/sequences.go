package database

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func ListChangesAt(path string) ([]struct {
	subpath string
	rev     string
	seq     uint64
}, error) {
	seqs := Open().Sub(BY_SEQ)
	defer seqs.Close()

	res := make([]struct {
		subpath string
		rev     string
		seq     uint64
	}, 0)

	basepath := path + "::"
	baselength := len(basepath)
	iter := seqs.NewIterator(util.BytesPrefix([]byte(basepath)), nil)
	for iter.Next() {
		seqstr := string(iter.Key())[baselength:]
		seq, _ := strconv.ParseUint(seqstr, 10, 64)
		valp := strings.Split(string(iter.Value()), "::")
		subpath := valp[0]
		rev := valp[1]

		res = append(res, struct {
			subpath string
			rev     string
			seq     uint64
		}{subpath, rev, seq})
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func LastSeqAt(path string) (uint64, error) {
	seqs := Open().Sub(BY_SEQ)
	defer seqs.Close()

	var seq uint64
	iter := seqs.NewIterator(util.BytesPrefix([]byte(path+"::")), nil)
	if iter.Last() {
		seq, _ = strconv.ParseUint(string(bytes.Split(iter.Key(), []byte{':', ':'})[1]), 10, 64)
	} else {
		seq = 0
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func GlobalUpdateSeq() uint64 {
	db := Open()
	defer db.Close()

	return getUpdateSeq(db)
}

func getUpdateSeq(db *sublevel.AbstractLevel) uint64 {
	seqs := db.Sub(BY_SEQ)
	v, err := seqs.Get([]byte(UPDATE_SEQ_KEY), nil)
	if err == nil {
		seq, _ := strconv.ParseUint(string(v), 10, 64)
		return seq
	}
	return 0
}
