package config

import (
	"os"
	"encoding/json"
)

func Read() (Config, error) {
	full_path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(full_path)
	if err != nil {
		return Config{}, err
	}
	
	new_config := Config{}
	err = json.Unmarshal(data, &new_config)
	if err != nil {
		return Config{}, err
	}

	return new_config, nil
}