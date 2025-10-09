package config 

import (
	"os"
)

const (
	configFileName = ".gatorconfig.json"
)

func getConfigFilePath() (string, error) {
	home_path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	full_path := home_path + "/" + configFileName
	return full_path, nil
}