package mediafire

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"../common"
)

type SessionToken struct {
	Action    string `xml:"action"`
	Token     string `xml:"session_token"`
	SecretKey string `xml:"secret_key"`
	Time      string `xml:"time"`
	Result    string `xml:"result"`
}

func (t *SessionToken) SetSecretKey(secret_key string) {
	t.SecretKey = secret_key
}

type FoldersContent struct {
	Action        string        `xml:"action"`
	FolderContent FolderContent `xml:"folder_content"`
	Result        string        `xml:"result"`
}

type FolderContent struct {
	ContentType string  `xml:"content_type"`
	Folders     Folders `xml:"folders"`
}

type Folders struct {
	XMLName    xml.Name `xml:"folders"`
	FolderList []Folder `xml:"folder"`
}

type Folder struct {
	XMLName     xml.Name `xml:"folder"`
	FolderKey   string   `xml:"folderkey"`
	Name        string   `xml:"name"`
	FileCount   int      `xml:"file_count"`
	FolderCount int      `xml:"folder_count"`
}

type FilesContent struct {
	Action      string      `xml:"action"`
	FileContent FileContent `xml:"folder_content"`
	Result      string      `xml:"result"`
}

type FileContent struct {
	ContentType string `xml:"content_type"`
	Files       Files  `xml:"files"`
}

type Files struct {
	XMLName  xml.Name `xml:"files"`
	FileList []File   `xml:"file"`
}

type File struct {
	XMLName  xml.Name `xml:"file"`
	FileKey  string   `xml:"quickkey"`
	FileName string   `xml:"filename"`
	Hash     string   `xml:"hash"`
	Link     Link     `xml:"links"`
}
type Link struct {
	XMLName      xml.Name `xml:"links"`
	View         string   `xml:"view"`
	DownloadLink string   `xml:"normal_download"`
}

func RetrieveFoldersFromMediafire(url string, signature string) FoldersContent {
	log.Println("Folder Request to http://www.mediafire.com" + url + "&signature=" + signature)
	resp, err := http.Get("http://www.mediafire.com" + url + "&signature=" + signature)
	common.ErrorHandler("Get folder content from mediafire failed", err)
	body, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))
	common.ErrorHandler("Read folder content from mediafire failed", err)
	var folder_content FoldersContent
	err = xml.Unmarshal(body, &folder_content)
	common.ErrorHandler("Conver folder content xml from mediafire failed", err)
	return folder_content
}

func RetrieveFilesFromMediafire(url string, signature string) FilesContent {
	log.Println("File Request to http://www.mediafire.com" + url + "&signature=" + signature)
	resp, err := http.Get("http://www.mediafire.com" + url + "&signature=" + signature)
	common.ErrorHandler("Get files content from mediafire failed", err)
	body, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))
	common.ErrorHandler("Read files content from mediafire failed", err)
	var files_content FilesContent
	err = xml.Unmarshal(body, &files_content)
	common.ErrorHandler("Conver files content xml from mediafire failed", err)
	return files_content
}

func MediafireFoldersContentMapper(mediafire_folders_content FoldersContent, parent_folder *common.Folder) {
	log.Println("mediafireFoldersContentMapper List folder size mediafire: " + fmt.Sprintf("%v", len(mediafire_folders_content.FolderContent.Folders.FolderList)))
	child_folders := []common.Folder{}
	for _, folder_content := range mediafire_folders_content.FolderContent.Folders.FolderList {
		child_folder := common.Folder{Key: folder_content.FolderKey, Name: folder_content.Name}
		child_folders = append(child_folders, child_folder)
		log.Println("mediafireFoldersContentMapper Folder Key: " + child_folder.Key + ",Name:" + child_folder.Name)
	}
	parent_folder.SetFolders(child_folders)
}

func MediafireFilesContentMapper(mediafire_files_content FilesContent, parent_folder *common.Folder) {
	log.Println("mediafireFilesContentMapper List file size mediafire: " + fmt.Sprintf("%v", len(mediafire_files_content.FileContent.Files.FileList)))
	files := []common.File{}
	for _, file := range mediafire_files_content.FileContent.Files.FileList {
		c_file := common.File{Key: file.FileKey, Name: file.FileName, DownloadLink: file.Link.DownloadLink}
		log.Println("mediafireFilesContentMapper File Key: " + c_file.Key + ",Name:" + c_file.Name + ",File Download Link:" + c_file.DownloadLink)
		files = append(files, c_file)
	}
	parent_folder.SetFiles(files)
}

