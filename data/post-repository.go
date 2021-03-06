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
	"sort"
)

const (
	PopularPostsLimit int = 100

	LocalPercent = 40

	AdminAreaPercent = 20

	GlobalPercent = 40
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
	if (!post.IsAnonymous) {
		post.Username = result.Username;
		if (len(result.Phone) > 0) {
			post.Jid = result.Phone + "@im.ripplin.in"
		}
	}
	if (post.IsGroup) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		post.GroupName = group.ShortName;
		post.GroupCategoryName = group.CategoryName
		UpdateGroupPostCount(group.Id.Hex());
	}

	post.Id = bson.NewObjectId();


	if (len(post.ImageData) > 0 ) {
		fileName := "post_" + post.Id.Hex();
		imageReader := strings.NewReader(post.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		var fileType string
		if (len(post.ImageType) > 0) {
			fileType = post.ImageType
		} else {
			fileType = "jpeg"
		}

		res, err := common.SendItemToCloudStorage(common.PostImage, fileName, fileType, dec);

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

	if (post.IsGroup) {
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
	if (err == nil) {
		if (post.IsGroup) {
			post.PlaceName = post.GroupName
			post.PlaceType = post.GroupCategoryName
		} else {
			post.PlaceName = post.Locality
			post.PlaceType = "location"
		}
	}
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

func addToRecentUserVotes(userId bson.ObjectId, id bson.ObjectId, isUpvote bool, isUndo bool, fieldType string) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	ids := bson.M{};
	if (fieldType == "comment") {
		ids = bson.M{
			"$pull": bson.M{"commentVotes" : bson.M{"_id" : id}}}
	} else {
		ids = bson.M{
			"$pull": bson.M{"postVotes" : bson.M{"_id" : id}}}
	}
	err := c.Update(bson.M{"userId": userId},
		ids)

	if (isUndo) {
		return;
	}

	var userVote models.UserVote;
	userVote.Id = id;
	userVote.IsUpvote = isUpvote;

	ids = bson.M{
		"postVotes": userVote,
	}

	if (fieldType == "comment") {
		ids = bson.M{
			"commentVotes": userVote,
		}
	}
	_, err = c.Upsert(bson.M{"userId": userId},
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

	if (len(post.Username) > 0) {
		isAvailable, errr := CheckUsernameAvailability(post.Username)
		if (!isAvailable || errr != nil) {
			return "", models.FardoError{"Username is not available"}
		}
	}

	if (len(post.GroupId) > 0) {
		groupContext := common.NewContext()
		groupCol := groupContext.DbCollection("groups")
		var group models.Group
		err = groupCol.FindId(post.GroupId).One(&group)
		groupContext.Close()
		if (err == nil) {
			post.GroupName = group.ShortName;
			post.GroupCategoryName = group.CategoryName
			post.Loc[0] = group.Loc[0];
			post.Loc[1] = group.Loc[1];
			UpdateGroupPostCount(group.Id.Hex());
		}
	}

	post.Id = bson.NewObjectId();

	if (len(post.ImageData) > 0 ) {
		fileName := "post_" + post.Id.Hex();
		imageReader := strings.NewReader(post.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		res, err := common.SendItemToCloudStorage(common.PostImage, fileName, "jpeg", dec);

		if (err != nil) {
			return "", models.FardoError{"Insert Post Image Error: " + err.Error()}
		}

		post.ImageUrl = res;

	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	post.IsActive = true;
	if (!post.CreatedOn.IsZero()) {
		post.CreatedOn = post.CreatedOn.UTC();
	}
	post.CreatedOn = time.Now().UTC();
	post.Score = redditPostRankingAlgorithm(post);

	err = c.Insert(&post)
	go addToCurrentPosts(post);

	return post.Id.Hex(), err
}

func addToCurrentPosts(post models.Post) {
	var postLite models.PostLite
	context := common.NewContext()
	defer context.Close()
	postLite.Id = post.Id
	postLite.GroupId = post.GroupId
	postLite.Loc = post.Loc
	postLite.CreatedOn = post.CreatedOn
	c := context.DbCollection("current_posts")
	err := c.Insert(&postLite)
	if (err != nil) {
		log.Print("Error Inserting to Current Post " + err.Error())
	}
}

func UpvotePost(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;
	voteStep := 1;
	if (undo) {
		step = -1;
		voteStep = -1;
	}

	voteType := getPostVoteType(result.Id, id);

	downvoteStep := 0;
	if(voteType == "downvote") {
		downvoteStep = -1;
		voteStep = 2;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
			"downvotes": downvoteStep,
			"votes": voteStep,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndPostScore(id, ActionUpvote);
		go checkVoteCount(result.Token, result.Id.Hex(), id, true);
		go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), true, undo, "post");
	}
	return
}

func DownvotePost(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;
	voteStep := -1;
	if (undo) {
		step = -1;
		voteStep = 1;
	}

	voteType := getPostVoteType(result.Id, id);

	upvoteStep := 0;
	if(voteType == "upvote") {
		upvoteStep = -1;
		voteStep = -2;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": upvoteStep,
			"downvotes": step,
			"votes": voteStep,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndPostScore(id, ActionDownvote);
		go checkVoteCount(result.Token, result.Id.Hex(), id, false);
		go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), false, undo, "post");
	}
	return
}

func SuspendPost(id string, isSilent bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})

	if (err == nil && !isSilent) {
		post, _ := findPostById(id)
		common.SendDeletePostNotification(post.UserId.Hex(), post.Content);
	}
	return
}

func SuspendComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	if (err != nil) {
		return
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})

	if (err == nil ) {
		comment, errr := findCommentById(id);
		if (errr != nil) {
			return
		}
		go updateReplyCount(comment.PostId.Hex(), false);
	}

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

	return float64(sign) * math.Log10(float64(z)) + float64(timeDiff) / 40000;
}

func GetMyCirclePosts(token string, lat float64, lng float64, lastUpdated time.Time) (posts[]models.Post, group *models.Group, err error) {
	context := common.NewContext()
	c := context.DbCollection("posts")
	defer context.Close()

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		err = models.FardoError{"Get Access Token: " + token + " " + err.Error()}
		return
	}

	group, err = GetNearGroup(lat, lng);

	var groupId string;

	if (err == nil) {
		groupId = group.Id.Hex();
	}

	go SetUserLocation(token, lat, lng);

	var prevPosts []models.Post;
	var currentPosts []models.Post;

	if (len(groupId) > 0) {
		options := []bson.M{}
		options = append(options, bson.M{"groupId" : bson.ObjectIdHex(groupId)});
		if (len(result.GroupId) > 0 && result.GroupId.Hex() != groupId) {
			options = append(options, bson.M{"groupId" : result.GroupId});
		}

		err = c.Find(bson.M{"$or":options, "createdOn": bson.M{"$gt": lastUpdated}, "isActive" : true}).Limit(50).Sort("-score").All(&currentPosts);
		err = c.Find(bson.M{"$or":options, "createdOn": bson.M{"$lt": lastUpdated}, "isActive" : true}).Limit(50).Sort("-score").All(&prevPosts);

		for index, _ := range currentPosts {
			posts = append(posts, currentPosts[index]);
		}

		for index, _ := range prevPosts {
			posts = append(posts, prevPosts[index]);
		}

	} else {
		currentLatLng := [2]float64{lng, lat}

		options := []bson.M{}

		options = append(options, bson.M{"loc": bson.M{"$geoWithin": bson.M{"$centerSphere": []interface{}{currentLatLng, 25 / 3963.2}}}})

		if (len(result.GroupId) > 0) {
			options = append(options, bson.M{"groupId" : result.GroupId});
		}

		err = c.Find(bson.M{"$or":options, "createdOn": bson.M{"$gt": lastUpdated}, "isActive" : true}).Limit(50).Sort("-score").All(&currentPosts);
		err = c.Find(bson.M{"$or":options, "createdOn": bson.M{"$lt": lastUpdated}, "isActive" : true}).Limit(50).Sort("-score").All(&prevPosts);

		var count = 0;
		for index, _ := range currentPosts {
			isHisGroup := len(result.GroupId) > 0 && currentPosts[index].GroupId.Hex() == result.GroupId.Hex()
			if (!currentPosts[index].IsGroup || count < 4 || isHisGroup) {
				posts = append(posts, currentPosts[index]);
				if (currentPosts[index].IsGroup && !isHisGroup) {
					count++;
				}
			}
		}

		count = 0;
		for index, _ := range prevPosts {
			isHisGroup := len(result.GroupId) > 0 && prevPosts[index].GroupId.Hex() == result.GroupId.Hex()
			if (!prevPosts[index].IsGroup || count < 4 || isHisGroup) {
				posts = append(posts, prevPosts[index]);
				if (prevPosts[index].IsGroup && !isHisGroup) {
					count++;
				}
			}
		}
	}

	for index, _ := range posts {
		distance := common.DistanceLatLong(posts[index].Loc[1], lat, posts[index].Loc[0], lng)
		isHisGroup := len(result.GroupId) > 0 && posts[index].GroupId.Hex() == result.GroupId.Hex()
		if (isHisGroup) {
			posts[index].Score += 0.8;
		} else if (distance < 1000) {
			posts[index].Score += 0.8;
		} else if (distance < 1500) {
			posts[index].Score += 0.7;
		} else if (distance < 2000) {
			posts[index].Score += 0.6;
		} else if (distance < 2600) {
			posts[index].Score += 0.5;
		}

		if (len(posts[index].GroupName) > 0) {
			posts[index].PlaceName = posts[index].GroupName;
			posts[index].PlaceType = posts[index].GroupCategoryName;
		} else {
			posts[index].PlaceName = posts[index].Locality;
			posts[index].PlaceType = "location"
		}

		if (len(posts[index].PlaceName) > 24) {
			posts[index].PlaceName = posts[index].PlaceName[0:24] + "...";
		}

	}

	sort.Sort(models.ScoreSorter(posts))

	if (posts == nil) {
		posts = []models.Post{}
	}

	posts = addUserVotes(token, posts);

	return
}

