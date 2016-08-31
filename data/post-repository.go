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

const (
	PopularPostsLimit int = 10

	LocalPercent = 5

	AdminAreaPercent = 2

	GlobalPercent = 3
)

func CreatePostUser(token string, post models.Post) (models.Post, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return post, models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this code
	post.UserId = result.Id;
	if(!post.IsAnonymous) {
		post.Username = result.Username;
	}
	if (post.IsGroup) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		if (err == nil) {
			post.GroupName = group.ShortName;
			post.GroupCategoryName = group.CategoryName
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
			return post, models.FardoError{"Insert Post Image Error: " + err.Error()}
		}

		post.ImageUrl = res;

	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	post.IsActive = true;
	post.CreatedOn = time.Now().UTC();
	post.Score = redditPostRankingAlgorithm(post);

	err = c.Insert(&post)

	if (err != nil) {
		return post, models.FardoError{"Insert Post Error: " + err.Error()}
	}

	if(post.IsGroup) {
		post.PlaceName = post.GroupName
		post.PlaceType = post.GroupCategoryName
	} else {
		post.PlaceName = post.Locality
		post.PlaceType = "location"
	}

	go addToCurrentPosts(post);

	go addToRecentUserPosts(result.Id, post.Id, "post");

	go common.SendNearByNotification(post)

	go CalculateUserScore(post, ActionCreate);

	return post, err
}

func GetPostById(id string) (post models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.FindId(bson.ObjectIdHex(id)).One(&post)
	return
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
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}

	post.UserId = result.Id;

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
	post.CreatedOn = time.Now().UTC();
	post.Score = redditPostRankingAlgorithm(post);

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

func UpvotePost(id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	step := 1;
	if(undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
		}, "$set": bson.M{
				"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndPostScore(id, ActionUpvote);
		go checkVoteCount(id, true);
	}
	return
}

func DownvotePost(id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	step := 1;

	if(undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndPostScore(id, ActionDownvote);
		go checkVoteCount(id, false);
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
			"modifiedOn": time.Now().UTC(),
		}})
	return
}

func SuspendComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})
	return
}

func updatePostScore(id string, score float64) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"score": score,
		}})
	return
}

func updateUserAndPostScore(id string, actionType ActionType) {
	post, err := findPostById(id);
	updatePostScore(id, redditPostRankingAlgorithm(post));

	if (err == nil) {
		CalculateUserScore(post, actionType);
	}
}

func GetAllPosts(page int, postParams models.Post) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	skip := page * 20;
	params := make(map[string]interface{})
	if (len(postParams.GroupName) > 0) {
		params["groupName"] = bson.RegEx{Pattern: postParams.GroupName, Options: "i"};
	}
	if (len(postParams.City) > 0) {
		params["city"] = bson.RegEx{Pattern: postParams.City, Options: "i"};
	}
	if (len(postParams.State) > 0) {
		params["state"] = bson.RegEx{Pattern: postParams.State, Options: "i"};
	}

	params["isActive"] = true;

	err = c.Find(params).Sort("-createdOn").Skip(skip).Limit(20).All(&posts)
	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func redditPostRankingAlgorithm(post models.Post) float64 {
	timeDiff := common.GetTimeSeconds(post.CreatedOn) - common.GetZingCreationTimeSeconds();
	votes := int64(post.Upvotes - post.Downvotes);
	var sign int64;
	var z int64 = 1;
	if votes > 0 {
		sign = 1
		z = votes

	}
	if votes == 0 {
		sign = 0
	}
	if votes < 0 {
		sign = -1
		z = votes * -1;
	}

	return float64(sign) * math.Log2(float64(z)) + float64(timeDiff) / 45000;
}

