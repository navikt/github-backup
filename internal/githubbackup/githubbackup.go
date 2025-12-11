package githubbackup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/navikt/github-backup/internal/git"
	"github.com/navikt/github-backup/internal/objstorage"
	"github.com/navikt/github-backup/internal/zippings"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

var basedir = filepath.Join(os.TempDir(), "ghbackup")

const MaxConcurrent = 3

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func Run(ctx context.Context) error {
	log := logrus.StandardLogger()

	if err := loadEnvFile(log); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	config, err := newConfig(ctx, envconfig.OsLookuper())
	if err != nil {
		return fmt.Errorf("error processing configuration: %w", err)
	}

	var repos []git.Repo
	for _, org := range config.GitHubOrgs {
		repos = append(repos, reposOrDie(org, config.GitHubToken)...)
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
			err := cloneZipAndStoreInBucket(r.FullName, config.BucketName, config.GitHubToken, goog)
			if err != nil {
				fmt.Printf("failed to backup repo '%s': %v\n", r.FullName, err)
			}
			<-workQueue
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
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