func getPostVoteType(userId bson.ObjectId, postId string) string {

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	var userPosts models.UserPost
	err := c.Find(bson.M{"userId": userId}).One(&userPosts)

	if(err != nil) {
		return "none";
	}

	for i := 0; i < len(userPosts.PostVotes); i++ {
		if(userPosts.PostVotes[i].Id.Hex() == postId) {
			if (userPosts.PostVotes[i].IsUpvote) {
				return "upvote";
			} else {
				return "downvote";
			}
		}
	}
	return "none";
}

func getCommentVoteType(userId bson.ObjectId, postId string) string {

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	var userPosts models.UserPost
	err := c.Find(bson.M{"userId": userId}).One(&userPosts)

	if(err != nil) {
		return "none";
	}

	for i := 0; i < len(userPosts.CommentVotes); i++ {
		if(userPosts.PostVotes[i].Id.Hex() == postId) {
			if (userPosts.PostVotes[i].IsUpvote) {
				return "upvote";
			} else {
				return "downvote";
			}
		}
	}
	return "none";
}

func addUserVotes(token string, posts []models.Post) []models.Post {

	if (posts == nil) {
		return posts;
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	result, err := GetUserInfo(token);

	var userPosts models.UserPost
	err = c.Find(bson.M{"userId": result.Id}).One(&userPosts)
	if (err != nil) {
		err = models.FardoError{"Get User Posts: " + err.Error()}
		return posts;
	}

	m := make(map[string]string)

	for i := 0; i < len(userPosts.PostVotes); i++ {
		if (userPosts.PostVotes[i].IsUpvote) {
			m[userPosts.PostVotes[i].Id.Hex()] = "upvote";
		} else {
			m[userPosts.PostVotes[i].Id.Hex()] = "downvote";
		}
	}

	for i := 0; i < len(posts); i++ {
		voteType := m[posts[i].Id.Hex()];
		if (len(voteType) == 0) {
			posts[i].VoteClicked = "none";
		} else {
			posts[i].VoteClicked = voteType;
		}
	}

	return posts;
}

func addUserCommentVotes(token string, comments []models.Comment) []models.Comment {

	if (comments == nil) {
		return comments;
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_posts")

	result, err := GetUserInfo(token);
	var userPosts models.UserPost
	err = c.Find(bson.M{"userId": result.Id}).One(&userPosts)
	if (err != nil) {
		err = models.FardoError{"Get User Posts: " + err.Error()}
		return comments;
	}

	m := make(map[string]string)

	for i := 0; i < len(userPosts.CommentVotes); i++ {
		if (userPosts.CommentVotes[i].IsUpvote) {
			m[userPosts.CommentVotes[i].Id.Hex()] = "upvote";
		} else {
			m[userPosts.CommentVotes[i].Id.Hex()] = "downvote";
		}
	}

	for i := 0; i < len(comments); i++ {
		voteType := m[comments[i].Id.Hex()];
		if (len(voteType) == 0) {
			comments[i].VoteClicked = "none";
		} else {
			comments[i].VoteClicked = voteType;
		}
	}

	return comments;
}

func GetPopularPosts(token string, lat float64, lng float64) (posts []models.Post, err error) {

	nearByPosts, err := getNearByPopularPosts(lat, lng); //50%
	globalPosts, _ := getGlobalPopularPosts(); //30%
	adminAreaPosts, _ := getPopularPostsAdminArea(lat, lng); //20%


	for index, _ := range nearByPosts {
		if (len(nearByPosts[index].GroupName) > 0) {
			nearByPosts[index].PlaceName = nearByPosts[index].GroupName;
			nearByPosts[index].PlaceType = nearByPosts[index].GroupCategoryName;
		} else {
			nearByPosts[index].PlaceName = nearByPosts[index].Locality;
			nearByPosts[index].PlaceType = "location"
		}
		if (len(nearByPosts[index].PlaceName) > 24) {
			nearByPosts[index].PlaceName = nearByPosts[index].PlaceName[0:24] + "...";
		}
	}

	for index, _ := range globalPosts {
		if (len(globalPosts[index].GroupName) > 0) {
			globalPosts[index].PlaceName = globalPosts[index].GroupName;
			globalPosts[index].PlaceType = globalPosts[index].GroupCategoryName;
		} else {
			globalPosts[index].PlaceName = globalPosts[index].City;
			globalPosts[index].PlaceType = "location"
		}
		if (len(globalPosts[index].PlaceName) > 24) {
			globalPosts[index].PlaceName = globalPosts[index].PlaceName[0:24] + "...";
		}
	}

	for index, _ := range adminAreaPosts {
		if (len(adminAreaPosts[index].GroupName) > 0) {
			adminAreaPosts[index].PlaceName = adminAreaPosts[index].GroupName;
			adminAreaPosts[index].PlaceType = adminAreaPosts[index].GroupCategoryName;
		} else {
			adminAreaPosts[index].PlaceName = adminAreaPosts[index].City;
			adminAreaPosts[index].PlaceType = "location"
		}
		if (len(adminAreaPosts[index].PlaceName) > 24) {
			adminAreaPosts[index].PlaceName = adminAreaPosts[index].PlaceName[0:24] + "...";
		}
	}

	totalCount := len(nearByPosts) + len(globalPosts) + len(adminAreaPosts)

	j := 0
	k := 0
	l := 0

	nearPostsLen := len(nearByPosts);
	globalPostsLen := len(globalPosts);
	adminAreaPostsLen := len(adminAreaPosts);

	for i := 0; i < totalCount; i++ {
		if (i % 3 == 0) {
			if (j < nearPostsLen && !idInPosts(nearByPosts[j].Id.Hex(), posts)) {
				posts = append(posts, nearByPosts[j])
			}
			j++
		}
		if (i % 3 == 1) {
			if (k < globalPostsLen && !idInPosts(globalPosts[k].Id.Hex(), posts)) {
				posts = append(posts, globalPosts[k])
			}
			k++
		}
		if (i % 3 == 2) {
			if (l < adminAreaPostsLen && !idInPosts(adminAreaPosts[l].Id.Hex(), posts)) {
				posts = append(posts, adminAreaPosts[l])
			}
			l++
		}
	}

	for ; j < nearPostsLen; j++ {
		if (!idInPosts(nearByPosts[j].Id.Hex(), posts)) {
			posts = append(posts, nearByPosts[j])
		}
	}

	for ; k < globalPostsLen; k++ {
		if (!idInPosts(globalPosts[k].Id.Hex(), posts)) {
			posts = append(posts, globalPosts[k])
		}
	}

	for ; l < adminAreaPostsLen; l++ {
		if (!idInPosts(adminAreaPosts[l].Id.Hex(), posts)) {
			posts = append(posts, adminAreaPosts[l])
		}
	}

	if (posts == nil) {
		posts = []models.Post{}
	}

	posts = addUserVotes(token, posts);

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
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }},
		"isActive" : true, "upvotes": bson.M{"$gt": 2}}).Sort("-score").Limit(LocalPercent).All(&posts);
	if (posts == nil) {
		posts = []models.Post{}
	}
	return;
}

