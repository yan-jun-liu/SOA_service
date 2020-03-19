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
	sendspace := sendspace.Sendspace{}
	err = sendspace.RetrieveSendspaceToken()
	err = sendspace.RetrieveSessionKey(user)
	if err != nil || strings.Contains(sendspace.Token.Result, "fail") {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Login to sendspace failed!" + ".Sendspace result:" + sendspace.Token.Result))
		log.Println("Sendspace fail:" + sendspace.Token.Result)
		return
	}

	// get mediafire session token
	mediafire := mediafire.Mediafire{}
	err = mediafire.RetrieveMediafireToken(user)
	if err != nil || mediafire.Token.Result == "Error" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("400 - Login to mediafire failed!" + ". Mediafire result:" + mediafire.Token.Result))
		log.Println("Mediafire fail:" + mediafire.Token.Result)
	} else {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("200 - Login successful and start processing!" + ".Sendspace result:" + sendspace.Token.Result + ". Mediafire result:" + mediafire.Token.Result))
		log.Println("Sendspace successful:" + sendspace.Token.Result)
		log.Println("Mediafire successful:" + mediafire.Token.Result)
	}
	//Get mediafire folder tree
	m_common_root_folder := mediafire.ReadRootMediafireFolderTree()
	common.ReadFolderTree(&m_common_root_folder)

	//Get sendspace folder tree
	sendspace_root_folder := common.Folder{Key: "0"}
	sendspace.RetrieveFolderTree(&sendspace_root_folder)
	common.ReadFolderTree(&sendspace_root_folder)
}

func main() {
	os.Setenv("SENDSPACE_KEY", "")
	os.Setenv("MEDIAFIRE_API_KEY", "")
	os.Setenv("MEDIAFIRE_API_ID", "")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", login)

	log.Println("Starting server on :8081...")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
