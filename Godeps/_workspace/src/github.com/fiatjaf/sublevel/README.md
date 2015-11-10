# sublevel

Separate sections of the same LevelDB. Compatible (at least in the basics -- I didn't test it minuciously) with [the nodejs sublevel](https://github.com/dominictarr/level-sublevel).

[![Travis-CI build status](https://travis-ci.org/fiatjaf/sublevel.svg)](https://travis-ci.org/fiatjaf/sublevel)

```
import "github.com/fiatjaf/sublevel"

sub, err := sublevel.OpenFile("example.db").Sub("specific-stuff")
if err != nil {
  panic(err)
}

sub.Put([]byte("this"), []byte("2007-04-01"), nil)
dateOfThis := sub.Get([]byte("this"), nil)
```

**sublevel** is built on top of [goleveldb](http://godoc.org/github.com/syndtr/goleveldb/leveldb) and supports most methods from there (not all, but in most cases everything you'll need).
