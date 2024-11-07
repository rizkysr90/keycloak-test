package login

import (
	auth "authorization_flow_oauth/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "login/login.tmpl", gin.H{
		"Title":       "Welcome",
		"Message":     "Please log in to continue",
		"KeycloakURL": "http://localhost:8081/login-keycloak",
	})
}

func WithKeycloak(c *gin.Context, authClient *auth.Client) {
	c.Redirect(http.StatusFound, authClient.Oauth.AuthCodeURL("sadasad"))
}
