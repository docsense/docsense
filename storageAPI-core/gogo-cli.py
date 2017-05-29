#!/usr/bin/python2

import sys

reload(sys)
sys.setdefaultencoding("utf-8")
import requests
import json
import re

APP_TOKEN = "RpHeouBZsYVZ8EPyMTi8u9QtJ8sKwvs5c53PiNbibENalM6k9UqoH0WIs0TGB8UB"

SERVER = "http://localhost:8000"


def run_request(data, method, url, files=None, form=None):
    r = None
    if method == "DELETE":
        r = requests.request(method, url, json=data, headers={'X-Auth-Token': APP_TOKEN,
                                                              'X-DELETE-CASCADE': 'true'}, data=form, files=files)
    else:
        r = requests.request(method, url, json=data, headers={'X-Auth-Token':
                                                                  APP_TOKEN}, data=form, files=files)
    if r.status_code >= 400:
        print r.status_code
        """
    try:
        print json.dumps(json.loads(r.text), sort_keys=True,
                         indent=4, separators=(',', ': '))
    except:
    """
    print r.text
    return r


def get_metadata(pos):
    metadata = []
    current = []
    for x in sys.argv[pos:]:
        current.append(x)
        if len(current) == 2:
            try:
                metadata.append({"Key": current[0], "Value": str(
                    int(current[1])), "Type": "number"})
            except:
                metadata.append(
                    {"Key": current[0], "Value": current[1], "Type": "text"})
            current = []
    return metadata


def get_id_type():
    if len(sys.argv) <= 2:
        return "allpackages"
    if re.match("[a-zA-Z0-9]+\.[a-zA-Z0-9]+/[a-zA-Z0-9]+", sys.argv[2]):
        return "file"
    if re.match("[a-zA-Z0-9]+\.[a-zA-Z0-9]+", sys.argv[2]):
        return "package"
    return "idk"


def do_get():
    t = get_id_type()
    if t == "file":
        run_request(None, "GET", SERVER + "/file?id=" + sys.argv[2])
    elif t == "package":
        run_request(None, "GET", SERVER + "/package?id=" + sys.argv[2])
    elif t == "allpackages":
        run_request(None, "GET", SERVER + "/package")
    else:
        print "id type not recognized: " + sys.argv[2]


def do_post():
    t = get_id_type()
    if t == "package":
        package = sys.argv[2]
        filename = sys.argv[3]
        metadata = get_metadata(4)
        files = {'file': open(filename, 'rb')}
        form = {'package': package}
        run_request(metadata, "post", SERVER + "/file", files=files, form=form)
    else:
        metadata = get_metadata(2)
        run_request(metadata, "post", SERVER + "/package")


def do_search():
    run_request(sys.argv[2], "POST", SERVER + "/search")


def do_deleteall():
    r = run_request(None, "GET", SERVER + "/package")
    obj = json.loads(r.text)
    for i in obj:
        run_request(i, "DELETE", SERVER + "/package")


def do_getall():
    r = run_request(None, "GET", SERVER + "/package")
    l = json.loads(r.text)
    for pkg in l:
        run_request(None, "GET", SERVER + "/package?id=" + pkg)


def main():
    if sys.argv[1] == "get":
        do_get()
    elif sys.argv[1] == "post":
        do_post()
    elif sys.argv[1] == "search":
        do_search()
    elif sys.argv[1] == "deleteall":
        do_deleteall()
    elif sys.argv[1] == "getall":
        do_getall()
    else:
        print "unknown action " + sys.argv[1]


if __name__ == "__main__":
    main()
