// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Sample bookshelf is a fully-featured app demonstrating several Google Cloud APIs, including Datastore, Cloud SQL, Cloud Storage.
// See https://cloud.google.com/go/getting-started/tutorial-app
package main

import (
	"github.com/mymachine8/fardo-api/common"
	"github.com/mymachine8/fardo-api/controllers"
	"google.golang.org/appengine"
)

func main() {
	common.StartUp();
	controllers.InitRoutes()
	appengine.Main()
}