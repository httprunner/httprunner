package postman2case

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var collectionPath = "../../../examples/data/postman2case/postman_collection.json"

func TestLoadPostmanCollection(t *testing.T) {
	c, err := NewCollection(collectionPath).load()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	if !assert.Equal(t, "postman collection demo", c.Info.Name) {
		t.Fatal()
	}
}

func TestGenJSON(t *testing.T) {
	jsonPath, err := NewCollection(collectionPath).GenJSON()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, jsonPath) {
		t.Fatal()
	}
}

func TestGenYAML(t *testing.T) {
	yamlPath, err := NewCollection(collectionPath).GenYAML()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, yamlPath) {
		t.Fatal()
	}
}
