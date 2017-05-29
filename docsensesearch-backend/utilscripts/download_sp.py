#!/usr/bin/python
# -*- coding: utf-8 -*-

import sys

import requests

reload(sys)
sys.setdefaultencoding('utf-8')
from bs4 import BeautifulSoup
import dateutil.parser
import pytz

link = sys.argv[1] + "/_api/web/lists(guid'" + sys.argv[2] + "')/Items"

LAST_MIGRATION_DATE_STR = sys.argv[3]
SP_LIST_ID = sys.argv[4]

processed_docs = 0

f = open("error.log", "w")

last_migration = pytz.utc.localize(dateutil.parser.parse(LAST_MIGRATION_DATE_STR))

while True:

    ids = []

    r = requests.get(link)

    bs = BeautifulSoup(r.text, features="xml")

    for i in bs.findAll('entry'):
        modified = i.findAll('Modified')[0].contents[0]
        modified_date = dateutil.parser.parse(modified)
        created = i.findAll('Created')[0].contents[0]
        created_date = dateutil.parser.parse(created)
        position = None
        year = None
        date = None
        try:
            position = i.findAll('ee')
            if not position:
                position = i.findAll('Pozycja')
            position = position[0].contents[0]
        except:
            print "position not found"

        try:
            year = i.findAll('te')
            if not year:
                year = i.findAll('Rok')
            year = year[0].contents[0]
        except:
            print "year not found"

        try:
            date = i.findAll('eeee')
            if not date:
                date = i.findAll('Data_x0020_wydania')
            date = date[0].contents[0]
        except:
            print "date not found"

        nazwa = "<nie znaleziono nazwy dokumentu>"
        try:
            nazwa = i.find('Dane_x0020_aktu_x0020_prawnego').contents[0]
        except:
            print "nazwa not found"

        id = i.find('properties').find('Id').contents[0]
        if modified_date > last_migration:
            file_tuple = (id, nazwa, created_date > last_migration, position, year, date)
            print file_tuple, created_date, last_migration
            ids.append(file_tuple)

    for i in ids:
        link = sys.argv[
                   1] + "/_api/web/lists(guid'" + sys.argv[2] + "')/Items(" + i[0] + ")/AttachmentFiles()"
        r = requests.get(link)
        bs2 = BeautifulSoup(r.text, features="xml")
        dllink = ""
        try:
            server_url = "/".join(sys.argv[1].split("/")[:3])
            dllinks = []
            for j in bs2.findAll('entry'):
                dllinks.append(server_url + j.find('properties').find('ServerRelativeUrl').contents[0])
        except Exception as e:
            print "0", e
        # print dllink.encode('utf-8'), "|", int(i[0]), "|", i[1].encode('utf-8'), "|", i[2].encode('utf-8')
        for dllink in dllinks:
            try:
                r = requests.get(dllink.encode('utf-8'), stream=True)
            except Exception as e:
                errstr = str(i[0]) + " DOWNLOAD LINK " + dllink + " FAILED WITH ERROR " + str(e)
                print errstr
                f.write(errstr + "\n")
                continue
            # r.raw
            files = {"file": r.raw}
            r2 = requests.post("http://localhost:8080/api/upload", files=files,
                               data={"dllink": dllink, "filename": i[1].encode('utf-8'), "spid": i[0],
                                     "splistid": SP_LIST_ID,
                                     "isnew": i[2], "position": i[3], "year": i[4], "date": i[5]})
            assert r2.status_code == 200
            # print skipped_docs + processed_docs, i[1].encode('utf-8')
            processed_docs += 1

    try:
        link = bs.findAll('link', {'rel': 'next'})[0]['href']
    except Exception as e:
        print "1", e
        break
f.close()
