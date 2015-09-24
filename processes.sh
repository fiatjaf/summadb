#!/bin/bash

trap 'killall' INT

killall() {
    trap '' INT TERM     # ignore INT and TERM while shutting down
    echo "**** Shutting down... ****"     # added double quotes
    kill -TERM 0         # fixed order, send TERM not INT
    wait
    echo DONE
    $BASE/test/drop_all.sh
    exit 0
}

. assert.sh
. base.sh

# purge queue
echo '> purging queue'
coffee $BASE/receive-webhooks/purge-queue.coffee;

# create database schema
echo '> creating database schema'
$BASE/model-updates/venv/bin/python $BASE/model-updates/manage.py db upgrade -d $BASE/model-updates/migrations;

# start all processes
echo '> starting all'
coffee $BASE/api/index.coffee &
coffee $BASE/receive-webhooks/index.coffee &
cd $BASE/website;
ls | entr bash -c 'sake js && sake html && sake css' &
instant -q --delay 15 4000 &
cd $GOPATH/src/trellocms;
go get && trellocms &
cd $BASE/test;
