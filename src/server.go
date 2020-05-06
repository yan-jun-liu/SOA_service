package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"./common"
	"./mediafire"
	"./sendspace"
)

// UploadResponse for reading upload status response
type UploadResponse struct {
	Status          string
	Message         string
	UploadSucceeded []string
	UploadFailed    []string
}

//LoginFailedResponse for handling login failed response
type LoginFailedResponse struct {
	Status  string
	Message string
}

func login(rw http.ResponseWriter, req *http.Request) {
	enableCors(&rw)
	rw.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(req.Body)
	bodyString := string(body)
	log.Println("Received login info: " + bodyString)
	var user common.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		loginFailedResponse := LoginFailedResponse{"Fail", "Email and password can't be passed:" + bodyString}
		bytesMessage, err := json.Marshal(loginFailedResponse)
		if err != nil {
			log.Println("Json format email and password wrong:" + bodyString)
			return
		}
		log.Println("Received JSON format wrong:" + bodyString)
		rw.Write(bytesMessage)
		return
	}
	log.Println(user.Email)
	// get sendspace session token
	sendspace := sendspace.Sendspace{}
	err = sendspace.RetrieveSendspaceToken()
	if err != nil {
		log.Println(err)
		return
	}
	err = sendspace.RetrieveSessionKey(user)
	if err != nil || strings.Contains(sendspace.Session.Status, "fail") {
		rw.WriteHeader(http.StatusUnauthorized)
		loginFailedResponse := LoginFailedResponse{"Fail", "Login to sendspace failed!.Sendspace result:" + sendspace.Session.Error.Text}
		bytesMessage, err := json.Marshal(loginFailedResponse)
		if err != nil {
			return
		}
		rw.Write(bytesMessage)
		log.Println("Sendspace fail:" + sendspace.Token.Result)
		return
	}

	// get mediafire session token
	mediafire := mediafire.Mediafire{}
	err = mediafire.RetrieveMediafireToken(user)
	if err != nil || mediafire.Token.Result == "Error" {
		rw.WriteHeader(http.StatusUnauthorized)
		loginFailedResponse := LoginFailedResponse{"Fail", "Login to mediafire failed!" + ". Mediafire result:" + mediafire.Token.Result}
		bytesMessage, err := json.Marshal(loginFailedResponse)
		if err != nil {
			return
		}
		rw.Write(bytesMessage)
		log.Println("Mediafire fail:" + mediafire.Token.Result)
		return
	}

	//Get mediafire folder tree
	mediafireRootFolder := mediafire.ReadRootMediafireFolderTree()

	//Get sendspace folder tree
	sendspaceRootFolder := common.Folder{Key: "0"}
	sendspace.RetrieveFolderTree(&sendspaceRootFolder)
	//To upload the file, requires to give a parent folder ID, create all the missing folders before upload the files
	sendspace.CreateMissingFolders(&mediafireRootFolder, &sendspaceRootFolder)
	//Compare two folder trees and return list missing of files in upload format
	uploadFiles := common.ListMissingFiles(mediafireRootFolder, sendspaceRootFolder)
	var channels []<-chan string
	for _, file := range uploadFiles {
		log.Println("Branch: " + file.Branch + ", Download link:" + file.DownloadLink)
		channels = append(channels, sendspace.UploadFile(file))
		time.Sleep(time.Second)
	}

	success := true
	successFilesMessage := []string{}
	failedFilesNames := []string{}
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
			successFilesMessage = append(successFilesMessage, message)
		} else {
			if success {
				success = false
			}
			failedFilesNames = append(failedFilesNames, message)
		}
	}

	var uploadResponse UploadResponse
	if success {
		rw.WriteHeader(http.StatusOK)
		uploadResponse = UploadResponse{"200", "Uploaded all files succeeded", successFilesMessage, failedFilesNames}
	} else {
		rw.WriteHeader(http.StatusInternalServerError)
		uploadResponse = UploadResponse{"500", "Something went wrong during the upload", successFilesMessage, failedFilesNames}
	}

	bytesMessage, err := json.Marshal(uploadResponse)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(bytesMessage)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", login)

	log.Println("Starting server on 8080.....")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
