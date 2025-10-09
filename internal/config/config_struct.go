package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUserName = user
	full_path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(full_path, data, 0777)
	if err != nil {
		return err
	}

	return nil
}