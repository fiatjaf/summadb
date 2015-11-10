// Copyright (c) 2015, Giovanni T. Parra <fiatjaf@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package sublevel

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/* basic db management */

/*  opens an abstract DB for the given file.
    the abstract DB is said to be "abstract" because it cannot
    do any operation, it is there only to generate sublevels.

    It is not possible to access the underlying leveldb with this
    package.
*/
func OpenFile(dbfile string, o *opt.Options) AbstractLevel {
	db, err := leveldb.OpenFile(dbfile, o)
	return AbstractLevel{
		leveldb: db,
		err:     err,
	}
}

type AbstractLevel struct {
	leveldb *leveldb.DB
	err     error
}

func (a AbstractLevel) Close() error {
	return a.leveldb.Close()
}

func (a AbstractLevel) Sub(store string) (*Sublevel, error) {
	if a.err != nil {
		return &Sublevel{}, a.err
	}
	return &Sublevel{
		namespace: []byte("!" + store + "!"),
		db:        a.leveldb,
	}, nil
}

func (a AbstractLevel) MustSub(store string) *Sublevel {
	sub, err := a.Sub(store)
	if err != nil {
		log.Fatal("couldn't open database file. ", err)
	}
	return sub
}

type Sublevel struct {
	namespace []byte
	db        *leveldb.DB
}

func (s Sublevel) Close() error {
	return s.db.Close()
}

/* methods */

func (s Sublevel) Delete(key []byte, wo *opt.WriteOptions) error {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Delete(key, wo)
}

func (s Sublevel) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Get(key, ro)
}

func (s Sublevel) Put(key []byte, value []byte, wo *opt.WriteOptions) error {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Put(key, value, wo)
}

func (s Sublevel) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Has(key, ro)
}

/* iterator */
func (s Sublevel) NewIterator(slice *util.Range, ro *opt.ReadOptions) SubIterator {
	slice = &util.Range{
		Start: append(append([]byte(nil), s.namespace...), slice.Start...),
		Limit: append(append([]byte(nil), s.namespace...), slice.Limit...),
	}

	return SubIterator{
		namespace: s.namespace,
		iterator:  s.db.NewIterator(slice, ro),
	}
}

type SubIterator struct {
	namespace []byte
	iterator  iterator.Iterator
}

func (si SubIterator) Key() []byte {
	key := si.iterator.Key()
	return key[len(si.namespace):]
}
func (si SubIterator) Value() []byte {
	return si.iterator.Value()
}
func (si SubIterator) Next() bool {
	return si.iterator.Next()
}
func (si SubIterator) Prev() bool {
	return si.iterator.Prev()
}
func (si SubIterator) Last() bool {
	return si.iterator.Last()
}
func (si SubIterator) First() bool {
	return si.iterator.First()
}
func (si SubIterator) Seek(key []byte) bool {
	key = append(append([]byte(nil), si.namespace...), key...)
	return si.iterator.Seek(key)
}
func (si SubIterator) Release() {
	si.iterator.Release()
}
func (si SubIterator) Error() error {
	return si.iterator.Error()
}

/* transactions */
type AbstractBatchOperation struct {
	kind  string // PUT or DELETE
	key   []byte
	value []byte
}

//  starts a new batch write for the specified sublevel.
func (s Sublevel) NewBatch() *SubBatch {
	return &SubBatch{namespace: s.namespace}
}

type SubBatch struct {
	namespace []byte
	ops       []AbstractBatchOperation
}

func (b *SubBatch) Delete(key []byte) {
	key = append(append([]byte(nil), b.namespace...), key...)
	b.ops = append(b.ops, AbstractBatchOperation{"DELETE", key, nil})
}

func (b *SubBatch) Put(key []byte, value []byte) {
	key = append(append([]byte(nil), b.namespace...), key...)
	b.ops = append(b.ops, AbstractBatchOperation{"PUT", key, value})
}

func (b *SubBatch) Dump() []byte {
	return makeBatchWithOps(b.ops).Dump()
}

func (b *SubBatch) Len() int {
	return len(b.ops)
}

func (b *SubBatch) Reset() {
	b.ops = make([]AbstractBatchOperation, 0)
}

func (s Sublevel) Write(b *SubBatch, wo *opt.WriteOptions) error {
	return s.db.Write(makeBatchWithOps(b.ops), wo)
}

/* transactions on different stores */
func (a AbstractLevel) NewBatch() *SuperBatch {
	return &SuperBatch{}
}

type SuperBatch struct {
	ops []AbstractBatchOperation
}

func (sb *SuperBatch) MergeSubBatch(b *SubBatch) {
	sb.ops = append(sb.ops, b.ops...)
}

func (sb *SuperBatch) Dump() []byte {
	return makeBatchWithOps(sb.ops).Dump()
}

func (sb *SuperBatch) Len() int {
	return len(sb.ops)
}

func (sb *SuperBatch) Reset() {
	sb.ops = make([]AbstractBatchOperation, 0)
}

func (a AbstractLevel) Write(sb *SuperBatch, wo *opt.WriteOptions) error {
	return a.leveldb.Write(makeBatchWithOps(sb.ops), wo)
}

func (a AbstractLevel) MultiBatch(subbatches ...*SubBatch) *SuperBatch {
	sb := a.NewBatch()
	for _, b := range subbatches {
		sb.MergeSubBatch(b)
	}
	return sb
}

func makeBatchWithOps(ops []AbstractBatchOperation) *leveldb.Batch {
	batch := new(leveldb.Batch)
	for _, op := range ops {
		if op.kind == "PUT" {
			batch.Put(op.key, op.value)
		} else if op.kind == "DELETE" {
			batch.Delete(op.key)
		}
	}
	return batch
}
