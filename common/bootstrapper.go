package common

func StartUp() {

	initConfig()

	// Start a MongoDB session
	createDbSession()

	InitCloudStorage()

	InitCronJobs()

}
