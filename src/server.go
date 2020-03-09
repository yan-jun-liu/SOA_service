package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"./common"
	"./mediafire"
	"./sendspace"
)

func login(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	log.Println(string(body))
	var user common.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Email and password can't be passed"))
	}
	log.Println(user.Email)

	var sendspace_token sendspace.SendspaceSessionToken
	sendspace_token, err = getSendspaceToken()
	if err != nil || strings.Contains(sendspace_token.Result, "fail") {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Login to sendspace failed!" + ".Sendspace result:" + sendspace_token.Result))
		log.Println("Sendspace fail:" + sendspace_token.Result)
		return
	}
	log.Println(sendspace_token.Token)

	var mediafire_token mediafire.MediafireSessionToken
	mediafire_token, err = getMediafireToken(user)
	log.Println(mediafire_token.Token + ".Time:" + mediafire_token.Time)
	if err != nil || mediafire_token.Result == "Error" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Login to mediafire failed!" + ". Mediafire result:" + mediafire_token.Result))
		log.Println("Mediafire fail:" + mediafire_token.Result)
	} else {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("200 - Login successful and start processing!" + ".Sendspace result:" + sendspace_token.Result + ". Mediafire result:" + mediafire_token.Result))
		log.Println("Sendspace successful:" + sendspace_token.Result)
		log.Println("Mediafire successful:" + mediafire_token.Result)
	}
}

func getSendspaceToken() (sendspace.SendspaceSessionToken, error) {
	var sendspaceKey = os.Getenv("SENDSPACE_KEY")
	log.Println("Sendspace API key env: " + sendspaceKey)
	resp, err := http.Get("http://api.sendspace.com/rest/?method=auth.createtoken&api_key=" + sendspaceKey + "&api_version=1.0&response_format=xml&app_version=0.1")
	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	var token sendspace.SendspaceSessionToken
	err = xml.Unmarshal(body, &token)
	return token, err
}

func getMediafireToken(user common.User) (mediafire.MediafireSessionToken, error) {
	var mediafireAPIKey = os.Getenv("MEDIAFIRE_API_KEY")
	var applicationID = os.Getenv("API_ID")
	log.Println("Mediafire API key env: " + mediafireAPIKey)
	log.Println("Application ID:" + applicationID)
	var hashValue = user.Email + user.Password + applicationID + mediafireAPIKey
	h := sha1.New()
	h.Write([]byte(hashValue))
	var signature = hex.EncodeToString(h.Sum(nil))
	resp, err := http.Get("https://www.mediafire.com/api/1.1/user/get_session_token.php?email=" + user.Email + "&password=" + user.Password + "&application_id=" + applicationID + "&signature=" + signature + "&token_version=2")
	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	var token mediafire.MediafireSessionToken
	err = xml.Unmarshal(body, &token)
	return token, err
}

func main() {
	os.Setenv("SENDSPACE_KEY", "")
	os.Setenv("MEDIAFIRE_API_KEY", "")
	os.Setenv("API_ID", "")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", login)

	log.Println("Starting server on :8081...")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
