#!/usr/bin/python

import json
import subprocess

config = {}

with open("config.json") as f:
    config = json.loads(f.read())

subprocess.call(["aws", "s3", "--endpoint-url",
                 "http://ceph-radosgw.net.uw.edu.pl:7480", "rb", "--force",
                 "s3://" + config["InstancePrefix"]])

subprocess.call(["aws", "s3", "--endpoint-url",
                 "http://ceph-radosgw.net.uw.edu.pl:7480", "mb",
                 "s3://" + config["InstancePrefix"]])
