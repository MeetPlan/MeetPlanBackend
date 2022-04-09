package sql

import (
	"encoding/json"
	"os"
)

type Config struct {
	DatabaseName   string `json:"database_name"`
	DatabaseConfig string `json:"database_config"`
	Debug          bool   `json:"debug"`
	Host           string `json:"host"`
}

func GetConfig() (Config, error) {
	var config Config
	file, err := os.ReadFile("config.json")
	if err != nil {
		marshal, err := json.Marshal(Config{
			DatabaseName:   "sqlite3",
			DatabaseConfig: "MeetPlanDB/meetplan.db",
			Debug:          true,
			Host:           "127.0.0.1:8000",
		})
		if err != nil {
			return config, err
		}
		err = os.WriteFile("config.json", marshal, 0600)
		if err != nil {
			return config, err
		}
		file, err = os.ReadFile("config.json")
		if err != nil {
			return config, err
		}
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	return config, err
}

func SaveConfig(config Config) error {
	marshal, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = os.Remove("config.json")
	if err != nil {
		return err
	}
	err = os.WriteFile("config.json", marshal, 0600)
	if err != nil {
		return err
	}
	return nil
}
