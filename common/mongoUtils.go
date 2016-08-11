package common

import (
	"log"
	"time"
	"gopkg.in/mgo.v2"
	"github.com/mymachine8/fardo-api/slack"
)

var session *mgo.Session

func GetSession() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:    []string{AppConfig.MongoDBHost},
			Username: AppConfig.DBUser,
			Password: AppConfig.DBPwd,
			Timeout:  30 * time.Second,
		})
		if err != nil {
			slack.Send(slack.PanicLevel, err.Error());
			log.Panic("[GetSession]: %s\n", err)
		}
	}
	return session
}

func createDbSession() {
	var err error
	session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{AppConfig.MongoDBHost},
		Username: AppConfig.DBUser,
		Password: AppConfig.DBPwd,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		slack.Send(slack.PanicLevel, err.Error());
		log.Panic("[createDbSession]: %s\n", err)
	}
	addIndexes()
}

// Add indexes into MongoDB
func addIndexes() {
	var err error
	userIndex := mgo.Index{
		Key:        []string{"token"},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}

	groupIndex := mgo.Index{
		Key:        []string{"$text:name"},
	}

	postIndex := mgo.Index{
		Key: []string{"$2d:loc"},
		Bits: 26,
	}

	currentPostTTL := mgo.Index{
		Key:         []string{"createdOn"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 400}

	// Add indexes into MongoDB
	session := GetSession().Copy()
	defer session.Close()
	userCol := session.DB(AppConfig.Database).C("users")
	postCol := session.DB(AppConfig.Database).C("posts")
	groupCol := session.DB(AppConfig.Database).C("groups")
	currentPostCol := session.DB(AppConfig.Database).C("current_posts")


	err = userCol.EnsureIndex(userIndex)
	if err != nil {
		log.Print("[addIndexes]: %s\n", err.Error())
	}
	err = postCol.EnsureIndex(postIndex)
	if err != nil {
		log.Print("[addIndexes]: %s\n", err.Error())
	}
	err = groupCol.EnsureIndex(groupIndex)
	if err != nil {
		log.Print("[addIndexes]: %s\n", err.Error())
	}
	if err := currentPostCol.EnsureIndex(currentPostTTL); err != nil {
		log.Print("[addIndexes]: %s\n", err.Error())
	}
}


// Struct used for maintaining HTTP Request Context
type Context struct {
	MongoSession *mgo.Session
}

// Close mgo.Session
func (c *Context) Close() {
	c.MongoSession.Close()
}

// Returns mgo.collection for the given name
func (c *Context) DbCollection(name string) *mgo.Collection {
	return c.MongoSession.DB(AppConfig.Database).C(name)
}

// Create a new Context object for each HTTP request
func NewContext() *Context {
	session := GetSession().Copy()
	context := &Context{
		MongoSession: session,
	}
	return context
}
