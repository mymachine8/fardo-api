package common

import (
	"encoding/json"
	"net/http"
	"github.com/mymachine8/fardo-api/models"
	"google.golang.org/cloud/storage"
	"golang.org/x/net/context"
	"strings"
	"log"
)

var (
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

type (
	appError struct {
		Error      string `json:"error"`
		Message    string `json:"message"`
		HttpStatus int    `json:"status"`
	}
	errorResource struct {
		Data appError `json:"data"`
	}
	configuration struct {
		MongoDBHost, DBUser, DBPwd, Database string
	}
)

func DisplayAppError(w http.ResponseWriter, handlerError error, message string, code int) {
	errObj := appError{
		Error:      handlerError.Error(),
		Message:    message,
		HttpStatus: code,
	}
	log.Printf("AppError]: %s\n", handlerError)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if j, err := json.Marshal(errorResource{Data: errObj}); err == nil {
		w.Write(j)
	}
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

var AppConfig configuration

// Initialize AppConfig
func initConfig() {

	var err error
	AppConfig.MongoDBHost = "104.155.143.185"
	AppConfig.DBUser = ""
	AppConfig.DBPwd = ""
	AppConfig.Database = "fardo"
	StorageBucketName = "go-server"
	StorageBucket, err = configureStorage(StorageBucketName)
	if(err != nil) {
		log.Print(err.Error())
	}
}

func SuccessResponseJSON(result interface{}) []byte {
	return ResponseJson(result, models.ResponseError{});
}

func ResponseJson(result interface{}, responseError models.ResponseError) []byte {
	response := struct {
		Data  interface{} `json:"data"`
		Error models.ResponseError `json:"error,omitempty"`
	}{
		result,
		responseError,
	}

	jsonResult, err := json.Marshal(response);
	if (err != nil) {
		log.Panic(err);
	}

	log.Println(string(jsonResult));

	return jsonResult
}

func ResponseError(code int, message string) models.ResponseError {
	log.Print(message);
	errObj := models.ResponseError{
		Code:      code,
		Message:    message,
	}
	return errObj;
}

func GetAccessToken(req *http.Request) string {
	if ah := req.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return ah[7:]
		}
	}
	return "";
}
