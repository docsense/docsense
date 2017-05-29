#!/usr/bin/python

import re
from subprocess import check_output
from datetime import datetime
import dateutil.parser
import pymongo

client = pymongo.MongoClient()

p = re.compile("guid\'[a-f0-9\-]+\'")

sp_lists = client.docsensesearch.sp_lists

# c.execute("select link, sp_id, last_migration, id from sp_lists")
for i in sp_lists.find(None):
    print "processing", i["link"]
    last_migration = dateutil.parser.parse(i["last_migration"])
    exec_list = ["./download_sp.py", str(i["link"]), str(i["sp_id"]), last_migration.isoformat(), str(i["id"])]
    out = check_output(exec_list)
    print "OUT", out
    sp_lists.replace_one({"id": i["id"]}, {"last_migration": datetime.now().isoformat(),
        "link": i["link"],
        "sp_id": i["sp_id"],
        "id": i["id"],
    })
