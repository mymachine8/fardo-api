package dbconn

import (
	"gopkg.in/mgo.v2"
	"fmt"
)


type mongoSession struct {
	session *mgo.Session
	hosts []string
	username string
	password string
	database string
}

var instance *mongoSession

func GetInstance() *mongoSession {
	if instance == nil {
		instance = &mongoSession{}   // <--- NOT THREAD SAFE
		instance.initializeDBConnection();
		instance.database = "fardo"
	}
	return instance;
}

func (ms mongoSession) initializeDBConnection() {
	/*mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    ms.hosts,
		Timeout:  60 * time.Second,
		Database: ms.database,
		Username: ms.username,
		Password: ms.password,
	}*/

	session, err := mgo.Dial("localhost:27017");


	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	//session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		fmt.Printf("CreateSession: %s\n", err)
		return;
	}

	ms.session = session;

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	ms.session.SetMode(mgo.Monotonic, true)
	ms.initializeIndexes();
}

func  (ms mongoSession) initializeIndexes(){
	index := mgo.Index{
		Key: []string{"$2d:loc"},
		Bits: 26,
	}
	collection := ms.session.DB(ms.database).C("posts");
	err := collection.EnsureIndex(index)
	if(err != nil) {
		fmt.Printf(err.Error());
	}
}

func (ms mongoSession) GetSession() *mgo.Session{
	return ms.session;
}

func (ms mongoSession) GetDatabaseName() string {
	return ms.database;
}