# SummaDB [![Travis-CI build status](https://travis-ci.org/fiatjaf/summadb.svg)](https://travis-ci.org/fiatjaf/summadb)

SummaDB is a [CouchDB](http://couchdb.org/)-inspired open source **hierarchical database** with infinitely nested sub-databases, all those independently **syncable with PouchDB**. It offers a full-blown REST interface similar to CouchDB (with PUT calls that can be pointed to any leaf of the hierarchical tree and a PATCH method for multiple arbitrary leaf updates) and a [GraphQL](http://facebook.github.io/graphql) read-only query interface

## Features

  - [x] Simple, CouchDB-like, HTTP API
  - [x] [PouchDB](http://pouchdb.com/) syncing
  - [x] Supports fine-grained document updates with PATCH
  - [x] Every document is also a database
  - [x] GraphQL read queries
  - [ ] CouchDB-like [map functions](http://docs.couchdb.org/en/1.6.1/couchapp/ddocs.html#map-functions), implemented at each sub database level and run in the background
  - [ ] Changes feed at each hierarchy level
    - [x] Static
    - [ ] Published continuously
      - [ ] long-polling
      - [ ] websockets
      - [ ] webhooks (_maybe_)
  - [ ] `/_format/fn` endpoint (_maybe_)
  - [ ] Users
    - [x] User accounts for REST access
    - [x] Basic HTTP Auth
    - [x] ACLs with read, write and admin access at each hierarchy level
    - [x] Users tied to each sub database, instead of only at root
    - [ ] [JWT](http://jwt.io/) sessions
  - [ ] Websocket server compatible with [socket-pouch](https://github.com/nolanlawson/socket-pouch) (_maybe_)

## Demo

You can see SummaDB in action right now, without installing anything. There is

  * a [live database](https://summadb-temp.herokuapp.com/) with prefilled data on Heroku
    * it is on "admin party" mode. you can modify the data as you want, or even create users and restrict access (Heroku, I'm counting on you to reset the changes made to the filesystem from time to time).
    * try visiting:
      * http://summadb-temp.herokuapp.com/players/
      * http://summadb-temp.herokuapp.com/players/kalimbaharkandant
      * http://summadb-temp.herokuapp.com/players/kalimbaharkandant/name
      * http://summadb-temp.herokuapp.com/players/kalimbaharkandant/name/_val
    * or perform some operations with curl:
      * `curl -X PUT 'https://summadb-temp.herokuapp.com/players/Aleianda' -d '{"name": "Aleianda", "class": "warrior"}'` to add a new player.
      * `curl -X PATCH 'https://summadb-temp.herokuapp.com/players?rev=2-0889863c34' -d '{"kalualualeiamba": {"class": "mage"}, "kalimbaharkandant": {"inventory": {"red potion": 8}}}` to change the class of a player and the number of potions of other.
      * `curl -X POST https://summadb-temp.herokuapp.com/_graphql -H 'content-type: application/json' -d '{"query": "query { players { kalimbaharkandant { inventory { cane } } }, items { cane { desc, cost, damage } } }"}'` to query specific values with GraphQL.
  * a [simple Javascript](https://summadb.github.io/admin/?summa=https://summadb-temp.herokuapp.com) app that allows browsing the hierarchical tree of data in that same live database
  * a small [demo of the PouchDB syncing capabilities](https://summadb.github.io/demo/)
    * it syncs two different sub-databases from the SummaDB to two different PouchDBs
    * the final result will be shown on a [PouchDB version of Fauxton](https://summadb.github.io/demo/fauxton/)

## Installing

```
go get github.com/fiatjaf/summadb
```

Then run it with `summadb` (`summadb --help` will show you some options).

There are also binaries that probably work at https://gobuilder.me/github.com/fiatjaf/summadb.


### Notice

This is pre-alpha, or whatever you want to call something that is in development by a single man and looking for feedback. Please give feedback, or use it and find bugs, or fix bugs. Or, better, if you have some experience with Go or databases, look at the code, find my stupid rookie mistakes and help me fix them. Thank you very much.
