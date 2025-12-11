package githubbackup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/navikt/github-backup/internal/github"
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

	repos, err := github.ReposInOrgs(ctx, config.GitHubOrgs, config.GitHubToken)
	if err != nil {
		return fmt.Errorf("error fetching repos: %w", err)
	}
	fmt.Printf("found %d repos\n", len(repos))

	goog, err := storage.NewClient(ctx)
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
			err := cloneZipAndStoreInBucket(r.GetFullName(), config.BucketName, config.GitHubToken, goog)
			if err != nil {
				fmt.Printf("failed to backup repo %q: %v\n", r.GetFullName(), err)
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

	err := github.CloneRepo(basedir, repo, "NAVGitHubBackup", githubToken)
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

func rm(entries []string) {
	for _, f := range entries {
		fmt.Printf("deleting %s\n", f)
		err := os.RemoveAll(f)
		if err != nil {
			fmt.Printf("unable to delete %s: %v\n", f, err)
		}
	}
}
