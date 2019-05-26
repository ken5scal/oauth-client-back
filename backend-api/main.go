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
	//"gopkg.in/square/go-jose.v2/json"
	"encoding/json"
)

func init() {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()
}

func main() {
	server := http.Server{Addr: "localhost:9000"}
	http.HandleFunc("/token", dumpRequest(handleTokenRequest))
	//http.HandleFunc("/token", handleTokenRequest)
	log.Fatal().Err(server.ListenAndServe())
}

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers","Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	oauthConfig := oauth2.Config{
		ClientID: "0oakuhp8brWUfRhGI0h7",
		ClientSecret: "HNhG1RVIPkqMyZ6PcLR7Ktoxs0geaWoEETRSSy25",
		RedirectURL: "http://localhost:3000/callback",
		Endpoint: oauth2.Endpoint {TokenURL: "https://dev-991803.oktapreview.com/oauth2/default/v1/token"},
	} //ConfigFromJSONの ConfigFromJSONが参考になる

	defer r.Body.Close()
	//body, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write([]byte(err.Error()))
	//	return
	//}

	var b struct {
		AuthzCode string `json:"authz_code"`
	}

	//if err = json.Unmarshal(body, &b); err != nil {
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), b.AuthzCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, fmt.Sprintf("failed token request: %v", err.Error()))
		return
	}
		fmt.Println(token.AccessToken)
		fmt.Println(token.Expiry)
		fmt.Println(token.RefreshToken)
		fmt.Println(token.TokenType)
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(token)
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
