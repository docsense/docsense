#!/usr/bin/python

import random
import string

print ''.join(
    random.SystemRandom().choice(string.ascii_uppercase + string.digits + string.ascii_lowercase) for _ in range(64))
