package common

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"github.com/mymachine8/fardo-api/models"
	"strings"
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
		Server, MongoDBHost, DBUser, DBPwd, Database string
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

// AppConfig holds the configuration values from config.json file
var AppConfig configuration

// Initialize AppConfig
func initConfig() {
	file, err := os.Open("common/config.json")
	defer file.Close()
	if err != nil {
		log.Fatalf("[loadConfig]: %s\n", err)
	}
	decoder := json.NewDecoder(file)
	AppConfig = configuration{}
	err = decoder.Decode(&AppConfig)
	if err != nil {
		log.Fatalf("[loadAppConfig]: %s\n", err)
	}
}

func SuccessResponseJSON(result interface{}) []byte {
	return ResponseJson(result, models.ResponseError{});
}

func ResponseJson(result interface{}, responseError models.ResponseError) []byte {
	response := struct {
		Data interface{} `json:"data"`
		Error models.ResponseError `json:"error,omitempty"`
	} {
		result,
		responseError,
	}

	jsonResult, err := json.Marshal(response);
	if(err !=nil) {
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

// Reads config.json and decode into AppConfig
