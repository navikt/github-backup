package main

import (
	"fmt"
	"github-backup/pkg/git"
	"github-backup/pkg/objstorage"
	"github-backup/pkg/zippings"
	"os"
	"path/filepath"
	"time"
)

var basedir = "/tmp/ghbackup"
var denylist = []string{".git/"}

func main() {
	compressedFileName := fmt.Sprintf("ghbackup_%s.tar.gz", time.Now().Format("2006-01-02T15:04:05-0700"))
	compressedFilePath := filepath.Join(basedir, compressedFileName)
	githubToken, found := os.LookupEnv("GITHUB_TOKEN")
	fmt.Println("retrieving list of repos from github")
	if !found {
		fmt.Println("'GITHUB_TOKEN' not found in env, I'm useless without it")
		os.Exit(1)
	}

	repos, err := git.ReposFor("navikt", githubToken)
	if err != nil {
		fmt.Printf("couldn't get list of repos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("found %d repos\n", len(repos))

	for _, repo := range repos {
		fmt.Printf("cloning %s\n", repo.FullName)
		err = git.CloneRepo(basedir, repo.FullName, "GitHubBackup", githubToken)
		if err != nil {
			fmt.Printf("unable to clone repo %s: %v\n", repo.FullName, err)
		}
	}

	err = zippings.CompressIt(basedir, compressedFilePath, denylist)
	if err != nil {
		fmt.Printf("unable to compress the cloned repos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote compressed file %s\n", compressedFilePath)

	file, err := os.Open(compressedFilePath)
	objstorage.CopyToBucket(file)
	file.Close()
}
