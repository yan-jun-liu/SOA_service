package common

import (
	"fmt"
	"log"
)

type Folder struct {
	Key      string
	Name     string
	IsParent bool
	Branch   string
	Folders  []Folder
	Files    []File
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
	DownloadLink string
}

var index = 0

func ReadFolderTree(parent_folder *Folder) {
	log.Println("readFolderTree parent Folder name: " + parent_folder.Name + ", Branch: " + parent_folder.Branch + ",size:" + fmt.Sprintf("%v", len(parent_folder.Folders)))
	for _, child_folder := range parent_folder.Folders {
		log.Println("readFolderTree child Folder name: " + child_folder.Name + ", Branch: " + child_folder.Branch + ", len: " + fmt.Sprintf("%v", len(child_folder.Folders)))
		//log.Println("readFolderTree Files len: " + child_folder.Name + "," + fmt.Sprintf("%v", len(child_folder.Files)))
		if len(child_folder.Files) != 0 {
			for _, file := range child_folder.Files {
				log.Println("readFolderTree File in " + child_folder.Name + ": " + file.Name + ",Link:" + file.DownloadLink + ", Branch: " + file.Branch)
			}
		}
		for {
			//log.Println("folder name:" + child_folder.Name + ", len(child_folder.Folders) == 0 || index > 200:" + fmt.Sprintf("%t", len(child_folder.Folders) == 0 || index > 20))
			if len(child_folder.Folders) == 0 || index > 200 {
				//	log.Println("Run here")
				break
			} else {
				index++
				//	log.Println("Print i:" + fmt.Sprintf("%v", index) + ",Passing:" + child_folder.Name)
				ReadFolderTree(&child_folder)
				break
			}
		}
	}
	for _, file := range parent_folder.Files {
		log.Println("readFolderTree File in " + parent_folder.Name + ": " + file.Name + ",Link:" + file.DownloadLink + ", Branch: " + file.Branch)
	}
}
