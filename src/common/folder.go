package common

import (
	"fmt"
	"log"
)

type Folder struct {
	Key     string
	Name    string
	Branch  string
	Folders []Folder
	Files   []File
}

func (t *Folder) SetFolders(folders []Folder) {
	t.Folders = folders
}

func (t *Folder) SetFiles(files []File) {
	t.Files = files
}

type File struct {
	Key          string
	Name         string
	Branch       string
	FolderBranch string
	DownloadLink string
}

type UploadFile struct {
	Name            string
	Branch          string
	DownloadLink    string
	TargetFolderKey string
}

//Find missing file in target for streaming later
func ListMissingFiles(origin Folder, target Folder) []UploadFile {
	files := []UploadFile{}
	origin_files := ListAllFilesFromTree(origin)
	target_files := ListAllFilesFromTree(target)
	for _, file := range origin_files {
		// Must find the parent target folder key after run create all missing folders
		ok, val := GetFolderKey(&target, file.FolderBranch)
		if !Contains(target_files, file) {
			if !ok {
				log.Println("Error finding branch " + file.FolderBranch + " from sendspace")
				break
			}
			upload_file := UploadFile{Name: file.Name, Branch: file.Branch, DownloadLink: file.DownloadLink, TargetFolderKey: val}
			files = append(files, upload_file)
		}
	}
	return files
}

func ListAllFilesFromTree(folder Folder) []File {
	files := []File{}
	for _, file := range folder.Files {
		files = append(files, file)
	}
	for _, child_folder := range folder.Folders {
		files = append(files, ListAllFilesFromTree(child_folder)...)
	}
	return files
}

func Contains(arr []File, file File) bool {
	for _, a := range arr {
		if a.Branch == file.Branch {
			return true
		}
	}
	return false
}

func ContainesUploadFile(arr []UploadFile, file UploadFile) bool {
	for _, a := range arr {
		if a.Branch == file.Branch {
			return true
		}
	}
	return false
}

func GetFolderKey(folder *Folder, folder_branch string) (bool, string) {
	//Root folder ID is 0 in sendspace
	if folder_branch == "" {
		return true, "0"
	} else {
		result := ""
		ok := false
		for _, child_folder := range folder.Folders {
			if child_folder.Branch == folder_branch {
				return true, child_folder.Key
			}
			ok, result = GetFolderKey(&child_folder, folder_branch)
			//Break when find the folder key in the children nodes
			if ok {
				break
			}
		}
		return ok, result
	}
}

// Read folder for testing purpose
func ReadFolderTree(parent_folder *Folder) {
	log.Println("readFolderTree parent Folder name: " + parent_folder.Name + ", Branch: " + parent_folder.Branch + ",size:" + fmt.Sprintf("%v", len(parent_folder.Folders)))
	for _, child_folder := range parent_folder.Folders {
		log.Println("readFolderTree child Folder name: " + child_folder.Name + ", Branch: " + child_folder.Branch + ", len: " + fmt.Sprintf("%v", len(child_folder.Folders)))
		log.Println("readFolderTree Files len: " + child_folder.Name + "," + fmt.Sprintf("%v", len(child_folder.Files)))
		if len(child_folder.Files) != 0 {
			for _, file := range child_folder.Files {
				//log.Println("readFolderTree File in " + child_folder.Name + ": " + file.Name + ",Link:" + file.DownloadLink + ", Branch: " + file.Branch)
				log.Println("Branch: " + file.Branch)
			}
		}
		for {
			//log.Println("folder name:" + child_folder.Name + ", len(child_folder.Folders) == 0 || index > 200:" + fmt.Sprintf("%t", len(child_folder.Folders) == 0 || index > 20))
			if len(child_folder.Folders) == 0 {
				//	log.Println("Run here")
				break
			} else {
				//	log.Println("Print i:" + fmt.Sprintf("%v", index) + ",Passing:" + child_folder.Name)
				ReadFolderTree(&child_folder)
				break
			}
		}
	}
	for _, file := range parent_folder.Files {
		//log.Println("readFolderTree File in " + parent_folder.Name + ": " + file.Name + ",Link:" + file.DownloadLink + ", Branch: " + file.Branch)
		log.Println("Branch: " + file.Branch)
	}
}
