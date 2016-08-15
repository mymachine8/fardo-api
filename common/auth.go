package common

import (
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

// using asymmetric crypto/RSA keys
// location of private/public key files
const (
	// openssl genrsa -out app.rsa 1024
	privKeyPath = "/rsakeys/app.rsa"
	// openssl rsa -in app.rsa -pubout > app.rsa.pub
	pubKeyPath = "/rsakeys/app.rsa.pub"
)

// Private key for signing and public key for verification
var (
	verifyKey, signKey []byte
)

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
			h(rw,r,ps)
		}
	}
}