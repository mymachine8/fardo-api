#! /bin/bash

# Copyright 2015 Google Inc. All rights reserved.
# Use of this source code is governed by the Apache 2.0
# license that can be found in the LICENSE file.

set -ex

TMP=$(mktemp -d -t gce-deploy-XXXXXX)

# [START cross_compile]
# Cross compile the app for linux/amd64
GOOS=linux GOARCH=amd64 go build -v -o $TMP/app ./app
# [END cross_compile]

# [START tar]
# Add the app binary

killall app || true

rm -rf /app/app

cp $TMP/* /app/

chmod +x /app/app

chown -R goapp:goapp /app

rm -rf $TMP