func GetMyCirclePosts(token string, lat float64, lng float64, lastUpdated time.Time, groupId string) (posts[]models.Post, err error) {
	//TODO: Location context is missing
	context := common.NewContext()
	c := context.DbCollection("posts")
	defer context.Close()

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		err = models.FardoError{"Get Access Token: " + err.Error()}
		return
	}

	if(len(groupId) > 0) {
		params := make(map[string]interface{})
		params["groupId"] = result.GroupId;
		err = c.Find(params).Limit(150).Sort("score").All(&posts);

	}else {
		currentLatLng := [2]float64{lng, lat}


		params := make(map[string]interface{});
		params = bson.M{"loc":
		bson.M{"$geoWithin":
		bson.M{"$centerSphere": []interface{}{currentLatLng, 2 / 3963.2} }},
			"createdOn": bson.M{"$gt": lastUpdated}};
		if (len(result.GroupId) > 0) {
			params["groupId"] = result.GroupId;
		}

		err = c.Find( bson.M{"$or":[]bson.M {params}}).Limit(150).Sort("score").All(&posts);
	}

	for index, _ := range posts {
		if((posts[index].GroupId.Hex() == groupId) || (posts[index].GroupId.Hex() == result.GroupId.Hex())) {
			posts[index].PlaceName = posts[index].GroupName;
			posts[index].PlaceType = posts[index].GroupCategoryName;
		} else {
			posts[index].PlaceName = posts[index].Locality;
			posts[index].PlaceType = "location"
		}
	}

	if (posts == nil) {
		posts = []models.Post{}
	}

	return
}

func GetMyCircleUpdates(token string, lat float64, lng float64, lastUpdated time.Time, fromDate time.Time) (posts[]models.Post, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("posts")

	var result []struct {
		Id         bson.ObjectId `bson:"_id" json:"id"`
		Upvotes    int  `bson:"upvotes" json:"upvotes"`
		Downvotes  int  `bson:"downvotes" json:"downvotes"`
		ReplyCount int `bson:"replyCount" json:"replyCount"`
		IsActive   bool `bson:"isActive" json:"isActive"`
	}
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 2 / 3963.2} }},
		"createdOn": bson.M{"$gt": fromDate, "$lt": lastUpdated},
		}).Sort("score").Limit(150).All(&result);
	return
}

func GetPopularPosts(token string, lat float64, lng float64) (posts []models.Post, err error) {

	nearByPosts, err := getNearByPopularPosts(lat, lng); //50%
	globalPosts, _ := getGlobalPopularPosts(); //30%
	adminAreaPosts, _ := getPopularPostsAdminArea(lat, lng); //20%

	nearByPostsLen := common.MinInt(len(nearByPosts), PopularPostsLimit);

	for index, _ := range nearByPosts {
		if(len(nearByPosts[index].GroupName) > 0) {
			nearByPosts[index].PlaceName = nearByPosts[index].GroupName;
			nearByPosts[index].PlaceType = nearByPosts[index].GroupCategoryName;
		} else {
			nearByPosts[index].PlaceName = nearByPosts[index].Locality;
			nearByPosts[index].PlaceType = "location"
		}
	}

	if (nearByPostsLen <= LocalPercent) {
		return nearByPosts, err
	}

	posts = nearByPosts[:LocalPercent]

	var count int = 0;
	for _, glb := range globalPosts {
		if(len(glb.GroupName) > 0) {
			glb.PlaceName = glb.GroupName;
			glb.PlaceType = glb.GroupCategoryName;
		} else {
			glb.PlaceName = glb.City;
			glb.PlaceType = "location"
		}
		if (!idInPosts(glb.Id.Hex(), posts) && count < GlobalPercent) {
			posts = append(posts, glb)
			count++;
		}
	}

	count = 0;
	for _, aa := range adminAreaPosts {
		if(len(aa.GroupName) > 0) {
			aa.PlaceName = aa.GroupName;
			aa.PlaceType = aa.GroupCategoryName;
		} else {
			aa.PlaceName = aa.City;
			aa.PlaceType = "location"
		}
		if (!idInPosts(aa.Id.Hex(), posts) && count < AdminAreaPercent) {
			posts = append(posts, aa)
			count++;
		}
	}

	resLen := PopularPostsLimit - len(posts)

	j := LocalPercent
	for i := 0; i < resLen && j < nearByPostsLen; i++ {
		posts = append(posts, nearByPosts[j])
		j++;
	}

	if (posts == nil) {
		posts = []models.Post{}
	}
	return
}