func getGlobalPopularPosts() (posts[]models.Post, err error) {

	context := common.NewContext()
	defer context.Close()

	c := context.DbCollection("posts")
	err = c.Find(bson.M{"isActive" : true, "upvotes": bson.M{"$gt": 2}}).Sort("-score").Limit(GlobalPercent).All(&posts);
	if (posts == nil) {
		posts = []models.Post{}
	}
	return;
}

func getPopularPostsAdminArea(lat float64, lng float64) (posts[]models.Post, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("posts")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 300 / 3963.2} }},
		"isActive" : true, "upvotes": bson.M{"$gt": 2}}).Sort("-score").Limit(AdminAreaPercent).All(&posts);
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

	err = postCol.Find(bson.M{"_id": bson.M{"$in": postIds}, "isActive" : true}).All(&posts)

	if (posts == nil) {
		posts = []models.Post{}
	}

	for index, _ := range posts {
		if (posts[index].IsGroup) {
			posts[index].PlaceName = posts[index].GroupName;
			posts[index].PlaceType = posts[index].GroupCategoryName;
		} else {
			posts[index].PlaceName = posts[index].Locality;
			posts[index].PlaceType = "location"
		}
	}
	//TODO: Optimize this to send user id if possible
	posts = addUserVotes(token, posts);

	return
}

