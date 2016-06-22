package dbconn

import (
	"gopkg.in/mgo.v2"
	"fmt"
	"time"
	"log"
	"github.com/mymachine8/fardo-api/models"
	"github.com/mymachine8/fardo-api/common"
)

const (
	MongoDBHosts = "104.155.143.185"
	AuthDatabase = "fardo"
)

type mongoSession struct {
	Session *mgo.Session
	Hosts []string
	Username string
	Password string
	Database string
}

var instance *mongoSession

func GetInstance() *mongoSession {
	if instance == nil {
		instance = &mongoSession{}   // <--- NOT THREAD SAFE
		instance.Hosts = []string {MongoDBHosts}
		instance.Database = AuthDatabase
		instance.initializeDBConnection();
	}
	return instance;
}

func (ms mongoSession) initializeDBConnection() {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    ms.Hosts,
		Timeout:  30 * time.Second,
		Database: ms.Database,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo);

	if err != nil {
		fmt.Printf("CreateSession: %s\n", err)
		return;
	}

	ms.Session = session;

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	ms.Session.SetMode(mgo.Monotonic, true)
	ms.initializeIndexes();
}

func  (ms mongoSession) initializeIndexes(){
	index := mgo.Index{
		Key: []string{"$2d:loc"},
		Bits: 26,
	}
	collection := ms.Session.DB(ms.Database).C("posts");
	err := collection.EnsureIndex(index)
	if(err != nil) {
		fmt.Printf(err.Error());
	}

	var categories [] models.Category;
	collection = ms.Session.DB(ms.Database).C("categories");
	query := collection.Find(nil)
	log.Printf("came here");
	err = query.All(&categories);
	if(err != nil) {
		fmt.Printf(err.Error());
	}

	fmt.Printf("%+v\n", categories);

}