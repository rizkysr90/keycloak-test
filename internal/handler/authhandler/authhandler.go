package authhandler

import (
	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/internal/store"
	"authorization_flow_oauth/internal/utils"
	"authorization_flow_oauth/pkg/auth"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg        *config.Config
	serverAddr string
	authClient *auth.Client
	authStore  store.AuthStore
}

func New(
	cfg *config.Config,
	serverAddr string,
	authClient *auth.Client,
	authStore store.AuthStore,
) *AuthHandler {
	return &AuthHandler{
		cfg:        cfg,
		serverAddr: serverAddr,
		authClient: authClient,
		authStore:  authStore,
	}
}

func (a *AuthHandler) RedirectToKeycloak(c *gin.Context) {
	stateID, err := utils.GenerateRandomBase64Str()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to generate redirect state : ", err))
	}
	if err = a.authStore.SetState(c, stateID); err != nil {
		c.Error(utils.ErrorBuilder("failed to save stateIDKey in redis : ", err))
	}
	c.Redirect(http.StatusFound, a.authClient.Oauth.AuthCodeURL(stateID))
}

func (a *AuthHandler) RenderLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login/login.tmpl", gin.H{
		"Title":       "Welcome",
		"Message":     "Please log in to continue",
		"KeycloakURL": fmt.Sprintf("http://%s/login-keycloak", a.serverAddr),
	})
}
