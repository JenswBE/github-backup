package git

import (
	"fmt"
	"net/url"
)

var _ fmt.Stringer = AuthenticatedURL("")

type AuthenticatedURL string

func (u AuthenticatedURL) String() string {
	return string(u)
}

func GetAuthenticatedURL(repoURL, user, password string) (AuthenticatedURL, error) {
	authURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s as URL: %w", repoURL, err)
	}
	authURL.User = url.UserPassword(user, password)
	return AuthenticatedURL(authURL.String()), nil
}
