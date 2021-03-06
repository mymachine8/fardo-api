package controllers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/mymachine8/fardo-api/models"
	"encoding/json"
	"github.com/mymachine8/fardo-api/data"
	"github.com/mymachine8/fardo-api/common"
	"github.com/mymachine8/fardo-api/slack"
	"log"
	"github.com/mymachine8/fardo-api/cors"
	"strconv"
	"time"
	"errors"
)

func InitRoutes() http.Handler {
	r := httprouter.New();

	r.GET("/", helloWorldHandler);

	http.Handle("/", r);


	//---------------  Main Endpoints -------------------------------


	r.GET("/api/near-groups", GetNearByGroupsHandler);
	r.GET("/api/user-near-groups", GetNearCollegesAndOfficesHandler);
	r.GET("/api/popular-groups", GetPopularGroupsHandler);
	r.GET("/api/my-circle", myCircleHandler);
	r.GET("/api/home", homeHandler);
	r.GET("/api/popular", popularPostsHandler);

	r.POST("/api/users", memberRegisterHandler);
	r.GET("/api/users", getUserInfoHandler);
	r.PUT("/api/users/group", updateUserGroupHandler);
	r.PUT("/api/users/remove-group", removeUserGroupHandler);
	r.PUT("/api/users/username", updateUsernameHandler);
	r.GET("/api/users/username-availability", checkUsernameAvailabilityHandler);
	r.PUT("/api/users/phone", updateUserPhoneHandler);
	r.GET("/api/users/username", getUsernameHandler);
	r.PUT("/api/users/home", updateUserHomeLocationHandler);
	r.PUT("/api/users/remove-home", removeHomeLocationHandler);
	r.GET("/api/users/score", getUserScoreHandler);
	r.PUT("/api/users/lock-group", lockUserGroupHandler);
	r.PUT("/api/users/unlock-group", unlockUserGroupHandler);
	r.PUT("/api/users/fcm-token", updateUserFcmTokenHandler);
	r.PUT("/api/users/location", updateUserLocationTokenHandler);
	r.GET("/api/my-recent-posts", recentUserPostsHandler);
	r.GET("/api/my-recent-comments", recentUserCommentedPostsHandler);

	r.GET("/api/posts/:id", getPostByIdHandler);
	r.PUT("/api/posts/:id/upvote", upvotePostHandler);
	r.PUT("/api/posts/:id/downvote", downvotePostHandler);
	r.PUT("/api/posts/:id/undo-upvote", undoUpvotePostHandler);
	r.PUT("/api/posts/:id/undo-downvote", undoDownvotePostHandler);
	r.PUT("/api/posts/:id/suspend", suspendPostHandler);
	r.PUT("/api/posts/:id/silent-suspend", silentSuspendHandler);
	r.PUT("/api/comments/:id/suspend", suspendCommentHandler);
	r.GET("/api/posts/:id/comments", commentListHandler);
	r.POST("/api/posts/:id/comments", createCommentHandler);
	r.PUT("/api/comments/:id/upvote", upvoteCommentHandler);
	r.PUT("/api/comments/:id/downvote", downvoteCommentHandler);
	r.PUT("/api/comments/:id/undo-upvote", undoUpvoteCommentHandler);
	r.PUT("/api/comments/:id/undo-downvote", undoDownvoteCommentHandler);
	r.POST("/api/posts", createPostHandler);
	r.POST("/api/feedback", submitFeedbackHandler);
	r.PUT("/api/report-spam", reportSpamHandler);

	r.POST("/api/news", createNewsHandler);
	r.GET("/api/news", getNewsHandler);
	r.GET("/api/news/:id", getNewsByIdHandler);
	r.GET("/api/news/:id/comments", newsCommentListHandler);
	r.POST("/api/news/:id/comments", createNewsCommentHandler);
	r.PUT("/api/news/:id/upvote", upvoteNewsHandler);
	r.PUT("/api/news/:id/downvote", downvoteNewsHandler);
	r.PUT("/api/news/:id/undo-upvote", undoUpvoteNewsHandler);
	r.PUT("/api/news/:id/undo-downvote", undoDownvoteNewsHandler);
	r.PUT("/api/news/:id/silent-suspend", suspendNewsHandler);
	r.PUT("/api/news-comments/:id/upvote", upvoteNewsCommentHandler);
	r.PUT("/api/news-comments/:id/downvote", downvoteNewsCommentHandler);
	r.PUT("/api/news-comments/:id/undo-upvote", undoUpvoteNewsCommentHandler);
	r.PUT("/api/news-comments/:id/undo-downvote", undoDownvoteNewsCommentHandler);
	r.PUT("/api/news-comments/:id/suspend", suspendNewsCommentHandler);
	r.PUT("/api/news-report-spam", reportNewsSpamHandler);

	//----------------  End of main endpoints -----------------------


	r.GET("/api/categories", categoryListHandler);
	r.GET("/api/categories/:id/sub-categories", subCategoryListHandler);
	r.POST("/api/sub-categories/bulk-insert", bulkInsertSubCategoryHandler);
	r.GET("/api/groups", groupListHandler);
	r.GET("/api/groups-search", getAllGroupsHandler);
	r.GET("/api/groups/:id", getGroupByIdHandler);
	r.POST("/api/groups", createGroupHandler);
	r.PUT("/api/groups/:id", updateGroupHandler);
	r.PUT("/api/suspend-groups/:id", suspendGroupHandler);
	r.PUT("/api/upload-group-icon/:id", uploadGroupIcon);
	r.PUT("/api/upload-group-image/:id", uploadGroupImage);
	r.DELETE("/api/groups/:id", removeGroupHandler);


	r.GET("/api/groups/:id/labels", groupLabelListHandler);
	r.GET("/api/labels", labelListHandler);
	r.GET("/api/user-labels", userLabelListHandler);
	r.GET("/api/labels/:id", getLabelByIdHandler);
	r.POST("/api/groups/:id/labels", createLabelHandler);
	r.PUT("/api/labels/:id", updateLabelHandler);
	r.PUT("/api/labels/:id/activate", updateLabelHandler);
	r.PUT("/api/labels/:id/suspend", updateLabelHandler);
	r.DELETE("/api/labels/:id", removeLabelHandler);
	r.POST("/api/labels/bulk", createLabelsBulkHandler);


	r.POST("/api/admin/register", registerAdminHandler);
	r.GET("/api/admin/users", getActiveUsersList);
	r.POST("/api/admin/login", loginAdminHandler);

	r.GET("/api/admin/posts", allPostsListHandler);
	r.GET("/api/admin/solr-collection", solrCollectionHandler);
	r.GET("/api/admin/posts/current", currentPostsListHandler);
	r.POST("/api/admin/posts", createAdminPostHandler);
	r.GET("/api/label-posts/:id", labelPostsListHandler);
	r.GET("/api/group-posts/:id", groupPostsGroupHandler);
	r.GET("/api/compute-score", groupTrendingScore);


	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://fardo.krishnakommanapalli.in", "http://localhost:9003"},
		AllowCredentials: true,
		Debug: true,
		AllowedMethods : []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders :[]string{"Origin", "Accept", "Content-Type", "Authorization"},
	})
	handler := c.Handler(r)

	return handler;
}

func helloWorldHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rw.Write(common.SuccessResponseJSON("DONE"));
}

func groupTrendingScore(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := data.CalculatePlacesTrendingScore();

	if (err != nil) {
		log.Print(err.Error())
	}

	rw.Write(common.SuccessResponseJSON("DONE"));
}

func myCircleHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	layout := "2006-01-02T15:04:05.000Z"
	last_updated, _ := time.Parse(
		layout,
		r.URL.Query().Get("last_updated"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	token := common.GetAccessToken(r);

	result, _, e := data.GetMyCirclePosts(token, lat, lng, last_updated);

	if (e != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, e);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));

}

func homeHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	layout := "2006-01-02T15:04:05.000Z"
	last_updated, _ := time.Parse(
		layout,
		r.URL.Query().Get("last_updated"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	token := common.GetAccessToken(r);

	result, group, e := data.GetMyCirclePosts(token, lat, lng, last_updated);
	response := struct {
		Posts []models.Post `json:"posts"`
		Group *models.Group `json:"group,omitempty"`
	}{
		result,
		group,
	}

	if (e != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, e);
		return
	}

	rw.Write(common.SuccessResponseJSON(response));

}

func popularPostsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	result, err := data.GetPopularPosts(token, lat, lng);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func GetNearByGroupsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error;
	var lat, lng float64;
	var limit int;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	result, err := data.GetNearByGroups(lat, lng, limit);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func GetNearCollegesAndOfficesHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	result, err := data.GetNearCollegesAndOffices(lat, lng);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func GetPopularGroupsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	result, err := data.GetPopularGroups(lat, lng);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func labelPostsListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetLabelPosts(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func recentUserCommentedPostsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);

	result, err := data.GetRecentUserPosts(token, "comment");
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func recentUserPostsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);

	result, err := data.GetRecentUserPosts(token, "post");
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func groupPostsGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);

	result, err := data.GetGroupPosts(token, p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func solrCollectionHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var groups []models.Group
	var err error
	groups, err = data.GetAllGroups("");

	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}

	var subCategories []models.GroupSubCategory
	subCategories, err = data.GetAllSubCategories();

	var results []models.SolrSchema

	groupsLen := len(groups);
	subCategoriesLen := len(subCategories);

	i := 0

	for i = 0; i < groupsLen; i++ {
		result := models.SolrSchema{Id: groups[i].Id, Name: groups[i].Name, ShortName:groups[i].ShortName, Type: "group"};
		results = append(results, result)
	}

	for i = 0; i < subCategoriesLen; i++ {
		result := models.SolrSchema{Id: subCategories[i].Id, Name: subCategories[i].Name, Type: "subcategory"};
		results = append(results, result)
		i++;
	}

	rw.Header().Set("Content-Disposition", "attachment; filename=solr_collection.json")
	rw.Header().Set("Content-Type", "application/json")
	jsonResult, err := json.Marshal(results);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(jsonResult);
}

func allPostsListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var queryParams models.Post;

	queryParams.GroupName = r.URL.Query().Get("groupname_like");
	queryParams.City = r.URL.Query().Get("city_like");
	queryParams.State = r.URL.Query().Get("state_like");

	var page int64
	page, _ = strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	result, err := data.GetAllPosts(int(page), queryParams);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func currentPostsListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetCurrentPosts();
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func createAdminPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)

	if (err != nil) {
		writeErrorResponse(rw, r, p, post, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	id, err := data.CreatePostAdmin(token, post);

	if (err != nil) {
		writeErrorResponse(rw, r, p, post, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)

	if (err != nil) {
		writeErrorResponse(rw, r, p, post, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	result, err := data.CreatePostUser(token, post);

	if (err != nil) {
		post.ImageData = "";
		writeErrorResponse(rw, r, p, post, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func createNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var news models.News
	err := json.NewDecoder(r.Body).Decode(&news)

	if (err != nil) {
		writeErrorResponse(rw, r, p, news, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	result, err := data.CreateNews(token, news, false);

	if (err != nil) {
		news.ImageData = "";
		writeErrorResponse(rw, r, p, news, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error;
	var lat, lng float64;
	lat, err = strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err = strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	result, err := data.GetNews(token, lat, lng);
	if (err != nil) {
		writeErrorResponse(rw, r, p, []byte{}, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func createCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var comment models.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)

	if (err != nil) {
		writeErrorResponse(rw, r, p, comment, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	id, err := data.AddComment(token, p.ByName("id"), comment);

	if (err != nil) {
		writeErrorResponse(rw, r, p, comment, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var comment models.NewsComment
	err := json.NewDecoder(r.Body).Decode(&comment)

	if (err != nil) {
		writeErrorResponse(rw, r, p, comment, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	id, err := data.AddNewsComment(token, p.ByName("id"), comment);

	if (err != nil) {
		writeErrorResponse(rw, r, p, comment, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func upvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvotePost(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoUpvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvotePost(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvotePost(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoDownvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvotePost(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func upvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteComment(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoUpvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteComment(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteComment(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoDownvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteComment(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendPost(p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func silentSuspendHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendPost(p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendComment(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendNewsComment(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func upvoteNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteNews(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoUpvoteNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteNews(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvoteNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteNews(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoDownvoteNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteNews(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func upvoteNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteNewsComment(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoUpvoteNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.UpvoteNewsComment(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvoteNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteNewsComment(token, p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func undoDownvoteNewsCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.DownvoteNewsComment(token, p.ByName("id"), true);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendNewsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendNews(p.ByName("id"), false);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func updateUserGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		GroupId string `json:"groupId"`
		Lat     float64 `json:"lat,omitempty"`
		Lng     float64 `json:"lng,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	token := common.GetAccessToken(r);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	isGroupLocked,group, er := data.UpdateUserGroup(token, body.GroupId, body.Lat, body.Lng);

	if (er != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, er);
		return
	}

	response := struct {
		Group       models.Group `json:"group"`
		IsGroupLocked bool `json:"isGroupLocked"`
	}{
		group,
		isGroupLocked,
	}

	rw.Write(common.SuccessResponseJSON(response));
}

func removeUserGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);

	err := data.RemoveUserGroup(token);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func removeHomeLocationHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);

	err := data.RemoveHomeLocation(token);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func lockUserGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.LockUserGroup(token, true);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func unlockUserGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	err := data.LockUserGroup(token, false);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("SUCCESS"));
}

func updateUserFcmTokenHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		FcmToken string `json:"fcmToken"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	token := common.GetAccessToken(r);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	err = data.SetUserFcmToken(token, body.FcmToken);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("SUCCESS"));
}

func updateUsernameHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	token := common.GetAccessToken(r);

	var result string;
	result, err = data.SetUsernameToken(token, body.Username);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, err);
		return
	}

	if (result == "success") {
		rw.Write(common.SuccessResponseJSON("success"));
	} else {
		writeErrorResponse(rw, r, p, body, http.StatusPreconditionFailed, errors.New(result));
	}
}

func checkUsernameAvailabilityHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	isAvailable, err := data.CheckUsernameAvailability(r.URL.Query().Get("username"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, r.URL.Query().Get("username"), http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(isAvailable));
}

func updateUserPhoneHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		Imei        string `json:"imei"`
		SessionId   uint64 `json:"sessionId"`
		Token       string `json:"token"`
		TokenSecret string `json:"tokenSecret"`
		Phone       string `json:"phone"`
		FcmToken    string `json:"fcmToken"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	token := common.GetAccessToken(r);


	user, group, errr := data.ChangeUserPhone(token, body.Imei, body.SessionId, body.Token, body.TokenSecret, body.Phone, body.FcmToken);

	if (errr != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, errr);
		return
	}

	response := struct {
		User  models.User `json:"user"`
		Group models.Group `json:"group,omitempty"`
	}{
		user,
		group,
	}

	rw.Write(common.SuccessResponseJSON(response));
}

func getUsernameHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	phone := r.URL.Query().Get("phone");

	username, err := data.GetUsernameByPhone(phone);

	if (err != nil) {
		writeErrorResponse(rw, r, p, username, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(username));
}

func updateUserHomeLocationHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		HomeAddress string `json:"homeAddress"`
		Lat         float64 `json:"lat,omitempty"`
		Lng         float64 `json:"lng,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	token := common.GetAccessToken(r);


	err = data.ChangeUserHomeLocation(token, body.HomeAddress, body.Lat, body.Lng);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("SUCCESS"));
}

func getUserScoreHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);


	score, err := data.GetUserScore(token);

	if (err != nil) {
		writeErrorResponse(rw, r, p, score, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(score));
}

func updateUserLocationTokenHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct {
		lat float64 `json:"lat"`
		lng float64 `json:"lng"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)

	token := common.GetAccessToken(r);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusBadRequest, err);
		return
	}

	err = data.SetUserLocation(token, body.lat, body.lng);

	if (err != nil) {
		writeErrorResponse(rw, r, p, body, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("SUCCESS"));
}

func createGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	id, err := data.CreateGroup(group);

	if (err != nil) {
		group.ImageData = "";
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var label models.Label
	err := json.NewDecoder(r.Body).Decode(&label)

	if (err != nil) {
		writeErrorResponse(rw, r, p, label, http.StatusInternalServerError, err);
		return
	}

	id, err := data.CreateLabel(p.ByName("id"), label);

	if (err != nil) {
		writeErrorResponse(rw, r, p, label, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func removeLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := data.RemoveLabel(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func createLabelsBulkHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//Defered for next release
}

func categoryListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetAllCategories();
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));

}

func subCategoryListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetSubCategories(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));

}

func groupListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var err error
	var page int64
	var result []models.Group

	var queryParams models.Group;

	queryParams.Name = r.URL.Query().Get("name_like");
	queryParams.City = r.URL.Query().Get("city_like");
	queryParams.State = r.URL.Query().Get("state_like");

	page, err = strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)

	searchStr := r.URL.Query().Get("name");

	if (len(searchStr) > 0) {
		result, err = data.GetAllGroups(searchStr);
	} else {
		result, err = data.GetGroups(int(page), queryParams);
	}

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func getAllGroupsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	searchStr := r.URL.Query().Get("name");
	var err error
	var result []models.Group
	result, err = data.GetAllGroups(searchStr);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func groupLabelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetGroupLabels(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func commentListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	result, err := data.GetAllComments(token, p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func newsCommentListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	token := common.GetAccessToken(r);
	result, err := data.GetAllNewsComments(token, p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getGroupByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetGroupById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getPostByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetPostById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getNewsByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetNewsById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func labelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetAllLabels();

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func userLabelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	token := common.GetAccessToken(r);
	groupId := r.URL.Query().Get("groupId");
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	result, err := data.GetUserLabels(token, groupId, lat, lng);

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func updateGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateGroup(p.ByName("id"), group);

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendGroup(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func uploadGroupIcon(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateGroupLogo(p.ByName("id"), group.LogoData);

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func uploadGroupImage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateGroupImage(p.ByName("id"), group.ImageData);

	if (err != nil) {
		writeErrorResponse(rw, r, p, group, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func updateLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var label models.Label
	err := json.NewDecoder(r.Body).Decode(&label)

	if (err != nil) {
		writeErrorResponse(rw, r, p, label, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateLabel(p.ByName("id"), label);

	if (err != nil) {
		writeErrorResponse(rw, r, p, label, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func removeGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.RemoveGroup(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func getLabelByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetLabelById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func loginAdminHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	var token string

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	err = data.SetUserToken(token, user.Username);

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	authUser := struct {
		User  models.User `json:"user"`
		Token string `json:"token"`
	}{
		user,
		token,
	}

	rw.Write(common.SuccessResponseJSON(authUser));
}

func memberRegisterHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var user models.User;
	err := json.NewDecoder(r.Body).Decode(&user);
	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	var result models.User;
	result, err = data.RegisterAppUser(user);

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getUserInfoHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	token := common.GetAccessToken(r);
	user, err := data.GetUserInfo(token);
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(user));
}

func getActiveUsersList(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := data.GetUsers();
	if (err != nil) {
		writeErrorResponse(rw, r, p, "", http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(user));
}

func registerAdminHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user);

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	err = data.RegisterUser(user);

	if (err != nil) {
		writeErrorResponse(rw, r, p, user, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(user));

}

func bulkInsertSubCategoryHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var subCategories []models.GroupSubCategory
	err := json.NewDecoder(r.Body).Decode(&subCategories);
	if (err != nil) {
		writeErrorResponse(rw, r, p, subCategories, http.StatusInternalServerError, err);
		return
	}

	err = data.CreateSubCategories(subCategories)

	if (err != nil) {
		writeErrorResponse(rw, r, p, subCategories, http.StatusInternalServerError, err);
		return
	}
}

func submitFeedbackHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var feedback struct {
		Phone   string `json:"phone,omitempty" bson:"phone,omitempty"`
		Email   string `json:"email,omitempty" bson:"email,omitempty"`
		Content string `json:"content,omitempty" bson:"content,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&feedback)

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	err = data.SetUserFeedback(token, feedback.Content, feedback.Phone, feedback.Email);

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func reportSpamHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var feedback struct {
		PostId  string `json:"postId,omitempty" bson:"postId,omitempty"`
		Content string `json:"content,omitempty" bson:"content,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&feedback)

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	err = data.ReportSpam(feedback.PostId, feedback.Content);

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func reportNewsSpamHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var feedback struct {
		NewsId  string `json:"newsId,omitempty" bson:"newsId,omitempty"`
		Content string `json:"content,omitempty" bson:"content,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&feedback)

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	err = data.ReportNewsSpam(feedback.NewsId, feedback.Content);

	if (err != nil) {
		writeErrorResponse(rw, r, p, feedback, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON("success"));
}

func writeErrorResponse(rw http.ResponseWriter, r *http.Request, p httprouter.Params, body interface{}, statusCode int, err error) {
	errMsg := r.Method + ": " + r.URL.String();
	errMsg += " ";
	if ((r.Method == "GET" || r.Method == "PUT") && len(p.ByName("id")) > 0) {
		errMsg += "Params: " + p.ByName("id");
		errMsg += " ";
	}
	if (r.Method == "POST") {
		bodyBuff, _ := json.Marshal(body)
		if (len(string(bodyBuff)) > 0) {
			errMsg += "Req Body: " + string(bodyBuff);
			errMsg += " ";
		}
	}

	errMsg += err.Error();

	slack.Send(slack.ErrorLevel, errMsg)
	rw.WriteHeader(statusCode);
	rw.Write(common.ResponseJson(nil, common.ResponseError(statusCode, err.Error())))
}
