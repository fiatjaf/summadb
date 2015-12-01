if (typeof window == 'undefined') {
  expect = require('chai').expect
  PouchDB = require('pouchdb')
  fetch = require('node-fetch')
} else {
  expect = chai.expect
}

var local
var summa = "http://spooner.alhur.es:5000/subdb"

const value = v => Object({_val: v})

describe('integration', function () {
  this.timeout(4000)

  before(() => {
    return Promise.resolve().then(() => {
      return fetch(summa + '/docid').then(r => r.json())
    }).then(function (doc) {
      return fetch(summa + '/docid?rev=' + doc._rev, {method: 'DELETE'})
    }).catch(() => {
       // no need to remove
    }).then(() => {
      return new PouchDB("summadb-test")
    }).then(db => {
      local = db
      return local.destroy()
    }).then(() => {
      return new PouchDB("summadb-test")
    }).then((db) => {
      local = db
    })
  })

  describe('basic crud', () => {
    it('should add a doc', () => {
      return Promise.resolve().then(() => {
        return fetch(summa + '/docid', {method: 'PUT', body: JSON.stringify({what: 'a doc'})})
      })
    })
  })

  describe('replication', () => {
    it('should replicate from summa root to pouchdb', () => {
      return Promise.resolve().then(() => {
        return PouchDB.replicate(summa, local)
      }).then(() => {
        return local.get('docid')
      }).then((doc) => {
        expect(doc).to.have.all.keys(['_id', '_rev', 'what'])
        expect(doc.what).to.deep.equal(value('a doc'))
      })
    })

    it('should replicate from pouchdb to summa root', () => {
      var revs = []

      return Promise.resolve().then(()=> {
        return local.bulkDocs([
          {_id: 'this', sub: 'this is a document'},
          {_id: 'that', sub: 'that is a document'},
          {_id: 'array', array: [1,2,3,4,5]},
          {_id: 'complex', array: [
            ['a', {letter: 'a'}],
            {'subarray': [
              1, 2, ['xxx']
            ]
          }, true, 5]}
        ])
      }).then((res) => {
        revs = res.map(r => r.rev)
        return PouchDB.replicate(local, summa)
      }).then(() => {
        return Promise.all([
          fetch(summa + '/that').then(r => r.json()),
          fetch(summa + '/array').then(r => r.json()),
          fetch(summa + '/complex').then(r => r.json()),
        ]).then(that, array, complex => {
          expect(that._id).to.equal('that')
          expect(that._rev).to.equal(revs[1])
          expect(array._id).to.equal('array')
          expect(array._rev).to.equal(revs[2])
          expect(array.array).to.deep.equal({
            '0': value(1),
            '1': value(2),
            '2': value(3),
            '3': value(4),
            '4': value(5),
          })
          expect(complex.array).to.deep.equal({
            '0': {
              '0': value('a'),
              '1': {
                letter: value('a')
              }
            },
            '1': {
              subarray: {
                '0': value(1),
                '1': value(2),
                '2': {
                  '0': value('xxx')
                }
              }
            },
            '2': value(true),
            '3': value(5)
          })
        })
      })
    })
  })

})
