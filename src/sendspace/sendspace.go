package sendspace

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"

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
	Key string `xml:"session_key"`
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
