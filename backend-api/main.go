package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"net/http/httputil"
	"fmt"
	"golang.org/x/oauth2"
	"context"
	"encoding/json"
	"strconv"
	"time"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var oauthConfig oauth2.Config
var port string

func init() {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()

	tomlInBytes, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatal().AnErr("Failed reading config file", err)
	}

	config, err := toml.LoadBytes(tomlInBytes)
	if err != nil {
		log.Fatal().AnErr("Failed parsing toml file", err)
	}

	// Maybe server config
	port = strconv.FormatInt(config.Get("env.dev.port").(int64), 10)

	oauthConfig = oauth2.Config{
		ClientID: config.Get("env.dev.as.okta.client_id").(string),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL: config.Get("env.dev.as.okta.callback").(string),
		Endpoint: oauth2.Endpoint {TokenURL: config.Get("env.dev.as.okta.token_endpoint").(string)},
	} //ConfigFromJSONの ConfigFromJSONが参考になる
}

func main() {
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})

	r := mux.NewRouter()
	r.HandleFunc("/token", dumpRequest(handleTokenRequest)).Methods(http.MethodPost, http.MethodOptions)
	srv := &http.Server{
		Handler: handlers.CORS(allowedOrigins, allowedHeaders)(r),
		Addr:    "localhost:" + port,
	}

	//server := http.Server{Addr: "localhost" + ":" + port}
	//http.HandleFunc("/token", dumpRequest(handleTokenRequest)) //limited to Okta for now
	log.Fatal().Err(srv.ListenAndServe())
}

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	//w.Header().Set("Access-Control-Allow-Headers","Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	defer r.Body.Close()
	var b struct {
		AuthzCode string `json:"authz_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		return
	}

	fmt.Println(b.AuthzCode)
	fmt.Println(oauthConfig.ClientID)
	fmt.Println(oauthConfig.ClientSecret)
	fmt.Println(oauthConfig.RedirectURL)
	token, err := oauthConfig.Exchange(context.Background(), b.AuthzCode)
	if err != nil {
		//var tokenResponseError TokenResponseError
		//if json.Unmarshal([]byte(err.Error()), &tokenResponseError) != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	fmt.Fprintln(w, "failed parsing error in token response")
		//	return
		//}
		w.WriteHeader(http.StatusBadRequest)
		//json.NewEncoder(w).Encode(tokenResponseError)

		//oauth2: cannot fetch token: 401 Unauthorized
		//Response: {"error":"invalid_client","error_description":"Client authentication failed. Either the client or the client credentials are invalid."}
		fmt.Fprintln(w, err)
		return
	}

	fmt.Println("Refresh Token: "+ token.RefreshToken)

	tokenForFront := &struct {
		AccessToken  string    `json:"access_token"`
		TokenType    string    `json:"token_type"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		Expiry       time.Time `json:"expiry"`
		//TODO Implement for OIDC
		//RefreshToken string    `json:"refresh_token"` //this is actually an option
		//Scope       string `json:"scope"`
		//IDToken     string `json:"id_token"`
		//IdToken    string      `json:"id_token"`
		//Scope      string      `json:"scope"`
	}{
		AccessToken: token.AccessToken,
		TokenType: token.TokenType,
		Expiry: token.Expiry,
	}

	// They are required in https://tools.ietf.org/html/rfc6749#section-5.1
	// Assuming Front-Channel would be distributed by some kind of proxy
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenForFront)
}

func dumpRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(requestDump) + "\n")
		next.ServeHTTP(w, r)
	}
}

// TokenResponseError is https://tools.ietf.org/html/rfc6749#section-5.2
type TokenResponseError struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorUri         string `json:"error_uri,omitempty"`
}