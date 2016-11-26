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

	news.UserId = result.Id;
	news.Username = result.Username;
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
		fileName := "news_" + news.Id.Hex();
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
			return news, models.FardoError{"Insert News Image Error: " + err.Error()}
		}

		news.ImageUrl = res;

	}

	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	news.IsActive = true;
	news.CreatedOn = time.Now().UTC();
	news.Score = redditNewsRankingAlgorithmNews(news);

	err = c.Insert(&news)

	if (err != nil) {
		return news, models.FardoError{"Insert News Error: " + err.Error()}
	}

	go addToRecentUserNews(result.Id, news.Id, "news");

	go CalculateUserScoreForNews(news, ActionCreate);

	return news, err
}

func redditNewsRankingAlgorithmNews(news models.News) float64 {
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

func GetNews(token string, lat float64, lng float64) (news []models.News, err error) {

	nearByNews, err := getNearByNews(lat, lng); //50%
	globalNews, _ := getGlobalNews(); //30%
	adminAreaNews, _ := getPopularNewsAdminArea(lat, lng); //20%


	for index, _ := range nearByNews {
		nearByNews[index].PlaceName = nearByNews[index].Locality;
		nearByNews[index].PlaceType = "location"
		if (len(nearByNews[index].PlaceName) > 24) {
			nearByNews[index].PlaceName = nearByNews[index].PlaceName[0:24] + "...";
		}
	}

	for index, _ := range globalNews {
		globalNews[index].PlaceName = globalNews[index].City;
		globalNews[index].PlaceType = "location"
		if (len(globalNews[index].PlaceName) > 24) {
			globalNews[index].PlaceName = globalNews[index].PlaceName[0:24] + "...";
		}
	}

	for index, _ := range adminAreaNews {
		adminAreaNews[index].PlaceName = adminAreaNews[index].City;
		adminAreaNews[index].PlaceType = "location"
		if (len(adminAreaNews[index].PlaceName) > 24) {
			adminAreaNews[index].PlaceName = adminAreaNews[index].PlaceName[0:24] + "...";
		}
	}

	totalCount := len(nearByNews) + len(globalNews) + len(adminAreaNews)

	j := 0
	k := 0
	l := 0

	nearNewsLen := len(nearByNews);
	globalNewsLen := len(globalNews);
	adminAreaNewsLen := len(adminAreaNews);

	for i := 0; i < totalCount; i++ {
		if (i % 3 == 0) {
			if (j < nearNewsLen && !idInNews(nearByNews[j].Id.Hex(), news)) {
				news = append(news, nearByNews[j])
			}
			j++
		}
		if (i % 3 == 1) {
			if (k < globalNewsLen && !idInNews(globalNews[k].Id.Hex(), news)) {
				news = append(news, globalNews[k])
			}
			k++
		}
		if (i % 3 == 2) {
			if (l < adminAreaNewsLen && !idInNews(adminAreaNews[l].Id.Hex(), news)) {
				news = append(news, adminAreaNews[l])
			}
			l++
		}
	}

	for ; j < nearNewsLen; j++ {
		if (!idInNews(nearByNews[j].Id.Hex(), news)) {
			news = append(news, nearByNews[j])
		}
	}

	for ; k < globalNewsLen; k++ {
		if (!idInNews(globalNews[k].Id.Hex(), news)) {
			news = append(news, globalNews[k])
		}
	}

	for ; l < adminAreaNewsLen; l++ {
		if (!idInNews(adminAreaNews[l].Id.Hex(), news)) {
			news = append(news, adminAreaNews[l])
		}
	}

	if (news == nil) {
		news = []models.News{}
	}
	return
}

func idInNews(id string, list []models.News) bool {
	for _, b := range list {
		if b.Id.Hex() == id {
			return true
		}
	}
	return false
}

func getNearByNews(lat float64, lng float64) (news[]models.News, err error) {

	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("news")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 30 / 3963.2} }},
		"isActive" : true}).Sort("-score").Limit(LocalPercent).All(&news);
	if (news == nil) {
		news = []models.News{}
	}
	return;
}

func getGlobalNews() (news[]models.News, err error) {

	context := common.NewContext()
	defer context.Close()

	c := context.DbCollection("news")
	err = c.Find(bson.M{"isActive" : true, "upvotes": bson.M{"$gt": 2}}).Sort("-score").Limit(GlobalPercent).All(&news);
	if (news == nil) {
		news = []models.News{}
	}
	return;
}

