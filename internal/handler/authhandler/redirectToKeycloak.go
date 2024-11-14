package authhandler

import (
	"authorization_flow_oauth/internal/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func (a *AuthHandler) RedirectToKeycloak(c *gin.Context) {
	stateID, err := utils.GenerateRandomBase64Str()
	log.Println("HEREE : ", stateID)
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to generate redirect state : ", err))
	}
	codeVerifier, err := utils.GenerateRandomBase64Str()
	if err != nil {
		c.Error(utils.ErrorBuilder("failed to generate code verifier : ", err))
	}
	log.Println("HEREEEEEXX : ", codeVerifier)
	codeChallenge := utils.GenerateCodeChallenge(codeVerifier)
	if err = a.authStore.SetState(c, stateID); err != nil {
		c.Error(utils.ErrorBuilder("failed to save stateIDKey in redis : ", err))
	}
	if err = a.authStore.SetCodeVerifier(c, codeVerifier, stateID); err != nil {
		c.Error(utils.ErrorBuilder("failed to save codeVerifierKey in redis : ", err))
	}
	oauth2opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	c.Redirect(http.StatusFound, a.authClient.Oauth.AuthCodeURL(stateID, oauth2opts...))
}
