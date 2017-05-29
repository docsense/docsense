#!/usr/bin/python
import sys

import requests

reload(sys)
sys.setdefaultencoding('utf-8')
from bs4 import BeautifulSoup

link = sys.argv[1] + "/_api/web/lists(guid'" + sys.argv[2] + "')/Items"
ids = []
try:
    r = requests.get(link)
    bs = BeautifulSoup(r.text, features="xml")
    for i in bs.findAll('entry'):
        for j in i.findAll('properties'):
            nazwa = "<nie znaleziono nazwy dokumentu>"
            try:
                nazwa = j.find('Dane_x0020_aktu_x0020_prawnego').contents[0]
            except:
                print "nazwa not found"
            id = j.find('Id').contents[0]
            ids.append((id, nazwa))
    print len(ids)
except:
    print "-1"