func getPopularNewsAdminArea(lat float64, lng float64) (news []models.News, err error) {
	context := common.NewContext()
	defer context.Close()

	currentLatLng := [2]float64{lng, lat}
	c := context.DbCollection("news")
	err = c.Find(bson.M{"loc":
	bson.M{"$geoWithin":
	bson.M{"$centerSphere": []interface{}{currentLatLng, 200 / 3963.2} }},
		"isActive" : true, "upvotes": bson.M{"$gt": 1}}).Sort("-score").Limit(AdminAreaPercent).All(&news);
	if (news == nil) {
		news = []models.News{

		}
	}
	return;
}

func checkVoteCountNews(token string, userId string, id string, isUpvote bool) (err error) {
	news, err := GetNewsById(id);
	votes := news.Upvotes - news.Downvotes;

	if (isUpvote) {
		if (news.Upvotes == 2 || news.Upvotes == 6 || news.Upvotes == 11 || (news.Upvotes > 14 && common.DivisbleByPowerOf2(news.Upvotes))) {
			news := []models.News{news}
			news = addUserNewsVotes(token, news);
			var postType string;
			if (len(news[0].ImageUrl) > 0) {
				postType = "image_news_upvote";
			} else {
				postType = "news_upvote";
			}
			common.SendUpvoteNotification(userId, news[0].Id.Hex(), news[0].UserId.Hex(), news[0].Upvotes - news[0].Downvotes, postType, news[0].Content);
		}
	}

	if (!isUpvote) {
		if (votes <= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendNews(id, false);
		}
	}
	return;
}

func updateUserAndNewsScore(id string, actionType ActionType) {
	news, err := GetNewsById(id);
	updateNewsScore(id, redditNewsRankingAlgorithmNews(news));

	if (err == nil) {
		CalculateUserScoreForNews(news, actionType);
	}
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

func UpvoteNews(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	result, err := GetUserInfo(token)

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
		go addToRecentUserNewsVotes(result.Id, bson.ObjectIdHex(id), true, undo, "news");
	}
	return
}

func DownvoteNews(token string, id string, undo bool) (err error) {
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
			"downvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})
	if (err == nil) {
		go updateUserAndNewsScore(id, ActionDownvote);
		go checkVoteCountNews(result.Token, result.Id.Hex(), id, false);
		go addToRecentUserNewsVotes(result.Id, bson.ObjectIdHex(id), false, undo, "news");
	}
	return
}

func SuspendNews(id string, isSilent bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})

	if (err == nil && !isSilent) {
		news, _ := GetNewsById(id)
		common.SendDeletePostNotification(news.UserId.Hex(), news.Content);
	}
	return
}

func SuspendNewsComment(id string) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	if (err != nil) {
		return
	}

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"isActive": false,
			"modifiedOn": time.Now().UTC(),
		}})

	if (err == nil ) {
		comment, errr := findNewsCommentById(id);
		if (errr != nil) {
			return
		}
		go updateReplyCount(comment.NewsId.Hex(), false);
	}

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

	var userNews models.UserNews
	err = c.Find(bson.M{"userId": result.Id}).One(&userNews)
	if (err != nil) {
		err = models.FardoError{"Get User News: " + err.Error()}
		return news;
	}

	m := make(map[string]string)

	for i := 0; i < len(userNews.Votes); i++ {
		if (userNews.Votes[i].IsUpvote) {
			m[userNews.Votes[i].Id.Hex()] = "upvote";
		} else {
			m[userNews.Votes[i].Id.Hex()] = "downvote";
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

func UpvoteNewsComment(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	step := 1;
	if (undo) {
		step = -1;
	}

	result, err := GetUserInfo(token)

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"upvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	go checkNewsCommentVoteCount(result.Id.Hex(), id, true);
	go addToRecentUserNewsVotes(result.Id, bson.ObjectIdHex(id), true, undo, "comment");
	return
}

func DownvoteNewsComment(token string, id string, undo bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	step := 1;
	if (undo) {
		step = -1;
	}

	result, err := GetUserInfo(token)

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"downvotes": step,
		}, "$set": bson.M{
			"modifiedOn": time.Now().UTC()}})

	go checkNewsCommentVoteCount(result.Id.Hex(), id, false);
	go addToRecentUserNewsVotes(result.Id, bson.ObjectIdHex(id), false, undo, "comment");
	return
}

