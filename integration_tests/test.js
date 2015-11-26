var expect = chai.expect

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
  })

})
