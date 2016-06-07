package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2/bson"
	"log"
)

func CreatePostUser(token string, post models.Post) (string, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err != nil) {
		return "", err
	}
	post.GroupId = result.GroupId;
	if (len(post.LabelId) > 0) {
		labelContext := common.NewContext()
		labelCol := labelContext.DbCollection("labels")
		var label models.Label
		err = labelCol.FindId(post.LabelId).One(&label)
		labelContext.Close()
		if (err == nil) {
			post.LabelName = label.Name;
		}
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	obj_id := bson.NewObjectId()
	post.Id = obj_id
	post.IsActive = true;
	post.CreatedOn = time.Now()

	err = c.Insert(&post)
	go addToCurrentPosts(post);

	return obj_id.Hex(), err
}

func CreatePostAdmin(token string, post models.Post) (string, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err != nil) {
		return "", err
	}
	log.Print(len(post.GroupId))
	if (len(post.GroupId) > 0) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		if (err == nil) {
			post.GroupName = group.Name;
		}
	}
	log.Print(len(post.LabelId))
	if (len(post.LabelId) > 0) {
		labelContext := common.NewContext()
		labelCol := labelContext.DbCollection("labels")
		var label models.Label
		err = labelCol.FindId(post.LabelId).One(&label)
		labelContext.Close()
		if (err == nil) {
			post.LabelName = label.Name;
		}
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	obj_id := bson.NewObjectId()
	post.Id = obj_id
	post.IsActive = true;
	post.CreatedOn = time.Now()

	err = c.Insert(&post)
	go addToCurrentPosts(post);

	return obj_id.Hex(), err
}

func addToCurrentPosts(post models.Post) {
	var postLite models.PostLite
	context := common.NewContext()
	defer context.Close()
	postLite.Id = post.Id
	postLite.GroupId = post.GroupId
	postLite.LabelId = post.LabelId
	postLite.Loc = post.Loc
	postLite.CreatedOn = post.CreatedOn
	c := context.DbCollection("current_posts")
	err := c.Insert(&postLite)
	if (err != nil) {
		log.Print("Error Inserting to Current Post " + err.Error())
	}
}

func UpvotePost(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": 1,
		}})
	return
}

func DownvotePost(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": 1,
		}})
	return
}

func SuspendPost(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
		}})
	return
}

func GetAllPosts() (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(nil).All(&posts)
	return
}

func GetCurrentPosts() (posts []models.PostLite, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("current_posts")

	err = c.Find(nil).All(&posts)
	return
}