func GetGroupPosts(token string, groupId string) (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"groupId": bson.ObjectIdHex(groupId), "isActive" : true}).All(&posts)
	if (posts == nil) {
		posts = []models.Post{}
	}

	for index, _ := range posts {
		posts[index].PlaceName = posts[index].GroupName;
		posts[index].PlaceType = posts[index].GroupCategoryName;
	}

	posts = addUserVotes(token, posts);

	return
}

func GetCurrentPosts() (posts []models.Post, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	err = c.Find(bson.M{"isActive": true}).All(&posts)

	if (posts == nil) {
		posts = []models.Post{}
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
	if (len(result.Phone) > 0) {
		comment.Jid = result.Phone + "@im.ripplin.in"
	}

	if (!comment.IsAnonymous && !result.IsAdmin) {
		comment.Username = result.Username;
	}

	err = c.Insert(&comment)

	if (err == nil) {
		go addToRecentUserPosts(result.Id, comment.PostId, "comment");
		post, err := findPostById(postId);
		if (err == nil ) {
			go updateReplyCount(postId, true);
			go common.SendCommentNotification(post.UserId.Hex(), post.Id.Hex(), comment.UserId.Hex(), comment.Id.Hex(), post.Content, comment.Content,comment.Username, comment.Jid, "comment");
		}
	}

	return comment.Id.Hex(), err
}

func updateReplyCount(id string, isIncrement bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	var step int
	if (isIncrement) {
		step = 1;
	} else {
		step = -1;
	}
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"replyCount": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	return
}

func UpvoteComment(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;
	voteStep := 1;
	if (undo) {
		step = -1;
		voteStep = -1;
	}

	voteType := getCommentVoteType(result.Id, id);

	downvoteStep := 0;
	if(voteType == "downvote") {
		downvoteStep = -1;
		voteStep = 2;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
			"downvotes": downvoteStep,
			"votes": voteStep,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	go checkCommentVoteCount(result.Id.Hex(), id, true);
	go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), true, undo, "comment");
	return
}

