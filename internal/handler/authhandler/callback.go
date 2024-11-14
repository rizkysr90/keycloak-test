package authhandler

import (
	"authorization_flow_oauth/internal/store"
	"errors"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func (a *AuthHandler) Callback(c *gin.Context) {
	callbackData, err := newCallbackData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	if err = callbackData.verify(c, a.authStore); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	}
	var oauth2Token *oauth2.Token
	oauth2Token, err = a.authClient.Oauth.Exchange(c, callbackData.authzCode, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	var oidcToken *oidcToken
	oidcToken, err = newOIDCToken(oauth2Token, a.authClient.OIDC)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	var userInfoClaims *userInfoClaims
	userInfoClaims, err = oidcToken.getClaims(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	log.Println(oauth2Token.AccessToken)
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  oauth2Token.AccessToken,
		"refresh_token": oauth2Token.RefreshToken,
		"user_info":     userInfoClaims,
	})
	// a.authClient.OIDC.Verify(c, oauth2Toke)
}

type callbackData struct {
	stateID   string
	authzCode string
}

func newCallbackData(c *gin.Context) (*callbackData, error) {
	stateID := c.Query("state")
	if stateID == "" {
		return nil, errors.New("stateID is required")
		// nil, c.Error(utils.ErrorBuilder("stateID is required", nil))
	}
	authorizationCode := c.Query("code")
	if authorizationCode == "" {
		// c.Error(utils.ErrorBuilder("authorizationCode is required", nil))
		return nil, errors.New("authorizationCode is required")
	}
	return &callbackData{
		stateID:   stateID,
		authzCode: authorizationCode,
	}, nil
}
func (c *callbackData) verify(ctx *gin.Context, authStore store.AuthStore) error {
	stateIDData, err := authStore.GetState(ctx, c.stateID)
	if err != nil {
		return err
	}
	if stateIDData != c.stateID {
		return errors.New("invalid stateID")
	}
	if err = authStore.DeleteState(ctx, stateIDData); err != nil {
		return err
	}
	return nil
}

type oidcToken struct {
	rawIDToken string
	verifier   *oidc.IDTokenVerifier
}

func newOIDCToken(oauthToken *oauth2.Token,
	verifier *oidc.IDTokenVerifier) (*oidcToken, error) {
	// Get raw ID token
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in response")
	}
	return &oidcToken{
		rawIDToken: rawIDToken,
		verifier:   verifier,
	}, nil
}

type userInfoClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Get raw ID token
func (o *oidcToken) getClaims(c *gin.Context) (*userInfoClaims, error) {
	idToken, err := o.verifier.Verify(c, o.rawIDToken)
	if err != nil {
		return nil, errors.New("failed to verify ID token")
	}
	userInfoClaims := &userInfoClaims{}
	if err := idToken.Claims(&userInfoClaims); err != nil {
		return nil, errors.New("failed to extract claims")
	}
	return userInfoClaims, nil
}
