package common

import (
	"encoding/json"
	"net/http"
	"github.com/mymachine8/fardo-api/models"
	"google.golang.org/cloud/storage"
	"golang.org/x/net/context"
	"strings"
	"log"
	"math"
	"time"
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
	AppConfig.MongoDBHost = "localhost:27017"
	AppConfig.DBUser = "prodUserAdmin"
	AppConfig.DBPwd = "machine8"
	AppConfig.Database = "zing-prod"
	StorageBucketName = "go-server"
	StorageBucket, err = configureStorage(StorageBucketName)
	if (err != nil) {
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

func DivisbleByPowerOf2(num int) bool {
	var result = 5;
	for i := 3; result <= num; i++ {
		result = int(math.Pow(2, float64(i)));
		if (num % result == 0) {
			return true;
		}
	}
	return false;
}

func GetTimeSeconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Second)
}

func GetZingCreationTimeSeconds() int64 {
	t1, _ := time.Parse(
		time.RFC3339,
		"2016-07-01T00:00:00+00:00");

	return t1.UnixNano() / int64(time.Second)
}

func MinInt(a int, b int) int {
	if (a < b) {
		return a;
	}
	return b;
}

func IsPowerOf2(a int) bool {
	if (a <= 1) {
		return false;
	}
	if (a == 2) {
		return true;
	}

	return IsPowerOf2(a / 2)
}

func DistanceLatLong(lat1 float64, lat2 float64, lon1 float64, lon2 float64) float64 {

	R := 6371; // Radius of the earth

	latDistance := Radians(lat2 - lat1);
	lonDistance := Radians(lon2 - lon1);
	a := math.Sin(latDistance / 2) * math.Sin(latDistance / 2) + math.Cos(Radians(lat1)) * math.Cos(Radians(lat2)) * math.Sin(lonDistance / 2) * math.Sin(lonDistance / 2);
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1 - a));
	distance := float64(R) * c * 1000; // convert to meters

	distance = math.Pow(distance, 2);

	return math.Sqrt(distance);

}

func Radians(d float64) float64 {
	const x = math.Pi / 180;
	return d * x;
}