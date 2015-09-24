package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"os"
)

func Sub(store string) sublevel {
	dbfile := os.Getenv("LEVELDB_PATH")
	if dbfile == "" {
		dbfile = "example.db"
	}

	db, err := leveldb.OpenFile(dbfile, nil)
	if err != nil {
		log.Print("couldn't open database file. ", err)
	}
	return sublevel{
		namespace: []byte(store),
		db:        db,
	}
}

type sublevel struct {
	namespace []byte
	db        *leveldb.DB
}

func (s sublevel) Close() error {
	return s.db.Close()
}

func (s sublevel) Delete(key []byte, wo *opt.WriteOptions) error {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Delete(key, wo)
}

func (s sublevel) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Get(key, ro)
}

func (s sublevel) Put(key []byte, value []byte, wo *opt.WriteOptions) error {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Put(key, value, wo)
}

func (s sublevel) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	key = append(append([]byte(nil), s.namespace...), key...)
	return s.db.Has(key, ro)
}

/* iterator */
func (s sublevel) NewIterator(slice *util.Range, ro *opt.ReadOptions) subIterator {
	slice = &util.Range{
		Start: append(append([]byte(nil), s.namespace...), slice.Start...),
		Limit: append(append([]byte(nil), s.namespace...), slice.Limit...),
	}

	return subIterator{
		namespace: s.namespace,
		iterator:  s.db.NewIterator(slice, ro),
	}
}

type subIterator struct {
	namespace []byte
	iterator  iterator.Iterator
}

func (si subIterator) Key() []byte {
	key := si.iterator.Key()
	return key[len(si.namespace):]
}
func (si subIterator) Value() []byte {
	return si.iterator.Value()
}
func (si subIterator) Next() bool {
	return si.iterator.Next()
}
func (si subIterator) Prev() bool {
	return si.iterator.Prev()
}
func (si subIterator) Last() bool {
	return si.iterator.Last()
}
func (si subIterator) First() bool {
	return si.iterator.First()
}
func (si subIterator) Seek(key []byte) bool {
	key = append(append([]byte(nil), si.namespace...), key...)
	return si.iterator.Seek(key)
}
func (si subIterator) Release() {
	si.iterator.Release()
}
func (si subIterator) Error() error {
	return si.iterator.Error()
}

/* transactions */
func (s sublevel) NewBatch() *Batch {
	return &Batch{
		namespace: s.namespace,
		batch:     new(leveldb.Batch),
	}
}

type Batch struct {
	namespace []byte
	batch     *leveldb.Batch
}

func (b *Batch) Delete(key []byte) {
	key = append(append([]byte(nil), b.namespace...), key...)
	b.batch.Delete(key)
}

func (b *Batch) Put(key []byte, value []byte) {
	key = append(append([]byte(nil), b.namespace...), key...)
	b.batch.Put(key, value)
}

func (b *Batch) Len() int {
	return b.batch.Len()
}

func (b *Batch) Reset() {
	b.batch.Reset()
}

func (s sublevel) Write(b *Batch, wo *opt.WriteOptions) (err error) {
	return s.db.Write(b.batch, wo)
}
