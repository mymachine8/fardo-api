package common

func StartUp() {

	initConfig()
	// Initialize private/public keys for JWT authentication
	initKeys()
	// Start a MongoDB session
	createDbSession()

	InitCloudStorage()

	InitCronJobs()

}
