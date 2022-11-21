package main

import (
	"fmt"
	"github-backup/pkg/git"
	"github-backup/pkg/objstorage"
	"github-backup/pkg/zippings"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var basedir = "/tmp/ghbackup"
var denylist []string

const MaxConcurrent = 10

func main() {
	compressedFileName := fmt.Sprintf("ghbackup_%s.tar.gz", time.Now().Format("2006-01-02T15:04:05-0700"))
	compressedFilePath := filepath.Join(basedir, compressedFileName)

	bucketName, found := os.LookupEnv("BUCKET_NAME")
	if !found {
		fmt.Println("'BUCKET_NAME' not found in env, I'm useless without it")
		os.Exit(1)
	}

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

	workQueue := make(chan git.Repo, MaxConcurrent)
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for _, repo := range repos {
		r := repo
		workQueue <- r
		go func() {
			fmt.Printf("cloning %s\n", r.FullName)
			err = git.CloneRepo(basedir, r.FullName, "GitHubBackup", githubToken)
			if err != nil {
				fmt.Printf("unable to clone repo %s: %v\n", r.FullName, err)
			}
			<-workQueue
			wg.Done()
		}()
	}
	wg.Wait()

	err = zippings.CompressIt(basedir, compressedFilePath, denylist)
	if err != nil {
		fmt.Printf("unable to compress the cloned repos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote compressed file %s\n", compressedFilePath)

	file, err := os.Open(compressedFilePath)
	if err != nil {
		fmt.Printf("unable to open file '%s' %v\n", compressedFilePath, err)
		os.Exit(1)
	}
	objstorage.CopyToBucket(file, bucketName)
	file.Close()
}
