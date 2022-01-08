package scaffold

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/httprunner/hrp/internal/ga"
	"github.com/rs/zerolog/log"
)

func CreateScaffold(projectName string) error {
	// report event
	ga.SendEvent(ga.EventTracking{
		Category: "Scaffold",
		Action:   "hrp startproject",
	})

	// check if projectName exists
	if _, err := os.Stat(projectName); err == nil {
		log.Warn().Str("projectName", projectName).
			Msg("project name already exists, please specify a new one.")
		return fmt.Errorf("project name already exists")
	}

	log.Info().Str("projectName", projectName).Msg("create new scaffold project")

	// create project folders
	if err := createFolder(projectName); err != nil {
		return err
	}
	if err := createFolder(path.Join(projectName, "har")); err != nil {
		return err
	}
	if err := createFolder(path.Join(projectName, "testcases")); err != nil {
		return err
	}
	if err := createFolder(path.Join(projectName, "reports")); err != nil {
		return err
	}

	// create demo testcases
	tCase, _ := demoTestCase.ToTCase()
	err := tCase.Dump2JSON(path.Join(projectName, "testcases", "demo.json"))
	if err != nil {
		log.Error().Err(err).Msg("create demo.json testcase failed")
		return err
	}
	err = tCase.Dump2YAML(path.Join(projectName, "testcases", "demo.yaml"))
	if err != nil {
		log.Error().Err(err).Msg("create demo.yml testcase failed")
		return err
	}

	// create .gitignore
	if err := createFile(path.Join(projectName, ".gitignore"), demoIgnoreContent); err != nil {
		return err
	}
	// create .env
	if err := createFile(path.Join(projectName, ".env"), demoEnvContent); err != nil {
		return err
	}

	return nil
}

func createFolder(folderPath string) error {
	log.Info().Str("path", folderPath).Msg("create folder")
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("create folder failed")
		return err
	}
	return nil
}

func createFile(filePath string, data string) error {
	log.Info().Str("path", filePath).Msg("create file")
	err := ioutil.WriteFile(filePath, []byte(data), 0o644)
	if err != nil {
		log.Error().Err(err).Msg("create file failed")
		return err
	}
	return nil
}
