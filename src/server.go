package main

import (
	"encoding/json"
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
	// get sendspace session token
	var sendspace_token sendspace.SendspaceSessionToken
	sendspace_token, err = sendspace.RetrieveSendspaceToken()
	if err != nil || strings.Contains(sendspace_token.Result, "fail") {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Login to sendspace failed!" + ".Sendspace result:" + sendspace_token.Result))
		log.Println("Sendspace fail:" + sendspace_token.Result)
		return
	}
	log.Println(sendspace_token.Token)

	// get mediafire session token
	mediafire_token, err := mediafire.RetrieveMediafireToken(user)
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
	m_common_root_folder := mediafire.ReadRootMediafireFolderTree(mediafire_token)
	common.ReadFolderTree(&m_common_root_folder)
	for _, file := range m_common_root_folder.Files {
		log.Println("readFolderTree File in " + file.Name + ": " + file.Name + ",Link:" + file.DownloadLink)
	}
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
