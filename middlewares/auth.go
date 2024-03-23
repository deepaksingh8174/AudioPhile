package middlewares

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"test/models"
	"test/utils"
)

func JWTMiddleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			utils.RespondJSON(w,http.StatusUnauthorized,utils.Status{Message: "unauthorised access"})
			return
		}
		token,err:= jwt.ParseWithClaims(tokenString,&models.Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("secretKey")),nil
		})
		if err != nil || !token.Valid {
			utils.RespondJSON(w,http.StatusUnauthorized,utils.Status{Message: "unauthorised access"})
			return
		}
		ctx := context.WithValue(r.Context(),"claims",token.Claims.(*models.Claims))
		next.ServeHTTP(w,r.WithContext(ctx))
	})
}


