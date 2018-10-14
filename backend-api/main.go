package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type AuthServer struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
}

type Client struct {
	ClientId     string
	ClientSecret string
	RedirectURIs []string
	Scopes       []string
}

var state, scope, accessToken string
var port = 9000
var as *AuthServer
var client *Client

func init() {
	as = &AuthServer{
		AuthorizationEndpoint: "http://localhost:9001/authorize",
		TokenEndpoint:         "http://localhost:9001/token",
	}

	client = &Client{
		ClientId:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURIs: []string{"http://localhost:9000/callback"},
		Scopes:       []string{"foo", "bar"},
		//Scopes:       []string{"email", "profile", "openid"},
	}

	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()
}

func main() {
	server := http.Server{Addr: "localhost:9000"}
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/callback", callback)
	log.Fatal().Err(server.ListenAndServe())
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	endpoint := as.AuthorizationEndpoint

	state = uuid.NewV4().String()
	nonce := "this is nonce"

	params := url.Values{
		"client_id":     {client.ClientId},
		"response_type": {"code"},
		"scope":         {strings.Join(client.Scopes, " ")},
		"redirect_uri":  client.RedirectURIs,
		"state":         {state},
		"nonce":         {nonce},
	}

	if !strings.Contains(as.AuthorizationEndpoint, "?") {
		endpoint = endpoint + "?"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Redirect(w, r, endpoint+params.Encode(), http.StatusFound)
	log.Debug().Msg("redirect to AuthZ endpoint")
}

func callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	//// check state
	//// if state is set, but has different value, then, some one is stealing the code.
	//if state != "" && state != q.Get("key") {
	//	return errors.New("State is different")
	//}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Debug().Msg("Receiving a response from AuthZ endopoint")
	w.Write([]byte(q.Get("code")))
}