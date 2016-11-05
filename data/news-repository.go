package data

import (
	"github.com/mymachine8/fardo-api/models"
	"github.com/mymachine8/fardo-api/common"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"encoding/base64"
	"time"
	"math"
)

func CreateNews(token string, news models.News, isAdmin bool) (models.News, error) {
	var err error

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)
	if (err != nil) {
		return news, models.FardoError{"Get Access Token: " + err.Error()}
	}
	//TODO: Have to revisit this code
	news.UserId = result.Id;
	if (len(news.LabelId) > 0) {
		labelContext := common.NewContext()
		labelCol := labelContext.DbCollection("labels")
		var label models.Label
		err = labelCol.FindId(news.LabelId).One(&label)
		labelContext.Close()
		if (err == nil) {
			news.LabelName = label.Name;
		}
	}

	news.Id = bson.NewObjectId();


	if (len(news.ImageData) > 0 ) {
		fileName := "post_" + news.Id.Hex();
		imageReader := strings.NewReader(news.ImageData);

		dec := base64.NewDecoder(base64.StdEncoding, imageReader);

		var fileType string
		if (len(news.ImageType) > 0) {
			fileType = news.ImageType
		} else {
			fileType = "jpeg"
		}

		res, err := common.SendItemToCloudStorage(common.PostImage, fileName, fileType, dec);

		if (err != nil) {
			return news, models.FardoError{"Insert Post Image Error: " + err.Error()}
		}

		news.ImageUrl = res;

	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	news.IsActive = true;
	news.CreatedOn = time.Now().UTC();
	news.Score = redditPostRankingAlgorithmNews(news);

	err = c.Insert(&news)

	if (err != nil) {
		return news, models.FardoError{"Insert News Error: " + err.Error()}
	}

	go addToRecentUserPosts(result.Id, news.Id, "news");

	go CalculateUserScoreForNews(news, ActionCreate);

	return news, err
}

func redditPostRankingAlgorithmNews(news models.News) float64 {
	timeDiff := common.GetTimeSeconds(news.CreatedOn) - common.GetZingCreationTimeSeconds();
	votes := int64(news.Upvotes - news.Downvotes);
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

func GetNewsById(id string) (news models.News, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	err = c.FindId(bson.ObjectIdHex(id)).One(&news)
	return
}

func checkVoteCountNews(token string, userId string, id string, isUpvote bool) (err error) {
	news, err := GetNewsById(id);
	votes := news.Upvotes - news.Downvotes;

	if (isUpvote) {
		if (news.Upvotes == 2 || news.Upvotes == 6 || news.Upvotes == 11 || (news.Upvotes > 14 && common.DivisbleByPowerOf2(news.Upvotes))) {
			news := []models.News{news}
			news = addUserNewsVotes(token, news);
			var postType string;
			if(len(news[0].ImageUrl) > 0) {
				postType = "image_news_upvote";
			} else {
				postType = "news_upvote";
			}
			common.SendUpvoteNotification(userId, news[0].Id.Hex(),news[0].UserId.Hex(), news[0].Upvotes - news[0].Downvotes,postType, news[0].Content);
		}
	}

	if (!isUpvote) {
		if (votes <= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendPost(id, false);
		}
	}
	return;
}

func updateUserAndNewsScore(id string, actionType ActionType) {
	news, err := GetNewsById(id);
	updatePostScore(id, redditPostRankingAlgorithmNews(news));

	if (err == nil) {
		CalculateUserScoreForNews(news, actionType);
	}
}

func UpvoteNews(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;
	if (undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndNewsScore(id, ActionUpvote);
		go checkVoteCountNews(result.Token, result.Id.Hex(), id, true);
		go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), true, undo, "post");
	}
	return
}

func DownvoteNews(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("posts")

	tokenContext := common.NewContext()
	defer tokenContext.Close()
	tokenCol := tokenContext.DbCollection("users")
	var result models.User
	err = tokenCol.Find(bson.M{"token": token}).One(&result)

	step := 1;

	if (undo) {
		step = -1;
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndPostScore(id, ActionDownvote);
		go checkVoteCount(result.Token, result.Id.Hex(), id, false);
		go addToRecentUserVotes(result.Id, bson.ObjectIdHex(id), false, undo, "post");
	}
	return
}

func SuspendNews(id string, isSilent bool) (err error) {
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
		common.SendDeletePostNotification(post);
	}
	return
}

func SuspendNewsComment(id string) (err error) {
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

func updateNewsScore(id string, score float64) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"score": score,
		}})
	return
}

func addUserNewsVotes(token string, news []models.News) []models.News {

	if (news == nil) {
		return news;
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_news")

	result, err := GetUserInfo(token);

	var userPosts models.UserNews
	err = c.Find(bson.M{"userId": result.Id}).One(&userPosts)
	if (err != nil) {
		err = models.FardoError{"Get User News: " + err.Error()}
		return news;
	}

	m := make(map[string]string)

	for i := 0; i < len(userPosts.PostVotes); i++ {
		if (userPosts.PostVotes[i].IsUpvote) {
			m[userPosts.PostVotes[i].Id.Hex()] = "upvote";
		} else {
			m[userPosts.PostVotes[i].Id.Hex()] = "downvote";
		}
	}

	for i := 0; i < len(news); i++ {
		voteType := m[news[i].Id.Hex()];
		if (len(voteType) == 0) {
			news[i].VoteClicked = "none";
		} else {
			news[i].VoteClicked = voteType;
		}
	}

	return news;
}

func addUserNewsCommentVotes(token string, comments []models.NewsComment) []models.NewsComment {

	if (comments == nil) {
		return comments;
	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_news")

	result, err := GetUserInfo(token);
	var userNews models.UserNews
	err = c.Find(bson.M{"userId": result.Id}).One(&userNews)
	if (err != nil) {
		err = models.FardoError{"Get User News: " + err.Error()}
		return comments;
	}

	m := make(map[string]string)

	for i := 0; i < len(userNews.CommentVotes); i++ {
		if (userNews.CommentVotes[i].IsUpvote) {
			m[userNews.CommentVotes[i].Id.Hex()] = "upvote";
		} else {
			m[userNews.CommentVotes[i].Id.Hex()] = "downvote";
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
