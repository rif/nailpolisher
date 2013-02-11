#! /usr/bin/env sh
./build.sh

GOPATH=`pwd`

./bin/app &

inotifywait -m -r -e close_write src/ --exclude '~.*' | while read line
do 
	go build -o bin/app app && pkill app && ./bin/app &
done

