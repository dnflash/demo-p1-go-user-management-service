package server

import (
	"github.com/dnflash/demo-p1-go-user-management-service/internal/context"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"log"
	"net/http"
	"strings"
)

func (s Server) authMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		at := r.Header.Get("Authorization")
		if strings.HasPrefix(at, "Bearer ") {
			at = strings.TrimPrefix(at, "Bearer ")
			token, err := jwt.Parse([]byte(at), jwt.WithKey(jwa.HS256, s.AccessTokenSecret), jwt.WithValidate(true))
			if err != nil {
				log.Printf("authMw: Failed to validate access token, err: %v", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			typeClaim, ok := token.Get("type")
			if !ok {
				log.Printf("authMw: Invalid access token, missing type")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			tokenType, ok := typeClaim.(string)
			if !ok || tokenType != "access-token" {
				log.Printf("authMw: Invalid token type")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			subClaim, ok := token.Get("sub")
			if !ok {
				log.Printf("authMw: Invalid access token, missing subject")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			userID, ok := subClaim.(string)
			if !ok {
				log.Printf("authMw: Invalid access token subject")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			roleClaim, ok := token.Get("role")
			if !ok {
				log.Printf("authMw: Invalid access token, missing role")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			role, ok := roleClaim.(string)
			if !ok {
				log.Printf("authMw: Invalid access token, invalid role")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.SetUserContext(r.Context(), context.UserContext{UserID: userID, Role: role})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	})
}

func (s Server) adminAccessMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uc, err := context.GetUserContext(r.Context())
		if err != nil {
			log.Printf("adminAccessMw: Error getting user context, err: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if uc.Role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	})
}
