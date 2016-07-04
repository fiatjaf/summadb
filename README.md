# SummaDB [![Travis-CI build status](https://travis-ci.org/fiatjaf/summadb.svg)](https://travis-ci.org/fiatjaf/summadb)

SummaDB is a [CouchDB](http://couchdb.org/)-inspired open source **hierarchical database** with infinitely nested sub-databases, all those independently **syncable with PouchDB**. It offers a full-blown REST interface similar to CouchDB (with PUT calls that can be pointed to any leaf of the hierarchical tree and a PATCH method for multiple arbitrary leaf updates).

## Features

  - [x] Simple, CouchDB-like, HTTP API
  - [x] [PouchDB](http://pouchdb.com/) syncing
  - [x] Supports fine-grained document updates with PATCH
  - [x] Every document is also a database
  - [ ] GraphQL read queries
  - [ ] CouchDB-like [map functions](http://docs.couchdb.org/en/1.6.1/couchapp/ddocs.html#map-functions), implemented at each sub database level and run in the background
  - [ ] Changes feed at each hierarchy level
    - [x] Static
    - [ ] Published continuously through
      - [ ] long-polling
      - [ ] websockets
      - [ ] webhooks
  - [ ] Users
    - [x] User accounts for REST access
    - [x] Basic HTTP Auth
    - [x] ACLs with read, write and admin access at each hierarchy level
    - [x] Users tied to each sub database, instead of only at root
    - [ ] [JWT](http://jwt.io/) sessions
  - [ ] Websocket server compatible with [socket-pouch](https://github.com/nolanlawson/socket-pouch)