func ListFoldersWithFilesFromMediafire(url string, mediafire_token *SessionToken, mediafire_folders_parent_content *FoldersContent, m_parent_folder *common.Folder) {
	child_folders := []common.Folder{}
	// Get mediafire folder information
	for _, folder := range mediafire_folders_parent_content.FolderContent.Folders.FolderList {
		log.Println("Folder Key: " + folder.FolderKey + ",Folder Name:" + folder.Name)
		child_folder := common.Folder{Key: folder.FolderKey, Name: folder.Name}

		if folder.FileCount != 0 {
			url_retrieve_files := "/api/1.1/folder/get_content.php?folder_key=" + folder.FolderKey + "&session_token=" + mediafire_token.Token + "&content_type=files"
			signature_mediafire, mediafire_token := CalculateMediafireSignature(url_retrieve_files, mediafire_token)
			log.Println("After set mediafire new key:" + mediafire_token.SecretKey)
			mediafire_file_content := RetrieveFilesFromMediafire(url_retrieve_files, signature_mediafire)
			MediafireFilesContentMapper(mediafire_file_content, &child_folder)
		}
		//Ignore empty folder
		if folder.FolderCount != 0 {
			url_retrieve_folder := "/api/1.1/folder/get_content.php?folder_key=" + folder.FolderKey + "&session_token=" + mediafire_token.Token + "&content_type=folders"
			signature_mediafire, mediafire_token := CalculateMediafireSignature(url_retrieve_folder, mediafire_token)
			log.Println("After set mediafire new key:" + mediafire_token.SecretKey)
			mediafire_child_folder_content := RetrieveFoldersFromMediafire(url_retrieve_folder, signature_mediafire)
			MediafireFoldersContentMapper(mediafire_child_folder_content, &child_folder)
			ListFoldersWithFilesFromMediafire(url_retrieve_folder, mediafire_token, &mediafire_child_folder_content, &child_folder)
		}
		if folder.FileCount != 0 || folder.FolderCount != 0 {
			child_folders = append(child_folders, child_folder)
		}
	}
	m_parent_folder.SetFolders(child_folders)
}

func ReadRootMediafireFolderTree(mediafire_token *SessionToken) common.Folder {
	m_common_root_folder := common.Folder{Name: "root", IsParent: true}
	//Init the parent folder call
	url_retrieve_parent_folder := "/api/1.1/folder/get_content.php?session_token=" + mediafire_token.Token + "&content_type=folders"
	signature_mediafire, mediafire_token := CalculateMediafireSignature(url_retrieve_parent_folder, mediafire_token)
	mediafire_root_folders_content := RetrieveFoldersFromMediafire(url_retrieve_parent_folder, signature_mediafire)
	ListFoldersWithFilesFromMediafire(url_retrieve_parent_folder, mediafire_token, &mediafire_root_folders_content, &m_common_root_folder)
	//Init the parent file call
	url_retrieve_root_files := "/api/1.1/folder/get_content.php?session_token=" + mediafire_token.Token + "&content_type=files"
	signature_mediafire_files, mediafire_token := CalculateMediafireSignature(url_retrieve_root_files, mediafire_token)
	mediafire_file_content := RetrieveFilesFromMediafire(url_retrieve_root_files, signature_mediafire_files)
	MediafireFilesContentMapper(mediafire_file_content, &m_common_root_folder)

	return m_common_root_folder
}

func CalculateMediafireSignature(url string, mediafire_token *SessionToken) (string, *SessionToken) {
	secret, _ := strconv.Atoi(mediafire_token.SecretKey)
	mod := secret % 256
	prefix := fmt.Sprintf("%v", mod) + mediafire_token.Time
	hash := md5.Sum([]byte(prefix + url))
	hashResult := hex.EncodeToString(hash[:])
	//set the new key for next call
	new_secret := fmt.Sprintf("%v", secret*16807%2147483647)
	mediafire_token.SetSecretKey(new_secret)
	return hashResult, mediafire_token
}

func RetrieveMediafireToken(user common.User) (*SessionToken, error) {
	mediafireAPIKey := os.Getenv("MEDIAFIRE_API_KEY")
	applicationID := os.Getenv("API_ID")
	log.Println("Mediafire API key env: " + mediafireAPIKey)
	log.Println("Application ID:" + applicationID)
	hashValue := user.Email + user.Password + applicationID + mediafireAPIKey
	h := sha1.New()
	h.Write([]byte(hashValue))
	signature := hex.EncodeToString(h.Sum(nil))
	resp, err := http.Get("https://www.mediafire.com/api/1.1/user/get_session_token.php?email=" + user.Email + "&password=" + user.Password + "&application_id=" + applicationID + "&signature=" + signature + "&token_version=2")
	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	var token SessionToken
	err = xml.Unmarshal(body, &token)
	return &token, err
}