func DownvoteComment(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;
	voteStep := -1;
	if (undo) {
		step = -1;
		voteStep = 1;
	}

	voteType := getPostVoteType(result.Id, id);

	upvoteStep := 0;
	if(voteType == "upvote") {
		upvoteStep = -1;
		voteStep = -2;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": upvoteStep,
			"downvotes": step,
			"votes": voteStep,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	go checkCommentVoteCount(result.Id.Hex(), id, false);
	go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), false, undo, "comment");
	return
}

func checkCommentVoteCount(userId string, id string, isUpvote bool) (err error) {
	comment, err := findCommentById(id);
	votes := comment.Upvotes - comment.Downvotes;

	if (isUpvote) {
		if (comment.Upvotes == 2 || comment.Upvotes == 6 || comment.Upvotes == 9 || (comment.Upvotes > 15 && common.DivisbleByPowerOf2(comment.Upvotes))) {
			common.SendCommentUpvoteNotification(userId, comment.UserId.Hex(), comment.Id.Hex(), comment.PostId.Hex(), "comment_upvote", comment.Upvotes - comment.Downvotes, comment.Content);
		}
	}

	if (!isUpvote) {
		if (votes <= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendComment(id);
		}
	}
	return;
}

func checkVoteCount(token string, userId string, id string, isUpvote bool) (err error) {
	post, err := findPostById(id);
	votes := post.Upvotes - post.Downvotes;

	if (isUpvote) {
		if (post.Upvotes == 3 || post.Upvotes == 7 || post.Upvotes == 12 || (post.Upvotes > 15 && common.DivisbleByPowerOf2(post.Upvotes))) {
			posts := []models.Post{post}
			posts = addUserVotes(token, posts);
			var postType string;
			if (len(posts[0].ImageUrl) > 0) {
				postType = "image_post_upvote";
			} else {
				postType = "post_upvote";
			}
			common.SendUpvoteNotification(userId, posts[0].Id.Hex(), posts[0].UserId.Hex(), posts[0].Upvotes - posts[0].Downvotes, postType, posts[0].Content);
		}
	}

	if (!isUpvote) {
		if (votes <= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendPost(id, false);
		}
	}
	return;
}

func ReportSpam(id string, reason string) (err error) {
	spamReason := bson.M{
		"spamReasons": reason,
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
		go checkSpamCountLimit(id);
	}

	return;
}

func checkSpamCountLimit(id string) (err error) {
	post, err := findPostById(id);
	if (post.SpamCount >= models.SPAM_COUNT_LIMIT) {
		err = SuspendPost(id, false);
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

func GetAllComments(token string, postId string) (comments []models.Comment, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("comments")

	err = c.Find(bson.M{"postId": bson.ObjectIdHex(postId), "isActive": true}).All(&comments)

	if (comments == nil) {
		comments = []models.Comment{}
	}

	comments = addUserCommentVotes(token, comments);
	return
}

