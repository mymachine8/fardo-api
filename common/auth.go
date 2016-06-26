package common

import (
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// using asymmetric crypto/RSA keys
// location of private/public key files
const (
	// openssl genrsa -out app.rsa 1024
	privKeyPath = "keys/app.rsa"
	// openssl rsa -in app.rsa -pubout > app.rsa.pub
	pubKeyPath = "keys/app.rsa.pub"
)

const (
	privateKey = "-----BEGIN RSA PRIVATE KEY-----MIICXgIBAAKBgQC8RzRevw1MVf80PhhDbKJZmhXYflOaeYIOsmnIFjDTggjWbNI2z8jvZVrKLZRo9wopae1TDY2YfwGv2pcJjqn79j6JHTUrqA6M5nE+xjQQSQIDAQABogjp1H3mGiKQyATOo8Ehp+KR5Xfdx93hy9xOEwlXHB0PWpZW5zDvW/zpbAImni+ZAoGBALTw3VSc2WPeVbfYYSsTEOd5nLsFlMUlNyd2sRB4uw3ZrzKbPF8C0/wcma30uflYML6CQ45bsPOzisHaXdOtPpcnsTr8cytncY6q8eZK1a4Fz4GD+F4AbBS07cG1AYdrMgwhMyx4NzVHL4oFznU+ahQL7On4zuWRQ9IbvEmiTyKdAkEA4b2pxnN5P9vAEVTkpq813ziJUlmsSyg7r0IwpL5YQRt+y8Z+abOWTuzhh7aUQIH5TaqJe17KiW8ZbGey85id4wJBANWEAFv1NnIlaQNjhZDIhpPNBwGRZiGszDVjlbLRgMnskmdty64bnPIIMVPcztpPZPXoNpQTQ/B0p0xGy0GcsOMCQC/yeP0NydMmecU0otxEmsyu1XwIT/Au4aeKHbVdyldDbF6l58b5fEdn0mUHikVcj5s2oa5u4s1bdD4tanH4MECQQDE63BVX2uujNg0WuZFqNuNlwt+I7ZZGoBgQQ9Ak74+/SPtpjKyyh7OjkXIPZ69c3n+3gLwQHBpZX0ieSxev//XAkEArxEa8k/QKyeVdx9hqsSIpDl0SS4wcGhN2LKSq8KyCMFnm4VxC5eO9JGhYTv8+N4ltmBu1VGvCUqjtyjh4aXZPw==-----END RSA PRIVATE KEY-----"
	publicKey = "-----BEGIN PUBLIC KEY-----MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC8RzRevw1MVf80PhhDbKJZmhXYflOaeYIOsmnIFjDTggjWbNI2ogjp1H3mGiKQyATOo8Ehp+KR5Xfdx93hy9xOEwlXHB0PWpZW5zDvW/zpbAImni+Zz8jvZVrKLZRo9wopae1TDY2YfwGv2pcJjqn79j6JHTUrqA6M5nE+xjQQSQIDAQAB-----END PUBLIC KEY-----"
)

// Private key for signing and public key for verification
var (
	verifyKey, signKey []byte
)

// Read the key files before starting http handlers
func initKeys() {
	signKey = []byte(privateKey);
	verifyKey = []byte(publicKey);
}

// Generate JWT token
func GenerateJWT(name, role string) (string, error) {
	// create a signer for rsa 256
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	// set claims for JWT token
	t.Claims["iss"] = "admin"
	t.Claims["UserInfo"] = struct {
		Name string
		Role string
	}{name, role}

	// set the expire time for JWT token

	//TODO: For admin expire it for every day
	t.Claims["exp"] = time.Now().Add(time.Hour * 800).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func BasicAuth(h httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// validate the token
		token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {

			// Verify the token with public key, which is the counter part of private key
			return verifyKey, nil
		})

		if err != nil {
			log.Print(err.Error());
			switch err.(type) {

			case *jwt.ValidationError: // JWT validation error
				vErr := err.(*jwt.ValidationError)

				switch vErr.Errors {
				case jwt.ValidationErrorExpired: //JWT expired
					rw.WriteHeader(http.StatusInternalServerError);
					rw.Write(ResponseJson(nil, ResponseError(http.StatusInternalServerError, "Token Expired")));
					return;

				default:
					rw.WriteHeader(http.StatusInternalServerError);
					rw.Write(ResponseJson(nil, ResponseError(http.StatusInternalServerError, "Error Validating Token")));
					return
				}

			}
		}

		if token.Valid {
			h(rw, r, ps)
		}
	}
}

// Middleware for validating JWT tokens

