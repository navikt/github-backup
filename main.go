package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/navikt/github-backup/pkg/git"
	"github.com/navikt/github-backup/pkg/objstorage"
	"github.com/navikt/github-backup/pkg/zippings"
)

var (
	basedir = filepath.Join(os.TempDir(), "ghbackup")
	orgs    = []string{"navikt", "nais"}
)

const MaxConcurrent = 3

func main() {
	bucketname := envOrDie("BUCKET_NAME")
	githubToken := envOrDie("GITHUB_TOKEN")

	var repos []git.Repo
	for _, org := range orgs {
		repos = append(repos, reposOrDie(org, githubToken)...)
	}
	fmt.Printf("found %d repos\n", len(repos))

	goog, err := storage.NewClient(context.Background())
	if err != nil {
		fmt.Printf("unable to create gcs client: %v\n", err)
		os.Exit(1)
	}
	defer goog.Close()

	workQueue := make(chan int, MaxConcurrent)
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for i, repo := range repos {
		fmt.Printf("processing repo %d of %d\n", i+1, len(repos))
		r := repo
		workQueue <- 1
		go func() {
			err := cloneZipAndStoreInBucket(r.FullName, bucketname, githubToken, goog)
			if err != nil {
				fmt.Printf("failed to backup repo '%s': %v\n", r.FullName, err)
			}
			<-workQueue
			wg.Done()
		}()
	}
	wg.Wait()
}

func cloneZipAndStoreInBucket(repo string, bucketname string, githubToken string, gcsClient *storage.Client) error {
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
	if err != nil {
		rm([]string{repodir, compressedFilePath})
		return err
	}
	defer file.Close()

	objBasePath := time.Now().Format("2006/01/02")
	err = objstorage.CopyToBucket(gcsClient, file, bucketname, objBasePath)
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
