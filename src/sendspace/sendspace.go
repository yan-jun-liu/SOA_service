package sendspace

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"../common"
)

type Sendspace struct {
	Token   SendspaceSessionToken
	Session SendspaceSessionKey
}

type SendspaceSessionToken struct {
	Result string `xml:"result"`
	Token  string `xml:"token"`
}

type SendspaceSessionKey struct {
	Status string `xml:"status,attr"`
	Error  Error  `xml:"error"`
	Key    string `xml:"session_key"`
}

type Error struct {
	Code string `xml:"code,attr"`
	Text string `xml:"text,attr"`
}

type SendspaceFolders struct {
	FolderList []SendspaceFolder `xml:"folder"`
	FileList   []SendspaceFile   `xml:"file"`
}

type SendspaceFolder struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	ParentId string `xml:"parent_folder_id,attr"`
}

type SendspaceFile struct {
	ID           string `xml:"id,attr"`
	Name         string `xml:"name,attr"`
	FolderId     string `xml:"folder_id,attr"`
	DownloadLink string `xml:"download_page_url"`
}
type XMLUploads struct {
	XMLName xml.Name `xml:"result"`
	Upload  Upload   `xml:"upload"`
}

type Upload struct {
	XMLName          xml.Name `xml:"upload"`
	URL              string   `xml:"url,attr"`
	ProgressURL      string   `xml:"progress_url,attr"`
	UploadIdentifier string   `xml:"upload_identifier,attr"`
	ExtraInfo        string   `xml:"extra_info,attr"`
}

func (s *Sendspace) RetrieveSendspaceToken() error {
	sendspaceKey := os.Getenv("SENDSPACE_KEY")
	resp, err := http.Get("http://api.sendspace.com/rest/?method=auth.createtoken&api_key=" + sendspaceKey + "&api_version=1.0&response_format=xml&app_version=0.1")
	body, err := ioutil.ReadAll(resp.Body)
	err = xml.Unmarshal(body, &s.Token)
	common.ErrorHandler("Read session token from sendspace went wrong: ", err)
	return err
}

func (s *Sendspace) RetrieveSessionKey(user common.User) error {
	md5_password := md5.Sum([]byte(user.Password))
	hash_password := hex.EncodeToString(md5_password[:])
	//lowercase(md5(token + lowercase(md5(password))))
	md5_token_password := md5.Sum([]byte(s.Token.Token + hash_password))
	hash_token_password := hex.EncodeToString(md5_token_password[:])
	resp, err := http.Get("http://api.sendspace.com/rest/?method=auth.login&token=" + s.Token.Token + "&user_name=" + user.Email + "&tokened_password=" + hash_token_password)
	body, err := ioutil.ReadAll(resp.Body)
	err = xml.Unmarshal(body, &s.Session)
	common.ErrorHandler("Read session key from sendspace went wrong: ", err)
	return err
}

func (s *Sendspace) RetrieveFolderTree(parent_folder *common.Folder) {
	resp, err := http.Get("http://api.sendspace.com/rest/?method=folders.getcontents&session_key=" + s.Session.Key + "&folder_id=" + parent_folder.Key)
	body, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))
	var folders SendspaceFolders
	err = xml.Unmarshal(body, &folders)
	common.ErrorHandler("Read folders from sendspace went wrong: ", err)
	files := []common.File{}
	for _, file := range folders.FileList {
		c_file := common.File{Key: file.ID, Name: file.Name, Branch: parent_folder.Branch + "/" + file.Name, DownloadLink: file.DownloadLink}
		files = append(files, c_file)
	}
	parent_folder.SetFiles(files)

	child_folders := []common.Folder{}
	for _, folder := range folders.FolderList {
		//log.Println("Retrieved folder ID:" + folder.ID + ",Name" + folder.Name + ", Parent_ID" + folder.ParentId)
		child_folder := common.Folder{Key: folder.ID, Name: folder.Name, Branch: parent_folder.Branch + "/" + folder.Name}
		s.RetrieveFolderTree(&child_folder)
		child_folders = append(child_folders, child_folder)
	}
	parent_folder.SetFolders(child_folders)
}

