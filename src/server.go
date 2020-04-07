package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"./common"
	"./mediafire"
	"./sendspace"
)

type UploadResponse struct {
	Status          string
	Message         string
	UploadSucceeded []string
	UploadFailed    []string
}

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
	}

	//Get mediafire folder tree
	mediafire_root_folder := mediafire.ReadRootMediafireFolderTree()

	//Get sendspace folder tree
	sendspace_root_folder := common.Folder{Key: "0"}
	sendspace.RetrieveFolderTree(&sendspace_root_folder)
	//To upload the file, requires to give a parent folder ID, create all the missing folders before upload the files
	sendspace.CreateMissingFolders(&mediafire_root_folder, &sendspace_root_folder)
	//Compare two folder trees and return list missing of files in upload format
	upload_files := common.ListMissingFiles(mediafire_root_folder, sendspace_root_folder)
	var channels []<-chan string
	for _, file := range upload_files {
		log.Println("Branch: " + file.Branch + ", Download link:" + file.DownloadLink)
		channels = append(channels, sendspace.UploadFile(file))
		time.Sleep(time.Second)
	}

	success := true
	success_files_message := []string{}
	failed_files_names := []string{}
	// wait until uploads are all done
	for _, channel := range channels {
		result := <-channel
		results := strings.Split(result, ":")
		var message string
		if len(result) > 0 && len(results) > 1 {
			message = results[1]
		} else {
			success = false
			message = result + " parse went wrong, len:" + fmt.Sprintf("%v", len(result)) + ",results len:" + fmt.Sprintf("%v", len(results))
		}
		if strings.Contains(result, "Successfully") {
			// file upload success
			success_files_message = append(success_files_message, message)
		} else {
			if success {
				success = false
			}
			failed_files_names = append(failed_files_names, message)
		}
	}

	rw.Header().Set("Content-Type", "application/json")
	var upload_response UploadResponse
	if success {
		rw.WriteHeader(http.StatusOK)
		upload_response = UploadResponse{"200", "Uploaded all files succeeded", success_files_message, failed_files_names}
	} else {
		rw.WriteHeader(http.StatusInternalServerError)
		upload_response = UploadResponse{"500", "Something went wrong during the upload", success_files_message, failed_files_names}
	}

	bytes_message, err := json.Marshal(upload_response)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(bytes_message)
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
