package data

import (
	"time"
	"github.com/mymachine8/fardo-api/models"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strings"
	"encoding/base64"
	"math"
)

func CreatePostUser(token string, post models.Post) (string, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this code
	post.GroupId = result.GroupId;
	post.UserId = result.UserId;
	if (len(post.GroupId) > 0 && post.IsGroup) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		if (err == nil) {
			post.GroupName = group.ShortName;
		}
	}
	if (len(post.LabelId) > 0 && post.IsGroup) {
		labelContext := common.NewContext()
		labelCol := labelContext.DbCollection("labels")
		var label models.Label
		err = labelCol.FindId(post.LabelId).One(&label)
		labelContext.Close()
		if (err == nil) {
			post.LabelName = label.Name;
		}
	}

	post.Id = bson.NewObjectId();


	if (len(post.ImageData) > 0 ) {
		fileName := "post_" + post.Id.Hex();
		imageReader := strings.NewReader(post.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		res, err := common.SendItemToCloudStorage(common.PostImage, fileName, dec);

		if (err != nil) {
			return "", models.FardoError{"Insert Post Image Error: " + err.Error()}
		}

		post.ImageUrl = res;

	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	post.IsActive = true;
	post.CreatedOn = time.Now()
	post.Score = redditPostRankingAlgorithm(post);

	err = c.Insert(&post)

	if (err != nil) {
		return "", models.FardoError{"Insert Post Error: " + err.Error()}
	}

	go addToCurrentPosts(post);

	go addToRecentUserPosts(result.UserId, post.Id, "post");

	go common.SendNearByNotification(post)

	go CalculateUserScore(post, ActionCreate);

	return post.Id.Hex(), err
}

func addToRecentUserPosts(userId bson.ObjectId, postId bson.ObjectId, fieldType string) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	ids := bson.M{
		"postIds": postId,
	}

	if (fieldType == "comment") {
		ids = bson.M{
			"commentPostIds": postId,
		}
	}
	_, err := c.Upsert(bson.M{"userId": userId},
		bson.M{"$push": ids})

	if (err != nil) {
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
	if (err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}

	post.UserId = result.UserId;

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

	if (len(post.ImageData) > 0 ) {
		fileName := "post_" + post.Id.Hex();
		imageReader := strings.NewReader(post.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		res, err := common.SendItemToCloudStorage(common.PostImage, fileName, dec);

		if (err != nil) {
			return "", models.FardoError{"Insert Post Image Error: " + err.Error()}
		}

		post.ImageUrl = res;

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
	if(err == nil) {
		go updateUserAndPostScore(id, ActionUpvote);
	}
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
	if(err == nil) {
		go updateUserAndPostScore(id, ActionDownvote);
	}
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

func updatePostScore(id string, score int) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"score": score,
		}})
	return
}

func updateUserAndPostScore(id string,actionType ActionType) {
	post, err := findPostById(id);
	 updatePostScore(id, redditPostRankingAlgorithm(post));

	if(err == nil) {
		 CalculateUserScore(post, actionType);
	}
}

func GetAllPosts(page int, postParams models.Post) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	skip := page*20;
	params := make(map[string]interface{})
	if(len(postParams.GroupName) > 0) {
		params["groupName"] = bson.RegEx{Pattern: postParams.GroupName, Options: "i"};
	}
	if(len(postParams.City) > 0) {
		params["city"] = bson.RegEx{Pattern: postParams.City, Options: "i"};
	}
	if(len(postParams.State) > 0) {
		params["state"] = bson.RegEx{Pattern: postParams.State, Options: "i"};
	}


	err = c.Find(params).Sort("-createdOn").Skip(skip).Limit(20).All(&posts)
	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func redditPostRankingAlgorithm(post models.Post) int {
	timeDiff := common.GetTimeSeconds(post.CreatedOn) - common.GetZingCreationTimeSeconds();
	votes := int64(post.Upvotes - post.Downvotes);
	var y int64;
	var z int64 = 1;
	if votes > 0 {
		y = 1
		z = votes

	}
	if votes == 0 {
		y = 0
	}
	if votes < 0 {
		y = -1
		z = votes * -1;
	}

	resultScore := math.Log10(float64(z)) + (float64(y) * float64(timeDiff))/45000;

	return int(resultScore);
}