func (s *Sendspace) UploadFile(uploadFile common.UploadFile) <-chan string {
	channel := make(chan string)

	log.Println("Upload of file " + uploadFile.Name + " into folder " + uploadFile.TargetFolderKey + " started")

	// upload file
	go func() {
		//Get sendspace upload info
		resp, err := http.Get("http://api.sendspace.com/rest/?method=upload.getinfo&session_key=" + s.Session.Key + "&speed_limit=0")
		common.ErrorHandler("Get upload info from sendspace went wrong: ", err)
		body, err := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		var XML_Upload XMLUploads
		err = xml.Unmarshal(body, &XML_Upload)
		common.ErrorHandler("Read Upload info from sendspace went wrong: ", err)

		//Download from mediafire
		resp, err = http.Get(uploadFile.DownloadLink)
		common.ErrorHandler("Download file from mediafire went wrong file name: "+uploadFile.Name, err)
		//Prepare for upload to sendspace
		reader := resp.Body
		var buf bytes.Buffer
		mwriter := multipart.NewWriter(&buf)
		w, err := mwriter.CreateFormFile("userfile", uploadFile.Name)
		if err != nil {
			return
		}
		io.Copy(w, reader)

		w, err = mwriter.CreateFormField("extra_info")
		if err != nil {
			log.Println(err)
		}
		_, err = w.Write([]byte(XML_Upload.Upload.ExtraInfo))
		if err != nil {
			log.Println(err)
		}

		w, err = mwriter.CreateFormField("FOLDER_ID")
		if err != nil {
			log.Println(err)
		}
		_, err = w.Write([]byte(uploadFile.TargetFolderKey))
		if err != nil {
			log.Println(err)
		}

		mwriter.Close()
		// construct url for http.Post with arguments from previous call
		req, err := http.NewRequest(http.MethodPost, XML_Upload.Upload.URL, &buf)
		if err != nil {
			log.Println(err)
			channel <- "Fail to upload file:" + uploadFile.Name
			return
		}
		req.Header.Add("Content-Type", mwriter.FormDataContentType())
		// Submit the request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			channel <- "Fail to upload file:" + uploadFile.Name
			return
		}
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			channel <- "Fail to upload file:" + uploadFile.Name
			return
		}

		// Check the response
		if strings.Contains(string(body), "upload_status=ok") {
			channel <- "Successfully uploaded:" + uploadFile.Name
		} else {
			channel <- "Fail to upload:" + uploadFile.Name
		}
	}()

	return channel
}

func (s *Sendspace) CreateMissingFolders(origin *common.Folder, target *common.Folder) {
	for _, originFolder := range origin.Folders {
		folderExists := false
		for index, targetFolder := range target.Folders {
			if targetFolder.Name == originFolder.Name {
				log.Println("folder " + targetFolder.Branch + " exists")
				folderExists = true
				s.CreateMissingFolders(&originFolder, &target.Folders[index])
			}
		}

		if !folderExists {
			resp, err := http.Get("http://api.sendspace.com/rest/?method=folders.create&session_key=" + s.Session.Key + "&name=" + originFolder.Name + "&shared=0&parent_folder_id=" + target.Key)
			common.ErrorHandler("Create folder sendspace went wrong: ", err)
			body, err := ioutil.ReadAll(resp.Body)
			var folders SendspaceFolders
			err = xml.Unmarshal(body, &folders)
			common.ErrorHandler("Convert create folders xml from sendspace went wrong: ", err)
			child_folders := target.Folders
			for _, folder := range folders.FolderList {
				child_folder := common.Folder{Key: folder.ID, Name: folder.Name, Branch: target.Branch + "/" + folder.Name}
				s.CreateMissingFolders(&originFolder, &child_folder)
				child_folders = append(child_folders, child_folder)
			}
			target.SetFolders(child_folders)
		}
	}
}
