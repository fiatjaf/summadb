package database

import (
	"bytes"
	"strconv"

	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func LastSeqAt(path string) (uint64, error) {
	seqs := Open().Sub(BY_SEQ)
	defer seqs.Close()

	var seq uint64
	it := seqs.NewIterator(util.BytesPrefix([]byte(path+":")), nil)
	if it.Last() {
		seq, _ = strconv.ParseUint(string(bytes.Split(it.Key(), []byte{':'})[1]), 10, 64)
	} else {
		seq = 0
	}
	it.Release()
	err := it.Error()
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func UpdateSeq() uint64 {
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
