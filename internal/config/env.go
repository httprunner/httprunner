package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var loadEnvOnce sync.Once

// LoadEnv loads environment variables from .env file
// it will search for .env file from current working directory upward recursively
// if not found, it will try to load from ~/.hrp/.env as fallback
// Priority: current working directory > ~/.hrp/.env > system environment variables
func LoadEnv() (err error) {
	loadEnvOnce.Do(func() {
		// first try to load from ~/.hrp/.env, override system env variables (medium priority)
		var homeDir string
		homeDir, err = os.UserHomeDir()
		if err != nil {
			log.Warn().Err(err).Msg("get user home directory failed")
		} else {
			globalEnvFile := filepath.Join(homeDir, ".hrp", ".env")
			if _, e := os.Stat(globalEnvFile); e == nil {
				// load global .env file and override system environment variables
				err = godotenv.Overload(globalEnvFile)
				if err != nil {
					log.Error().Err(err).
						Str("path", globalEnvFile).Msg("load global env file failed")
					return
				}
				log.Info().Str("path", globalEnvFile).Msg("load global env success")
			}
		}

		// get current working directory
		var cwd string
		cwd, err = os.Getwd()
		if err != nil {
			log.Error().Err(err).Msg("get current working directory failed")
			return
		}

		// locate .env file from current working directory upward recursively
		envPath := cwd
		for {
			envFile := filepath.Join(envPath, ".env")
			if _, e := os.Stat(envFile); e == nil {
				// found .env file
				// override existing env variables (highest priority)
				err = godotenv.Overload(envFile)
				if err != nil {
					log.Error().Err(err).
						Str("path", envFile).Msg("overload env file failed")
					return
				}
				log.Info().Str("path", envFile).Msg("overload env success")
				return
			}

			// reached root directory
			parent := filepath.Dir(envPath)
			if parent == envPath {
				log.Info().Msg("no .env file found from current directory to root")
				return
			}
			envPath = parent
		}
	})
	return err
}

func GetEnvConfig(key string) string {
	return os.Getenv(key)
}

func GetEnvConfigInJSON(key string) (map[string]interface{}, error) {
	value := GetEnvConfig(key)
	if value == "" {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetEnvConfigInBool(key string) bool {
	value := GetEnvConfig(key)
	if value == "" {
		return false
	}

	boolValue, _ := strconv.ParseBool(value)
	return boolValue
}

// GetEnvConfigOrDefault get env config or default value
func GetEnvConfigOrDefault(key, defaultValue string) string {
	value := GetEnvConfig(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvConfigInInt(key string, defaultValue int) int {
	value := GetEnvConfig(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
