#!/usr/bin/python
# -*- coding: utf-8 -*-

import os
import sys

import requests

reload(sys)
sys.setdefaultencoding('utf-8')

APP_TOKEN = "qFFCdTvLtId6FvDeUlr8BIlXxNBBffZOotEuRDtMp7SQXlQgE4Bd4ZW0HiGVKMM7"

upload_dir = sys.argv[1]

print "upload_dir: " + upload_dir

file_list = []

for root, dirs, files in os.walk(upload_dir):
    for file in files:
        if file[-4:] == ".pdf":
            file_list.append(file)

file_list = sorted(file_list)

for pdf in file_list:
    with open(upload_dir + "/" + pdf) as f:
        files = {"file": f}
        print "STARTING", pdf
        r2 = requests.post("http://localhost:8080/api/upload", files=files)
        assert r2.status_code == 200
        print "DONE", pdf

print len(file_list)
