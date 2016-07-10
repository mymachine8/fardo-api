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
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this algorithm
	post.GroupId = result.GroupId;
	post.UserId = result.UserId
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

	post.Id = bson.NewObjectId()
	post.IsActive = true;
	post.CreatedOn = time.Now()

	err = c.Insert(&post)

	if(err != nil) {
		return "", models.FardoError{"Insert Post Error: " + err.Error()}
	}

	go addToCurrentPosts(post);

	go addToRecentUserPosts(result.UserId, post.Id, "post");

	return post.Id.Hex(), err
}

func addToRecentUserPosts(userId bson.ObjectId, postId bson.ObjectId, fieldType string) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	ids := bson.M{
		"postIds": postId,
	}

	if(fieldType == "comment") {
		ids = bson.M{
			"commentPostIds": postId,
		}
	}
	_, err := c.Upsert(bson.M{"userId": userId},
		bson.M{"$push": ids})

	if(err != nil) {
		log.Print(err.Error());
	}
}

func CreatePostAdmin(token string, post models.Post) (string, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}

	if (len(post.GroupId) > 0) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		if (err == nil) {
			post.GroupName = group.Name;
			post.Loc[0] = group.Loc[0];
			post.Loc[1] = group.Loc[1];
		}
	}

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
	if(posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetLabelPosts(labelId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"labelId": bson.ObjectIdHex(labelId)}).All(&posts)
	if(posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetRecentUserPosts(token string, contentType string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err != nil) {
		err = models.FardoError{"Get Access Token: " + err.Error()}
		return
	}

	var userPosts models.UserPost
	err = c.Find(bson.M{"userId": result.UserId}).One(&userPosts)
	if(err != nil) {
		err = models.FardoError{"Get User Posts: " + err.Error()}
		return
	}

	postContext := common.NewContext()
	defer postContext.Close()
	postCol := postContext.DbCollection("posts")

	var postIds [] bson.ObjectId;

	postIds = userPosts.PostIds;
	if contentType == "comment" {
		postIds = userPosts.CommentPostIds;
	}

	err = postCol.Find(bson.M{"_id": bson.M{"$in": postIds}}).All(&posts)

	if(posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetGroupPosts(groupId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"groupId": bson.ObjectIdHex(groupId)}).All(&posts)
	if(posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetCurrentPosts() (posts []models.PostLite, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("current_posts")

	err = c.Find(nil).All(&posts)

	if(posts == nil) {
		posts = []models.PostLite{}
	}
	return
}

func AddComment(token string, comment models.Comment) (string, error) {
	var err error
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")
	comment.Id = bson.NewObjectId()
	comment.IsActive = true;
	comment.CreatedOn = time.Now()

	err = c.Insert(&comment)

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if(err == nil) {
		go addToRecentUserPosts(result.UserId, comment.PostId, "comment");
	}

	return comment.Id.Hex(), err
}

func UpvoteComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": 1,
		}})
	return
}

func DownvoteComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": 1,
		}})
	return
}

func GetAllComments(postId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Find(bson.M{"postId": bson.ObjectIdHex(postId)}).All(&posts)

	if(posts == nil) {
		posts = []models.Post{}
	}
	return
}

