package auth

import (
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testJWTsecret = "test-secret"

func TestCreateJWTAuthCookie(t *testing.T) {

	type args struct {
		userID    int
		jwtSecret string
	}

	tests := []struct {
		name    string
		args    args
		want    *http.Cookie
		wantErr bool
	}{
		{
			name: "should return cookie with valid token",
			args: args{
				userID:    1,
				jwtSecret: testJWTsecret,
			},
			want: &http.Cookie{
				Name:     JWTTokenCookieName,
				Value:    genJWTToken(1),
				HttpOnly: true,
				Path:     "/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateJWTAuthCookie(tt.args.userID, tt.args.jwtSecret)

			assert.Equal(t, tt.wantErr, (err != nil))
			assert.Equal(t, tt.want, got)
		})
	}
}

func genJWTToken(userID int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, _ := token.SignedString([]byte(testJWTsecret))

	return tokenString
}
