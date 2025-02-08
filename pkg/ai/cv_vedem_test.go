//go:build localtest

package ai

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestGetImageFromBuffer(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"
	file, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	buf.Read(file)

	service := NewAIService(
		WithCVService(CVServiceTypeVEDEM),
	)
	cvResult, err := service.ReadFromBuffer(buf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("cvResult: %v", cvResult))
}

func TestGetImageFromPath(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"
	service := NewAIService(
		WithCVService(CVServiceTypeVEDEM),
	)
	cvResult, err := service.ReadFromPath(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("cvResult: %v", cvResult))
}
