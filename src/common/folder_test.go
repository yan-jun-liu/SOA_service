package common

import (
	"fmt"
	"log"
	"testing"
)

var originRoot = Folder{
	Files: []File{
		File{Branch: "/file1", FolderBranch: "", DownloadLink: "link1"},
		File{Branch: "/file2", FolderBranch: "", DownloadLink: "link2"},
	},
	Folders: []Folder{
		Folder{
			Branch: "/folder1",
			Files: []File{
				File{Branch: "/folder1/file1", FolderBranch: "/folder1", DownloadLink: "link3"},
				File{Branch: "/folder1/file2", FolderBranch: "/folder1", DownloadLink: "link4"},
			},
			Folders: []Folder{
				Folder{
					Branch: "/folder1/folder3",
					Files: []File{
						File{Branch: "/folder1/folder3/file5", FolderBranch: "/folder1/folder3", DownloadLink: "link5"},
					},
					Folders: []Folder{
						Folder{
							Branch: "/folder1/folder3/folder5",
							Files: []File{
								File{Branch: "/folder1/folder3/folder6/file8", FolderBranch: "/folder1/folder3/folder5", DownloadLink: "link8"},
							},
						},
					},
				},
			},
		},
		Folder{
			Branch: "/folder2",
			Files: []File{
				File{Branch: "/folder2/file6", FolderBranch: "/folder2", DownloadLink: "link6"},
			},
			Folders: []Folder{
				Folder{
					Branch: "/folder2/folder4",
					Files: []File{
						File{Branch: "/folder1/folder4/file7", FolderBranch: "/folder2/folder4", DownloadLink: "link7"},
					},
				},
			},
		},
	},
}

var targetRoot = Folder{
	Key:    "0",
	Branch: "",
	Files: []File{
		File{Branch: "/file1"},
	},
	Folders: []Folder{
		Folder{
			Key:    "1",
			Branch: "/folder1",
			Files: []File{
				File{Branch: "/folder1/file1"},
				File{Branch: "/folder1/file3"},
			},
			Folders: []Folder{
				Folder{
					Key:    "3",
					Branch: "/folder1/folder3",
					Folders: []Folder{
						Folder{
							Key:    "4",
							Branch: "/folder1/folder3/folder4",
							Folders: []Folder{
								Folder{
									Key:    "6",
									Branch: "/folder1/folder3/folder4/folder6",
								},
							},
						},
						Folder{
							Key:    "5",
							Branch: "/folder1/folder3/folder5",
						},
					},
				},
			},
		},
		Folder{
			Key:    "2",
			Branch: "/folder2",
			Folders: []Folder{
				Folder{
					Key:    "6",
					Branch: "/folder2/folder6",
				},
			},
		},
	},
}

func TestGetFolderKey(t *testing.T) {
	if ok, val := GetFolderKey(&targetRoot, "/folder1"); !ok || val != "1" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder2"); !ok || val != "2" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder2/folder6"); !ok || val != "6" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder1/folder3"); !ok || val != "3" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder1/folder3/folder4"); !ok || val != "4" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder1/folder3/folder5"); !ok || val != "5" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder1/folder3/folder4/folder6"); !ok || val != "6" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, "/folder1/folder3/folder_none"); ok || val != "" {
		t.Fatal("incorrect folder key: " + val)
	}

	if ok, val := GetFolderKey(&targetRoot, ""); !ok || val != "0" {
		t.Fatal("incorrect folder key: " + val)
	}
}

func TestContains(t *testing.T) {
	expected := []File{
		File{Branch: "/file2", DownloadLink: "link2"},
		File{Branch: "/folder1/file2", DownloadLink: "link4"},
		File{Branch: "/folder1/folder3/file5", DownloadLink: "link5"},
		File{Branch: "/folder2/file6", DownloadLink: "link6"},
		File{Branch: "/folder1/folder4/file7", DownloadLink: "link7"},
		File{Branch: "/folder1/folder3/folder6/file8", DownloadLink: "link8"},
	}

	testFileTrue := File{Branch: "/file2", DownloadLink: "link2"}

	testFileFalse := File{Branch: "/file3", DownloadLink: "link3"}

	if !Contains(expected, testFileTrue) {
		t.Fatal("incorrect file " + testFileTrue.Branch)
	}

	if Contains(expected, testFileFalse) {
		t.Fatal("incorrect file " + testFileFalse.Branch)
	}
}

