package authhandler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/internal/utils"
	"authorization_flow_oauth/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	cfg                  *config.Config
	serverAddr           string
	authClient           *auth.Client
	redisClient          *redis.Client
	redisStateIDKey      string
	redisCodeVerifierKey string
}

func New(
	cfg *config.Config,
	serverAddr string,
	authClient *auth.Client,
	redisClient *redis.Client,
) *AuthHandler {
	return &AuthHandler{
		cfg:                  cfg,
		serverAddr:           serverAddr,
		authClient:           authClient,
		redisClient:          redisClient,
		redisStateIDKey:      "authstate",
		redisCodeVerifierKey: "authverifier",
	}
}

func (a *AuthHandler) RenderLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login/login.tmpl", gin.H{
		"Title":       "Welcome",
		"Message":     "Please log in to continue",
		"KeycloakURL": fmt.Sprintf("http://%s/login-keycloak", a.serverAddr),
	})
}
func (a *AuthHandler) RedirectToKeycloak(c *gin.Context) {
	stateID, err := utils.GenerateRandomBase64Str()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to generate redirect state : ", err))
	}
	codeVerifier, err := utils.GenerateRandomBase64Str()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to generate code verifier : ", err))
	}
	codeChallenge := utils.GenerateCodeChallenge(codeVerifier)
	// We need to store redirect state in redis for the future callback
	stateIDKey := fmt.Sprintf("%s:%s", a.redisStateIDKey, stateID)
	err = a.redisClient.Set(c, stateIDKey, stateID, 3*time.Minute).Err()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to save stateIDKey in redis : ", err))
	}
	// We need to store code challenge in redis for the future callback
	codeVerifierKey := fmt.Sprintf("%s:%s", a.redisCodeVerifierKey, stateID)
	err = a.redisClient.Set(c, codeVerifierKey, codeVerifier, 3*time.Minute).Err()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to save codeVerifierKey in redis : ", err))
	}
	oauth2opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	c.Redirect(http.StatusFound, a.authClient.Oauth.AuthCodeURL(stateID, oauth2opts...))
}
func (a *AuthHandler) Callback(c *gin.Context) {
	stateID := c.Query("state")
	if stateID == "" {
		c.Error(utils.ErrorBuilder("stateID is required", nil))
	}
	authzCode := c.Query("code")
	if authzCode == "" {
		c.Error(utils.ErrorBuilder("authzCode is required", nil))
	}
	stateIDKey := fmt.Sprintf("%s:%s", a.redisStateIDKey, stateID)
	stateIDValue, err := a.redisClient.Get(c, stateIDKey).Result()
	if err != nil || stateIDValue == "" {
		c.Error(utils.ErrorBuilder("failed to get stateIDKey value", err))
	}
	if stateIDValue != stateID {
		c.Error(utils.ErrorBuilder("invalid stateID", nil))
	}
	codeVerifierKey := fmt.Sprintf("%s:%s", a.redisCodeVerifierKey, stateID)
	codeVerifierValue, err := a.redisClient.Get(c, codeVerifierKey).Result()
	if err != nil || codeVerifierValue == "" {
		c.Error(utils.ErrorBuilder("failed to get codeVerifierKey value", err))
	}
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_verifier", codeVerifierKey),
	}
	oauth2Token, err := a.authClient.Oauth.Exchange(c, authzCode, opts...)
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to exchange oauth token", err))
	}
	log.Println(oauth2Token)

}
