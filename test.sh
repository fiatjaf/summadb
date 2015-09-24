#!/bin/bash

trap 'killall' INT

killall() {
    trap '' INT TERM     # ignore INT and TERM while shutting down
    echo "**** Shutting down... ****"     # added double quotes
    kill -TERM 0         # fixed order, send TERM not INT
    wait
    echo DONE
    exit 0
}

export LEVELDB_PATH=test.db

. assert.sh;

rm -rf $LEVELDB_PATH;
go get && summadb &
sleep 2;

crud () {
  curl -X PUT http://localhost:5000/frutas/banana/cor -d 'amarela';
  assert "curl http://localhost:5000/frutas/banana/cor" '"amarela"';
  assert "curl http://localhost:5000/frutas/banana/cor/_rev | py -x 'x[0]'" "1";
  assert "curl 'http://localhost:5000/frutas/banana?children=true' | jq '.cor._val'" '"amarela"'
  curl -X PUT http://localhost:5000/frutas/banana/cor -d 'azul';
  assert "curl http://localhost:5000/frutas/banana/cor" '"azul"';
  assert "curl http://localhost:5000/frutas/banana/cor/_rev | py -x 'x[0]'" "2";
  assert "curl 'http://localhost:5000/frutas/banana?children=true' | jq '.cor._val'" '"azul"'

  assert_end
}
        echo '  * check initial db again';
crud;

# -- os finalmentes.
cat;
killall;
