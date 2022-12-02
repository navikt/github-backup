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

var basedir = filepath.Join(os.TempDir(), "ghbackup")

const MaxConcurrent = 3

func main() {
	bucketname := envOrDie("BUCKET_NAME")
	githubToken := envOrDie("GITHUB_TOKEN")

	repos := reposOrDie("navikt", githubToken)
	fmt.Printf("found %d repos\n", len(repos))

	workQueue := make(chan int, MaxConcurrent)
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for i, repo := range repos {
		fmt.Printf("processing repo %d of %d\n", i, len(repos))
		r := repo
		workQueue <- 1
		go func() {
			err := cloneZipAndStoreInBucket(r.FullName, bucketname, githubToken)
			if err != nil {
				fmt.Printf("failed to backup repo '%s': %v\n", r.FullName, err)
			}
			<-workQueue
			wg.Done()
		}()
	}
	wg.Wait()
}

func cloneZipAndStoreInBucket(repo string, bucketname string, githubToken string) error {
	compressedFileName := zippings.FilenameFor(repo)
	compressedFilePath := filepath.Join(basedir, compressedFileName)
	repodir := filepath.Join(basedir, repo)

	err := git.CloneRepo(basedir, repo, "NAVGitHubBackup", githubToken)
	if err != nil {
		rm([]string{repodir})
		return err
	}

	err = zippings.CompressIt(repodir, compressedFilePath)
	if err != nil {
		rm([]string{repodir, compressedFilePath})
		return err
	}

	file, err := os.Open(compressedFilePath)
	defer file.Close()
	if err != nil {
		rm([]string{repodir, compressedFilePath})
		return err
	}
	err = objstorage.CopyToBucket(file, bucketname)
	if err != nil {
		rm([]string{repodir, compressedFilePath})
		return err
	}

	rm([]string{repodir, compressedFilePath})

	return nil
}

func envOrDie(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		fmt.Printf("unable to find env var '%s', I'm useless without it\n", name)
		os.Exit(1)
	}
	return value
}

func reposOrDie(org string, githubToken string) []git.Repo {
	repos, err := git.ReposFor(org, githubToken)
	if err != nil {
		fmt.Printf("couldn't get list of repos: %v\n", err)
		os.Exit(1)
	}
	return repos
}

func rm(entries []string) {
	for _, f := range entries {
		fmt.Printf("deleting %s\n", f)
		err := os.RemoveAll(f)
		if err != nil {
			fmt.Printf("unable to delete %s: %v\n", f, err)
		}
	}
}
