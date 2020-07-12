package auth

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
	"os"
	// "github.com/gin-gonic/gin"
	// "errors"
	// "net/http"
)

func CreateToken(userId int) (bool, string) {
	os.Setenv("LETSCONNECT_KEY", "bazinga")
	jwtClaims := jwt.MapClaims{}
	jwtClaims["authorized"] = true
	jwtClaims["user_id"] = userId
	jwtClaims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	token, err := at.SignedString([] byte(os.Getenv("LETSCONNECT_KEY")) )
	if err != nil {
		log.Println("JWT generation error: ", err)
		return false, ""
	}
	return true, token
}

// func TokenValid(r *http.Request) error {

// }