func TestListAllFilesFromTree(t *testing.T) {
	var expectedFiles = []File{
		File{Branch: "/file1"},
		File{Branch: "/folder1/file1"},
		File{Branch: "/folder1/file3"},
	}
	resultFiles := ListAllFilesFromTree(targetRoot)
	if len(resultFiles) != len(expectedFiles) {
		for _, expect := range expectedFiles {
			log.Println("incorrect expect file: " + expect.Branch)
		}
		for _, re := range resultFiles {
			log.Println("incorrect result file: " + re.Branch)
		}
		t.Fatal("len is resultFiles:" + fmt.Sprintf("%v", len(resultFiles)) + ",expectedFiles:" + fmt.Sprintf("%v", len(expectedFiles)))
	}
	for _, file := range resultFiles {
		if !Contains(expectedFiles, file) {
			t.Fatal("incorrect file: " + file.Branch)
		}
	}

	var originExpectedFiles = []File{
		File{Branch: "/file1", DownloadLink: "link1"},
		File{Branch: "/file2", DownloadLink: "link2"},
		File{Branch: "/folder1/file1", DownloadLink: "link3"},
		File{Branch: "/folder1/file2", DownloadLink: "link4"},
		File{Branch: "/folder1/folder3/file5", DownloadLink: "link5"},
		File{Branch: "/folder2/file6", DownloadLink: "link6"},
		File{Branch: "/folder1/folder4/file7", DownloadLink: "link7"},
		File{Branch: "/folder1/folder3/folder6/file8", DownloadLink: "link8"},
	}
	originResultFiles := ListAllFilesFromTree(originRoot)

	if len(originResultFiles) != len(originExpectedFiles) {
		for _, expect := range originExpectedFiles {
			log.Println("incorrect expect file: " + expect.Branch)
		}
		for _, re := range originResultFiles {
			log.Println("incorrect result file: " + re.Branch)
		}
		t.Fatal("len is originResultFiles:" + fmt.Sprintf("%v", len(originResultFiles)) + ",originExpectedFiles:" + fmt.Sprintf("%v", len(originExpectedFiles)))
	}
	for _, file := range originResultFiles {
		if !Contains(originExpectedFiles, file) {
			t.Fatal("incorrect file: " + file.Branch)
		}
	}
}

func TestListMissingFiles(t *testing.T) {
	expected := []UploadFile{
		UploadFile{Branch: "/file2", DownloadLink: "link2", TargetFolderKey: "0"},
		UploadFile{Branch: "/folder1/file2", DownloadLink: "link2", TargetFolderKey: "1"},
		UploadFile{Branch: "/folder1/folder3/file5", DownloadLink: "link5", TargetFolderKey: "3"},
		UploadFile{Branch: "/folder1/folder3/folder6/file8", DownloadLink: "link8", TargetFolderKey: "5"},
		UploadFile{Branch: "/folder2/file6", DownloadLink: "link6", TargetFolderKey: "2"},
	}

	result := ListMissingFiles(originRoot, targetRoot)

	if len(result) != len(expected) {
		for _, expect := range expected {
			log.Println("expect file: " + expect.Branch)
		}
		for _, re := range result {
			log.Println("result file: " + re.Branch)
		}
		t.Fatal("len is result:" + fmt.Sprintf("%v", len(result)) + ",expect:" + fmt.Sprintf("%v", len(expected)))
	}

	for _, file := range result {
		if !ContainesUploadFile(expected, file) {
			for _, expect := range expected {
				log.Println("incorrect expect file: " + expect.Branch + ", Target folder key:" + expect.TargetFolderKey)
			}
			t.Fatal("incorrect result file: " + file.Branch)
		}
	}
}
