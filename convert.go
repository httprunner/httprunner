package httpboomer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

func (tc *TestCase) toStruct() *TCase {
	tcStruct := TCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		tcStruct.TestSteps = append(tcStruct.TestSteps, step.ToStruct())
	}
	return &tcStruct
}

func (tc *TestCase) dump2JSON(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Printf("convert absolute path error: %v, path: %v", err, path)
		return err
	}
	log.Printf("dump testcase to json path: %s", path)
	tcStruct := tc.toStruct()
	file, _ := json.MarshalIndent(tcStruct, "", "    ")
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		log.Printf("dump json path error: %v", err)
		return err
	}
	return nil
}
