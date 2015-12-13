# SummaDB [![Travis-CI build status](https://travis-ci.org/fiatjaf/summadb.svg)](https://travis-ci.org/fiatjaf/summadb)

SummaDB is a [CouchDB](http://couchdb.org/)-inspired open source **hierarchical database** with infinitely nested sub-databases that **sync with PouchDB**. It is useful for backing offline-first applications with interesting calculations on server-side, users with hierarchical access to data working from a client-side app and live listeners.

## Features

* Simple HTTP API
* Syncs with [PouchDB](http://pouchdb.com/)
* Supports fine-grained document updates with PATCH
* Every document is also a database
* Changes feeds at each hierarchy level
* User accounts for REST access
* ACLs with read, write and admin access at each hierarchy level

## Features planned

* CouchDB-like [map functions](http://docs.couchdb.org/en/1.6.1/couchapp/ddocs.html#map-functions), implemented at each sub database level and run in the background
* Full-featured websocket server compatible with [socket-pouch](https://github.com/nolanlawson/socket-pouch)
* Changes feed published continuously through
  * long-polling
  * websockets
  * webhooks
* [JWT](http://jwt.io/) sessions
* Security
  * Users tied to each sub database, instead of only at root
  * User roles
  * Functions determining access at runtime
