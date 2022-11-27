package main

import (
	"fmt"
	"github-backup/pkg/git"
	"github-backup/pkg/objstorage"
	"github-backup/pkg/zippings"
	"os"
	"path/filepath"
	"sync"
)

var basedir = "/tmp/ghbackup"
var denylist []string

const MaxConcurrent = 5

func main() {
	bucketname := envOrDie("BUCKET_NAME")
	githubToken := envOrDie("GITHUB_TOKEN")

	repos := reposOrDie("navikt", githubToken)
	fmt.Printf("found %d repos\n", len(repos))

	workQueue := make(chan int, MaxConcurrent)
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for _, repo := range repos {
		r := repo
		workQueue <- 1
		go func() {
			err := cloneZipAndStoreInBucket(r.FullName, bucketname, githubToken)
			if err != nil {
				fmt.Printf("could'n clone '%s': %v", r.FullName, err)
			}
			<-workQueue
			wg.Done()
		}()
	}
	wg.Wait()
}

func cloneZipAndStoreInBucket(repo string, bucketname string, githubToken string) error {
	err := git.CloneRepo(basedir, repo, "NAVGitHubBackup", githubToken)
	if err != nil {
		return err
	}

	compressedFileName := objstorage.FilenameFor(repo)
	compressedFilePath := filepath.Join(basedir, compressedFileName)
	err = zippings.CompressIt(basedir, compressedFilePath, denylist)
	if err != nil {
		return err
	}

	file, err := os.Open(compressedFilePath)
	defer file.Close()
	if err != nil {
		return err
	}
	objstorage.CopyToBucket(file, bucketname)

	err = os.RemoveAll(filepath.Join(basedir, repo))
	if err != nil {
		return err
	}
	err = os.RemoveAll(compressedFilePath)
	if err != nil {
		return err
	}

	return nil
}

func envOrDie(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		fmt.Printf("unable to find env var '%s', I'm useless without it", name)
		os.Exit(1)
	}
	return value
}

func reposOrDie(org string, githubToken string) []git.Repo {
	repos, err := git.ReposFor("navikt", githubToken)
	if err != nil {
		fmt.Printf("couldn't get list of repos: %v\n", err)
		os.Exit(1)
	}
	return repos
}
