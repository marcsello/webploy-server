package authentication

import (
	"encoding/csv"
	"fmt"
	httpAuth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type BasicAuthProvider struct {
	creds                 map[string]string
	wwwAuthenticateHeader string
}

func NewBasicAuthProvider(htpasswdFilePath string) (*BasicAuthProvider, error) {
	creds, err := loadBasicAuthCredentials(htpasswdFilePath)
	if err != nil {
		return nil, err
	}

	return &BasicAuthProvider{
		creds:                 creds,
		wwwAuthenticateHeader: `Basic realm="webploy", charset="UTF-8"`,
	}, nil

}

func loadBasicAuthCredentials(htpasswdFilePath string) (map[string]string, error) {
	// Adopted from here: https://github.com/abbot/go-http-auth/blob/master/users.go
	var err error
	var f *os.File
	f, err = os.Open(htpasswdFilePath) //#nosec G304
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	var records [][]string
	records, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}

	users := make(map[string]string)
	for _, record := range records {
		name := record[0]
		encryptedPass := record[1]
		err = ValidateUsername(name)
		if err != nil {
			return nil, err
		}
		if encryptedPass == "" {
			return nil, fmt.Errorf("empty password field")
		}
		users[name] = encryptedPass
	}

	return users, nil
}

func validateUserPass(users map[string]string, username, password string) bool {
	storedHash, ok := users[username]
	if !ok {
		// invalid user
		return false
	}
	if !httpAuth.CheckSecret(password, storedHash) {
		// invalid password
		return false
	}

	return true
}

func (ba *BasicAuthProvider) NewMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username, password, ok := ctx.Request.BasicAuth()

		// we only validate usernames coming from "outside", the software may still use "invalid" usernames internally (e.g.: system user has prefix)
		if !ok || ValidateUsername(username) != nil || !validateUserPass(ba.creds, username, password) {
			// no credentials provided, or the provided credentials are bad
			ctx.Header("WWW-Authenticate", ba.wwwAuthenticateHeader)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Auth successful

		ctx.Set(ContextAuthenticatedUserKey, username)
	}
}
