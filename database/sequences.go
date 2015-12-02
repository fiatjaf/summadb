package database

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb/util"
)

type Change struct {
	Seq     uint64    `json:"seq"`
	Id      string    `json:"id"`
	Changes []justRev `json:"changes"`
}

type justRev struct {
	Rev string `json:"rev"`
}

func ListChangesAt(path string, since uint64) ([]Change, error) {
	seqs := db.Sub(BY_SEQ)

	res := make([]Change, 0)

	basepath := path + "::"
	baselength := len(basepath)
	iter := seqs.NewIterator(util.BytesPrefix([]byte(basepath)), nil)
	for iter.Next() {
		seqstr := string(iter.Key())[baselength:]
		seq, _ := strconv.ParseUint(seqstr, 10, 64)
		valp := strings.Split(string(iter.Value()), "::")
		subpath := valp[0]
		rev := valp[1]

		if seq <= since {
			continue
		}

		res = append(res, Change{
			Id:      subpath,
			Seq:     seq,
			Changes: []justRev{justRev{rev}},
		})
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func LastSeqAt(path string) (uint64, error) {
	seqs := db.Sub(BY_SEQ)

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
	return getUpdateSeq()
}

func getUpdateSeq() uint64 {
	seqs := db.Sub(BY_SEQ)
	v, err := seqs.Get([]byte(UPDATE_SEQ_KEY), nil)
	if err == nil {
		seq, _ := strconv.ParseUint(string(v), 10, 64)
		return seq
	}
	return 0
}

func RevsAt(path string) (revs []string, err error) {
	seqs := db.Sub(REV_STORE)

	basepath := path + "::"
	lenbase := len(basepath)
	already := make(map[string]bool)
	iter := seqs.NewIterator(util.BytesPrefix([]byte(basepath)), nil)
	i := 0
	for iter.Next() {
		rev := string(iter.Key())[lenbase:]
		revid := strings.Split(rev, "-")[0]
		if _, there := already[revid]; there {
			revs[i] = rev
		} else {
			revs = append(revs, rev)
			already[revid] = true
			i++
		}
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return revs, err
	}

	return revs, nil
}