func findNewsCommentById(id string) (comment models.NewsComment, err error, ) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	err = c.FindId(bson.ObjectIdHex(id)).One(&comment);
	return
}

func addToRecentUserNewsVotes(userId bson.ObjectId, id bson.ObjectId, isUpvote bool, isUndo bool, fieldType string) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_news")

	ids := bson.M{};
	if (fieldType == "comment") {
		ids = bson.M{
			"$pull": bson.M{"commentVotes" : bson.M{"_id" : id}}}
	} else {
		ids = bson.M{
			"$pull": bson.M{"votes" : bson.M{"_id" : id}}}
	}
	_ = c.Update(bson.M{"userId": userId},
		ids)

	if (isUndo) {
		return;
	}

	var userVote models.UserVote;
	userVote.Id = id;
	userVote.IsUpvote = isUpvote;

	ids = bson.M{
		"votes": userVote,
	}

	if (fieldType == "comment") {
		ids = bson.M{
			"commentVotes": userVote,
		}
	}

	_, _ = c.Upsert(bson.M{"userId": userId},
		bson.M{"$push": ids})
}

func checkNewsCommentVoteCount(userId string, id string, isUpvote bool) (err error) {
	comment, err := findNewsCommentById(id);
	votes := comment.Upvotes - comment.Downvotes;

	if (isUpvote) {
		if (comment.Upvotes == 2 || comment.Upvotes == 6 || comment.Upvotes == 9 || (comment.Upvotes > 15 && common.DivisbleByPowerOf2(comment.Upvotes))) {
			common.SendCommentUpvoteNotification(userId, comment.UserId.Hex(), comment.Id.Hex(), comment.NewsId.Hex(), "news_comment_upvote", comment.Upvotes - comment.Downvotes, comment.Content);
		}
	}

	if (!isUpvote) {
		if (votes <= models.NEGATIVE_VOTES_LIMIT) {
			err = SuspendNewsComment(id);
		}
	}
	return;
}

func AddNewsComment(token string, newsId string, comment models.NewsComment) (string, error) {
	var err error
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	result, err := GetUserInfo(token)

	if (err != nil ) {
		return "", err
	}

	comment.Id = bson.NewObjectId()
	comment.IsActive = true;
	comment.NewsId = bson.ObjectIdHex(newsId);
	comment.CreatedOn = time.Now()
	comment.UserId = result.Id;

	err = c.Insert(&comment)

	if (err == nil) {
		go addToRecentUserNews(result.Id, comment.NewsId, "comment");
		news, err := GetNewsById(newsId);
		if (err == nil ) {
			go updateReplyCountNews(newsId, true);
			go common.SendCommentNotification(news.UserId.Hex(), news.Id.Hex(), comment.UserId.Hex(), comment.Id.Hex(), news.Content, comment.Content, "news_comment");
		}
	}

	return comment.Id.Hex(), err
}

func addToRecentUserNews(userId bson.ObjectId, newsId bson.ObjectId, fieldType string) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("user_news")

	ids := bson.M{
		"newsIds": newsId,
	}

	if (fieldType == "comment") {
		ids = bson.M{
			"commentNewsIds": newsId,
		}
	}
	_, _ = c.Upsert(bson.M{"userId": userId},
		bson.M{"$push": ids})
}

func updateReplyCountNews(id string, isIncrement bool) (err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")

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

func GetAllNewsComments(token string, newsId string) (comments []models.NewsComment, err error) {
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news_comments")

	err = c.Find(bson.M{"newsId": bson.ObjectIdHex(newsId), "isActive": true}).All(&comments)

	if (comments == nil) {
		comments = []models.NewsComment{}
	}

	comments = addUserNewsCommentVotes(token, comments);
	return
}

func ReportNewsSpam(id string, reason string) (err error) {
	spamReason := bson.M{
		"spamReasons": reason,
	}
	context := common.NewContext()
	defer context.Close()
	c := context.DbCollection("news")
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"spamCount": 1},
			"$push": spamReason, })

	return;
}
