//go:build localtest

package ai

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetImageFromBuffer(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"
	file, err := os.ReadFile(imagePath)
	require.Nil(t, err)
	buf := new(bytes.Buffer)
	buf.Read(file)

	service, err := NewVEDEMImageService()
	require.Nil(t, err)
	cvResult, err := service.ReadFromBuffer(buf)
	assert.Nil(t, err)
	fmt.Printf("cvResult: %v\n", cvResult)
}

func TestGetImageFromPath(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"
	service, err := NewVEDEMImageService()
	require.Nil(t, err)
	cvResult, err := service.ReadFromPath(imagePath)
	assert.Nil(t, err)
	fmt.Printf("cvResult: %v\n", cvResult)
}
