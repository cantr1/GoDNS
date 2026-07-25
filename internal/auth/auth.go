package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(req http.Request) (string, error) {
	rawTokenString := req.Header.Get("Authorization")
	if rawTokenString == "" {
		return "", fmt.Errorf("authorization header missing")
	}
	token := strings.TrimPrefix(rawTokenString, "Bearer ")
	return token, nil
}
