package json

import (
	jsoniter "github.com/json-iterator/go"
)

// replace with third-party json library to improve performance
var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	Marshal       = json.Marshal
	MarshalIndent = json.MarshalIndent
	Unmarshal     = json.Unmarshal
	NewDecoder    = json.NewDecoder
	Get           = json.Get
)
