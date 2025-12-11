package githubbackup

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func loadEnvFile(log logrus.FieldLogger) error {
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		log.Info("no .env file found")
		return nil
	}

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	log.Info("loaded .env file")
	return nil
}
