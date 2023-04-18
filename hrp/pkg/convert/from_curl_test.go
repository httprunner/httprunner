package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var curlPath = "../../../examples/data/curl/curl_examples.txt"

func TestLoadCurlCase(t *testing.T) {
	tCase, err := LoadCurlCase(curlPath)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	if !assert.Equal(t, 6, len(tCase.TestSteps)) {
		t.Fatal()
	}

	// curl httpbin.org
	if !assert.Equal(t, "curl httpbin.org", tCase.TestSteps[0].Name) {
		t.Fatal()
	}
	if !assert.EqualValues(t, "GET", tCase.TestSteps[0].Request.Method) {
		t.Fatal()
	}
	if !assert.Equal(t, "http://httpbin.org", tCase.TestSteps[0].Request.URL) {
		t.Fatal()
	}

	// curl https://httpbin.org/get?key1=value1&key2=value2
	if !assert.Equal(t, "https://httpbin.org/get", tCase.TestSteps[1].Request.URL) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}, tCase.TestSteps[1].Request.Params) {
		t.Fatal()
	}

	// curl -H "Content-Type: application/json" \
	//    -H "Authorization: Bearer b7d03a6947b217efb6f3ec3bd3504582" \
	//    -d '{"type":"A","name":"www","data":"162.10.66.0","priority":null,"port":null,"weight":null}' \
	//    "https://httpbin.org/post"
	if !assert.EqualValues(t, "POST", tCase.TestSteps[2].Request.Method) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{
		"Authorization": "Bearer b7d03a6947b217efb6f3ec3bd3504582",
		"Content-Type":  "application/json",
	}, tCase.TestSteps[2].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]interface{}{
		"data":     "162.10.66.0",
		"name":     "www",
		"port":     nil,
		"priority": nil,
		"type":     "A",
		"weight":   nil,
	}, tCase.TestSteps[2].Request.Body) {
		t.Fatal()
	}

	// curl -F "dummyName=dummyFile" -F file1=@file1.txt -F file2=@file2.txt https://httpbin.org/post
	if !assert.Equal(t, map[string]interface{}{
		"dummyName": "dummyFile",
		"file1":     "@file1.txt",
		"file2":     "@file2.txt",
	}, tCase.TestSteps[3].Request.Upload) {
		t.Fatal()
	}

	// curl https://httpbin.org/post \
	//     -d 'shipment[to_address][id]=adr_HrBKVA85' \
	//     -d 'shipment[from_address][id]=adr_VtuTOj7o' \
	//     -d 'shipment[parcel][id]=prcl_WDv2VzHp' \
	//     -d 'shipment[is_return]=true' \
	//     -d 'shipment[customs_info][id]=cstinfo_bl5sE20Y'
	if !assert.Equal(t, map[string]interface{}{
		"shipment[customs_info][id]": "cstinfo_bl5sE20Y",
		"shipment[from_address][id]": "adr_VtuTOj7o",
		"shipment[is_return]":        "true",
		"shipment[parcel][id]":       "prcl_WDv2VzHp",
		"shipment[to_address][id]":   "adr_HrBKVA85",
	}, tCase.TestSteps[4].Request.Body) {
		t.Fatal()
	}

	// curl https://httpbing.org/post -H "Content-Type: application/x-www-form-urlencoded" \
	//     --data "key1=value+1&key2=value%3A2"
	if !assert.Equal(t, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, tCase.TestSteps[5].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]interface{}{
		"key1": "value 1",
		"key2": "value:2",
	}, tCase.TestSteps[5].Request.Body) {
		t.Fatal()
	}
}
