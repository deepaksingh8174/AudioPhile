package middlewares

import (
	"context"
	"net/http"
	"test/database/dbhelper"
	"test/models"
	"test/utils"
)

func UserMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter,r *http.Request) {
		claims := r.Context().Value("claims").(*models.Claims)
		userID := claims.UserId
		roles,userErr := dbhelper.CheckRole(userID)
		if userErr != nil {
			utils.RespondJSON(w,http.StatusInternalServerError,utils.Status{Message: "error occurred during fetching of db"})
			return
		}
		isUser := false
		for _,value := range roles {
				if value == models.UserRole {
					isUser = true
					break
				}
			}
		if !isUser {
			utils.RespondJSON(w,http.StatusUnauthorized,utils.Status{Message: "unauthorised access"})
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func AdminMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter,r *http.Request) {
		claims := r.Context().Value("claims").(*models.Claims)
		userID := claims.UserId
		roles,userErr := dbhelper.CheckRole(userID)
		if userErr != nil {
			utils.RespondJSON(w,http.StatusInternalServerError,utils.Status{Message: "error occurred during fetching of db"})
			return
		}
		isAdmin := false
		for _,value := range roles {
			if value == models.AdminRole{
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			utils.RespondJSON(w,http.StatusUnauthorized,utils.Status{Message: "unauthorised access"})
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