func idInPosts(id string, list []models.Post) bool {
	for _, b := range list {
		if b.Id.Hex() == id {
			return true
		}
	}
	return false
}

func getNearByPopularPosts(lat float64, lng float64) (posts[]models.Post, err error) {

	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("posts")
	now := time.Now().UTC()
	then := now.AddDate(0, -3, 0)
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }},
		"createdOn": bson.M{"$gt": then}}).Sort("score").All(&posts);
	if (posts == nil) {
		posts = []models.Post{}
	}
	return;
}

func getGlobalPopularPosts() (posts[]models.Post, err error) {

	context := common.NewContext()
	defer context.Close()

	now := time.Now().UTC()
	then := now.AddDate(0, -7, 0)
	c := context.DbCollection("posts")
	err = c.Find(bson.M{
		"createdOn": bson.M{"$gt": then}}).Sort("score").All(&posts);
	if (posts == nil) {
		posts = []models.Post{}
	}
	return;
}

func getPopularPostsAdminArea(lat float64, lng float64) (posts[]models.Post, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	now := time.Now().UTC()
	then := now.AddDate(0, -4, 0)
	c := context.DbCollection("posts")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 300 / 3963.2} }},
		"createdOn": bson.M{"$gt": then}}).Sort("score").All(&posts);
	if (posts == nil) {
		posts = []models.Post{}
	}
	return;
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
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		err = models.FardoError{"Get Access Token: " + err.Error()}
		return
	}

	var userPosts models.UserPost
	err = c.Find(bson.M{"userId": result.Id}).One(&userPosts)
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
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	if (err != nil ) {
		return "", err
	}

	comment.Id = bson.NewObjectId()
	comment.IsActive = true;
	comment.PostId = bson.ObjectIdHex(postId);
	comment.CreatedOn = time.Now()
	comment.UserId = result.Id;

	err = c.Insert(&comment)

	if (err == nil) {
		go addToRecentUserPosts(result.Id, comment.PostId, "comment");
		post, err := findPostById(postId);
		if (err == nil ) {
			go updateReplyCount(postId);
			go common.SendCommentNotification(post, comment)
		}
	}

	return comment.Id.Hex(), err
}

func updateReplyCount(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"replyCount": 1,
		},"$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	return
}

func AddReply(token string, commentId string, reply models.Reply) (string, error) {
	var err error
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return "", models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this code
	reply.UserId = result.Id;
	reply.CreatedOn = time.Now()

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(commentId)},
		bson.M{"$push": bson.M{"replies": reply}})

	return commentId, err
}

func UpvoteComment(id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	step := 1;
	if(undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
		},"$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	go checkCommentVoteCount(id, true);
	return
}

func DownvoteComment(id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	step := 1;
	if(undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": step,
		},"$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	go checkCommentVoteCount(id, false);
	return
}

func checkCommentVoteCount(id string, isUpvote bool) (err error) {
	comment, err := findCommentById(id);
	votes := comment.Upvotes - comment.Downvotes;

	if (isUpvote) {
		if (comment.Upvotes == 1 || common.DivisbleByPowerOf2(comment.Upvotes)) {
			common.SendCommentUpvoteNotification(comment);
		}
	}

	if (!isUpvote) {
		if (votes >= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendComment(id);
		}
	}
	return;
}

func checkVoteCount(id string, isUpvote bool) (err error) {
	post, err := findPostById(id);
	votes := post.Upvotes - post.Downvotes;

	if (isUpvote) {
		if (post.Upvotes == 1 || common.DivisbleByPowerOf2(post.Upvotes)) {
			common.SendUpvoteNotification(post);
		}
	}

	if (!isUpvote) {
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
	if (err == nil) {
		go updateUserAndPostScore(id, ActionSpam);
	}

	return;
}

func checkSpamCountLimit(id string) (err error) {
	post, err := findPostById(id);
	if (post.SpamCount >= models.SPAM_COUNT_LIMIT) {
		err = SuspendPost(id);
		if (err != nil) {
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