func GetMyCirclePosts(token string, lat float64, lng float64) (posts[]models.Post, err error) {
 	//TODO: Reddit Alogirthm
	return;
}

func GetPopularPosts(token string, lat float64, lng float64) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	getNearByPopularPosts(lat,lng); //50%
	getGlobalPopularPosts(); //30%
	getPopularPostsAdminArea(lat, lng); //20%

	err = c.Find(nil).All(&posts)
	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func getNearByPopularPosts(lat float64, lng float64) {
	//TODO: Get NearBy popular posts
}

func getGlobalPopularPosts() {
	//TODO: Get Global popular posts
}

func getPopularPostsAdminArea(lat float64, lng float64) {
	//TODO: Get State popular posts
}

func GetLabelPosts(labelId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"labelId": bson.ObjectIdHex(labelId)}).All(&posts)
	if (posts == nil) {
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
	if (err != nil) {
		err = models.FardoError{"Get Access Token: " + err.Error()}
		return
	}

	var userPosts models.UserPost
	err = c.Find(bson.M{"userId": result.UserId}).One(&userPosts)
	if (err != nil) {
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

	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetGroupPosts(groupId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"groupId": bson.ObjectIdHex(groupId)}).All(&posts)
	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func GetCurrentPosts() (posts []models.PostLite, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("current_posts")

	err = c.Find(nil).All(&posts)

	if (posts == nil) {
		posts = []models.PostLite{}
	}
	return
}

func AddComment(token string, postId string, comment models.Comment) (string, error) {
	var err error
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	if (err != nil ) {
		return "", err
	}

	comment.Id = bson.NewObjectId()
	comment.IsActive = true;
	comment.PostId = bson.ObjectIdHex(postId);
	comment.CreatedOn = time.Now()
	comment.UserId = result.UserId;

	err = c.Insert(&comment)

	if (err == nil) {
		go addToRecentUserPosts(result.UserId, comment.PostId, "comment");
		post, err := findPostById(postId);
		if(err == nil ) {
			go common.SendCommentNotification(post, comment)
		}
	}

	return comment.Id.Hex(), err
}

func AddReply(token string, commentId string, reply models.Reply) (string, error) {
	var err error
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("access_tokens")
	var result models.AccessToken
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this algorithm
	reply.UserId = result.UserId;
	reply.CreatedOn = time.Now()

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(commentId)},
		bson.M{"$push": bson.M{"replies": reply}})

	if(err == nil) {
		var comment models.Comment;
		comment, err = findCommentById(commentId);
		common.SendReplyNotification(comment, reply)
	}

	return commentId, err
}

func UpvoteComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": 1,
		}})
	go checkVoteCount(id, true);
	return
}

func DownvoteComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": 1,
		}})

	go checkVoteCount(id, false);
	return
}

func checkVoteCount(id string, isUpvote bool) (err error) {
	post, err := findPostById(id);
	votes := post.Upvotes - post.Downvotes;

	if(isUpvote) {
		if(post.Upvotes == 1) {
			common.SendUpvoteNotification(post);
		}
		if(post.Upvotes == 2 || common.DivisbleByPowerOf2(post.Upvotes)) {
			common.SendUpvoteNotification(post);
		}
	}

	if(!isUpvote) {
		if (votes >= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendPost(id);
		}
	}
	return;
}

func ReportSpam(id string, reason string) (err error) {
	spamReason := bson.M{
		"commentPostIds": reason,
	}
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"spamCount": 1},
			"$push": spamReason, })
	if(err == nil) {
		go updateUserAndPostScore(id, ActionSpam);
	}

	return;
}

func checkSpamCountLimit(id string) (err error) {
	post, err := findPostById(id);
	if (post.SpamCount >= models.SPAM_COUNT_LIMIT) {
		err = SuspendPost(id);
		if(err != nil) {
			common.SendDeletePostNotification(post);
		}
	}
	return
}

func findPostById(id string) (post models.Post, err error, ) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.FindId(bson.ObjectIdHex(id)).One(&post);
	return
}

func findCommentById(id string) (comment models.Comment, err error, ) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.FindId(bson.ObjectIdHex(id)).One(&comment);
	return
}

func GetAllComments(postId string) (comments []models.Comment, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Find(bson.M{"postId": bson.ObjectIdHex(postId)}).All(&comments)

	if (comments == nil) {
		comments = []models.Comment{}
	}
	return
}

