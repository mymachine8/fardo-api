package controllers

import (
	"github.com/julienschmidt/httprouter"
	"github.com/google/go-gcm"
	"net/http"
	"fmt"
	"github.com/mymachine8/fardo-api/models"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"github.com/mymachine8/fardo-api/data"
	"github.com/mymachine8/fardo-api/common"
	"log"
	"github.com/rs/cors"
	"github.com/mymachine8/fardo-api/slack"
)

func InitRoutes() http.Handler {
	r := httprouter.New();

	r.GET("/", helloWorldHandler);
	//http.HandleFunc("/", hellWorldHandler);

	http.Handle("/", r);

	//TODO: Other routers come here
	r.GET("/api/my-circle", myCircleHandler);


	r.GET("/api/categories", common.BasicAuth(categoryListHandler));
	r.GET("/api/categories/:id/sub-categories", common.BasicAuth(subCategoryListHandler));
	r.POST("/api/sub-categories/bulk-insert", bulkInsertSubCategoryHandler);


	r.GET("/api/groups", groupListHandler);
	r.GET("/api/groups/:id", getGroupByIdHandler);
	r.POST("/api/groups", createGroupHandler);
	r.PUT("/api/groups/:id", updateGroupHandler);
	r.DELETE("/api/groups/:id", removeGroupHandler);

	r.GET("/api/groups/:id/labels", groupLabelListHandler);
	r.GET("/api/labels", labelListHandler);
	r.GET("/api/labels/:id", getLabelByIdHandler);
	r.POST("/api/groups/:id/labels", createLabelHandler);
	r.PUT("/api/labels/:id", updateLabelHandler);
	r.PUT("/api/labels/:id/activate", updateLabelHandler);
	r.PUT("/api/labels/:id/suspend", updateLabelHandler);
	r.DELETE("/api/labels/:id", removeLabelHandler);
	r.POST("/api/labels/bulk", createLabelsBulkHandler);


	r.POST("/api/admin/register", registerAdminHandler);
	r.POST("/api/admin/login", loginAdminHandler);
	r.POST("/api/users", memberRegisterHandler);
	r.PUT("/api/users/group", updateUserGroupHandler);

	r.GET("/api/featured-groups", featuredGroupsHandler);

	r.GET("/api/admin/posts", allPostsListHandler);
	r.GET("/api/admin/posts/current", currentPostsListHandler);
	r.POST("/api/admin/posts", createAdminPostHandler);
	r.POST("/api/posts", createPostHandler);
	r.PUT("/api/posts/:id/upvote", upvotePostHandler);
	r.PUT("/api/posts/:id/downvote", downvotePostHandler);
	r.PUT("/api/posts/:id/suspend", suspendPostHandler);
	r.GET("/api/posts/:id/comments", commentListHandler);
	r.POST("/api/posts/:id/comments", createCommentHandler);
	r.PUT("/api/comments/:id/upvote", upvoteCommentHandler);
	r.PUT("/api/comments/:id/downvote", downvoteCommentHandler);

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
	fmt.Fprintln(rw, "Hello World")
}

func myCircleHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	v := r.URL.Query()
	fmt.Println(v["lat"]);
	var latLong [2]float64 = [2] float64{3232.323, 32323.3232};
	result := models.MyPosts(bson.NewObjectId(), latLong);
	rw.Header().Set("Content-Type", "application/json")

	var error models.FardoError
	response := struct {
		Data  [] models.Post `json:"data"`
		Error models.FardoError `json:"error,omitempty"`
	}{
		result,
		error,
	}

	apiKey := "AIzaSyDcXHtO1YB-tCRwFAvaVqJDud8gR00VJs4"
	m := gcm.HttpMessage{};
	m.To = "euF-S2d-RFk:APA91bHMBhCCHQVJ0uAeLOHiF9v3azYnzTn1tLx8fz7u2_7nCQMExJt6QZxJLeEQK4W1yrSi0vqCLtEVPOhQwHnXbkRKUZRunXDTz1ZOZKi4F5RrdWN-JRvgyFg2nV6yiwhZ2eWQbd2y" //Registration Token of the user
	m.CollapseKey = "" //A group tag which specifies all the messages send are of the same type, so that server can send recent msg when the device is awake
	m.Priority = "" //Valid values "normal"/ "high"
	m.DelayWhileIdle = true //When this parameter is set to true, it indicates that the message should not be sent until the device becomes active.
	// m.TimeToLive 0 to 2,419,200 seconds
	x := map[string]interface{}{
		"bar": "foo",
		"baz": "hello world",
	}
	//gcmMsg := models.GcmMessage {"Hello World", "GCM"}
	m.Data = x//This should be flat struct with strings, Values in string types are recommended. You have to convert values in objects or other non-string data types (e.g., integers or booleans) to string.
	//m.Notification will be sent as notification to the android device, to the notifications tray

	gcmResponse, e := gcm.SendHttp(apiKey, m);

	if (e != nil) {
		println(e.Error())
	}

	println(gcmResponse.Failure);
	println(gcmResponse.Success);

	jsonResult, _ := json.Marshal(response);

	rw.Write(jsonResult)
}

func featuredGroupsHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	token := common.GetAccessToken(r);
	result, err := data.GetFeaturedGroups(token);
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func allPostsListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetAllPosts();
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func currentPostsListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetCurrentPosts();
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));

}

func createAdminPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);

	id, err := data.CreatePostAdmin(token, post);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	log.Print("Post came here");
	log.Print(post);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	token := common.GetAccessToken(r);
	log.Print(token);

	id, err := data.CreatePostUser(token, post);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var comment models.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	log.Print(comment);

	id, err := data.AddComment(comment);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func upvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.UpvotePost(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvotePostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.DownvotePost(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func upvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.UpvoteComment(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func downvoteCommentHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.DownvoteComment(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func suspendPostHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.SuspendPost(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func updateUserGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var body struct{ GroupId string };
	err := json.NewDecoder(r.Body).Decode(&body)

	token := common.GetAccessToken(r);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusBadRequest, err);
		return
	}

	err = data.UpdateUserGroup(token, body.GroupId);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func createGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	id, err := data.CreateGroup(group);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func createLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var label models.Label
	err := json.NewDecoder(r.Body).Decode(&label)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	id, err := data.CreateLabel(p.ByName("id"), label);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(id));
}

func removeLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := data.RemoveLabel(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func createLabelsBulkHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//TODO: Create Labels for Bulk
}

func categoryListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetAllCategories();
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));

}

func subCategoryListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetSubCategories(p.ByName("id"));
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));

}

func groupListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	searchStr := r.URL.Query().Get("name");
	var err error
	var result []models.Group
	if (len(searchStr) > 0) {
		result, err = data.GetGroups(searchStr);
	} else {
		result, err = data.GetAllGroups();
	}
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
	rw.Write(common.SuccessResponseJSON(result));
}

func groupLabelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetGroupLabels(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func commentListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetAllComments(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func getGroupByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetGroupById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func labelListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	result, err := data.GetAllLabels();

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func updateGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateGroup(p.ByName("id"), group);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func updateLabelHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var label models.Label
	err := json.NewDecoder(r.Body).Decode(&label)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	err = data.UpdateLabel(p.ByName("id"), label);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func removeGroupHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	err := data.RemoveGroup(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(p.ByName("id")));
}

func getLabelByIdHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	result, err := data.GetLabelById(p.ByName("id"));

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	rw.Write(common.SuccessResponseJSON(result));
}

func loginAdminHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	var token string

	// Authenticate the login user
	//Return user
	var loginUser models.User;
	if loginUser, err = data.Login(user); err != nil {

		return
	}

	// Generate JWT token
	token, err = common.GenerateJWT(user.Username, "admin")
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	err = data.SetUserToken(token, loginUser.Id.Hex());

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	// Clean-up the hashpassword to eliminate it from response JSON
	user.HashPassword = nil;

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
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user);
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	var userId string;
	userId, err = data.RegisterAppUser(user);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	var token string

	token, err = common.GenerateJWT(user.Imei, "member")

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	log.Print(userId);
	err = data.SetUserToken(token, userId);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
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

func registerAdminHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	err = data.RegisterUser(user);

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	user.HashPassword = nil

	rw.Write(common.SuccessResponseJSON(user));

}

func bulkInsertSubCategoryHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var subCategories []models.GroupSubCategory
	err := json.NewDecoder(r.Body).Decode(&subCategories);
	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}

	err = data.CreateSubCategories(subCategories)

	if (err != nil) {
		writeErrorResponse(rw, http.StatusInternalServerError, err);
		return
	}
}

func writeErrorResponse(rw http.ResponseWriter, statusCode int, err error) {
	slack.Send(slack.ErrorLevel, err.Error())
	rw.WriteHeader(statusCode);
	rw.Write(common.ResponseJson(nil, common.ResponseError(statusCode, err.Error())))
}
