#! /bin/bash

# Copyright 2015 Google Inc. All rights reserved.
# Use of this source code is governed by the Apache 2.0
# license that can be found in the LICENSE file.

set -ex

TMP=$(mktemp -d -t gce-deploy-XXXXXX)

# [START cross_compile]
# Cross compile the app for linux/amd64
GOOS=linux GOARCH=amd64 godep go build -v -o $TMP/app ../app
# [END cross_compile]

# [START tar]
# Add the app binary
tar -c -f $TMP/bundle.tar -C $TMP app
# [END tar]

# [START gcs_push]
# FARDO_DEPLOY_LOCATION is something like "gs://my-bucket/FARDO-VERSION.tar".
gsutil cp $TMP/bundle.tar "gs://go-server/fardo-beta.tar"
# [END gcs_push]

rm -rf $TMP
