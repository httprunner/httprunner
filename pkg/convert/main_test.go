package convert

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/stretchr/testify/assert"
)

const (
	profilePath         = "../../../examples/data/profile.yml"
	profileOverridePath = "../../../examples/data/profile_override.yml"
)

var converter *TCaseConverter

func init() {
	converter = NewConverter("", "")
}

func TestLoadTCase(t *testing.T) {
	err := converter.loadCase(harPath, FromTypeHAR)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, converter.tCase) {
		t.Fatal()
	}
}

func TestLoadHARWithProfileOverride(t *testing.T) {
	err := converter.loadCase(harPath, FromTypeHAR)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, converter.tCase) {
		t.Fatal()
	}

	// override TCase with profile
	err = converter.overrideWithProfile(profileOverridePath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	for i := 0; i < 3; i++ {
		assert.Equal(t,
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			converter.tCase.Steps[i].Request.Headers)
		assert.Equal(t,
			map[string]string{"UserName": "debugtalk"},
			converter.tCase.Steps[i].Request.Cookies)
	}
}

func TestMakeRequestWithProfile(t *testing.T) {
	caseConverter := &TCaseConverter{
		tCase: &hrp.TestCaseDef{
			Steps: []*hrp.TStep{
				{
					Request: &hrp.Request{
						Method: hrp.HTTPMethod("POST"),
						Headers: map[string]string{
							"Content-Type": "application/json; charset=utf-8",
							"User-Agent":   "hrp",
						},
						Cookies: map[string]string{
							"abc":      "123",
							"UserName": "leolee",
						},
					},
				},
			},
		},
	}

	err := caseConverter.overrideWithProfile(profilePath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded", "User-Agent": "hrp",
	}, caseConverter.tCase.Steps[0].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{
		"UserName": "debugtalk", "abc": "123",
	}, caseConverter.tCase.Steps[0].Request.Cookies) {
		t.Fatal()
	}
}

func TestMakeRequestWithProfileOverride(t *testing.T) {
	caseConverter := &TCaseConverter{
		tCase: &hrp.TestCaseDef{
			Steps: []*hrp.TStep{
				{
					Request: &hrp.Request{
						Method: hrp.HTTPMethod("POST"),
						Headers: map[string]string{
							"Content-Type": "application/json; charset=utf-8",
							"User-Agent":   "hrp",
						},
						Cookies: map[string]string{
							"abc":      "123",
							"UserName": "leolee",
						},
					},
				},
			},
		},
	}

	// override TCase with profile
	err := caseConverter.overrideWithProfile(profileOverridePath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, caseConverter.tCase.Steps[0].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{
		"UserName": "debugtalk",
	}, caseConverter.tCase.Steps[0].Request.Cookies) {
		t.Fatal()
	}
}
