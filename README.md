[![Travis-CI build status](https://travis-ci.org/fiatjaf/summadb.svg)](https://travis-ci.org/fiatjaf/summadb)

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
