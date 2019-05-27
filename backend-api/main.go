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
	"time"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var oauthConfig oauth2.Config
var port string
var url string

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
	port = "9000"//strconv.FormatInt(config.Get("env.dev.port").(int64), 10)
	url = "localhost" //config.Get("env.dev.url").(string)

	oauthConfig = oauth2.Config{
		ClientID: config.Get("env.dev.client_id").(string), //"0oakuhp8brWUfRhGI0h7",
		ClientSecret: os.Getenv("CLIENT_SECRET"),//"HNhG1RVIPkqMyZ6PcLR7Ktoxs0geaWoEETRSSy25",
		RedirectURL: config.Get("env.dev.front_channel_url").(string), //"http://localhost:3000/callback",
		Endpoint: oauth2.Endpoint {TokenURL: config.Get("env.dev.token_endpoint.okta").(string)},
	} //ConfigFromJSONの ConfigFromJSONが参考になる
}

func main() {
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})

	r := mux.NewRouter()
	r.HandleFunc("/token", dumpRequest(handleTokenRequest)).Methods(http.MethodPost, http.MethodOptions)
	srv := &http.Server{
		Handler: handlers.CORS(allowedOrigins, allowedHeaders)(r),
		Addr:    url + ":" + port,
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

	token, err := oauthConfig.Exchange(context.Background(), b.AuthzCode)
	if err != nil {
		// TODO This needs to be improved
		// https://tools.ietf.org/html/rfc6749#section-5.2
		var tokenResponseError TokenResponseError
		if json.Unmarshal([]byte(err.Error()), &tokenResponseError) != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "failed parsing error in token response")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, fmt.Sprintf("failed token request: %v", err.Error()))
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
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}