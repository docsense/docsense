#!/usr/bin/python

import subprocess
import pymongo

client = pymongo.MongoClient()

subprocess.call(["mongo", "store", "--eval", "db.dropDatabase();"])

store = client.store

storagecore = store.storagecore
storagecore.insert_one({"_id": "maxdocument", "maxdocument": 0})

sp_lists = client.docsensesearch.sp_lists

sp_lists.drop()

id = 1

with open("idzienniki.txt") as f:
    for line in f:
        line = line[:-1]
        l = line.split()
        sp_lists.insert_one({"link": "http://dokumenty.uw.edu.pl/dziennik/" + l[0],
            "sp_id": l[1], "last_migration": "1970-01-01T01:00:00", "id": id})
        id += 1

sp_lists.insert_one({
    "link": "http://monitor.uw.edu.pl",
    "id": id,
    "sp_id": "ccac292c-c4f5-454b-9bcf-7d517cbf5af0",
    "last_migration": "1970-01-01T01:00:00"
})