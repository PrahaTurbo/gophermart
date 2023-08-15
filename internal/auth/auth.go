package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type UserIDKeyType string

const jwtTokenCookieName string = "token"
const UserIDKey UserIDKeyType = "userID"

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func CreateJWTAuthCookie(userID int, jwtSecret string) (*http.Cookie, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     jwtTokenCookieName,
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	}

	return cookie, nil
}

func Auth(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(jwtTokenCookieName)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				return []byte(jwtSecret), nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
