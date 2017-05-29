#!/bin/bash

touch $GOPATH/src/storage-fulltext/main.go
cd $GOPATH/src/storage-core
./reset_instance_files.py
cd $GOPATH/src/docsensesearch/utilscripts
./reset_mongo.py
# jeszcze fulltext
while true;
do
	sleep 1;
	curl http://localhost:8888/;
	if [ $? == 0 ]; then
		break
	fi
done
./crawler.py
