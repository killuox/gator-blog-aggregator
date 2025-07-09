package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const fileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var configData Config
	err = json.Unmarshal(content, &configData)
	if err != nil {
		return Config{}, err
	}

	return configData, nil
}

func (c *Config) SetUser(username string) error {
	configData, err := Read()
	if err != nil {
		return err
	}

	configData.CurrentUserName = username

	err = write(configData)
	if err != nil {
		return err
	}

	return nil
}

// Private functions
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("%s/%s", homeDir, fileName)

	return filePath, nil
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
