package config

type Config struct {
    KeycloakBaseURL string
    Realm           string
    ClientID        string
    RedirectURI     string
	ClientSecret string
}