#!/usr/bin/env bash

DBG_TEST=2
# Debug-level for app
DBG_APP=2
# Debug-level for server
DBG_SRV=3

. $GOPATH/src/gopkg.in/dedis/onet.v1/app/libtest.sh

main(){
    go build github.com/ineiti/cybermind
    startTest
#    test Build
    test List
    stopTest
    pkill -9 -f cybermind
}

testList(){
    runCM
    testGrep cli runCl module list
}

testBuild(){
    testOK runCl --help
}

runCl(){
    echo Running $@
    dbgRun ./cli -d $DBG_APP -c config $@
}

runCM(){
    pkill -9 -f cybermind
    rm -rf config/*
    ./cybermind -c config -d $DBG_SRV &
    while [ ! -p config/cli.fifo ]; do
        sleep .1
    done
}

main