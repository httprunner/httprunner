package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp"
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
		if !assert.Equal(t,
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			converter.tCase.TestSteps[i].Request.Headers) {
			t.FailNow()
		}
		if !assert.Equal(t,
			map[string]string{"UserName": "debugtalk"},
			converter.tCase.TestSteps[i].Request.Cookies) {
			t.FailNow()
		}
	}
}

func TestMakeRequestWithProfile(t *testing.T) {
	caseConverter := &TCaseConverter{
		tCase: &hrp.TCase{
			TestSteps: []*hrp.TStep{
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
	}, caseConverter.tCase.TestSteps[0].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{
		"UserName": "debugtalk", "abc": "123",
	}, caseConverter.tCase.TestSteps[0].Request.Cookies) {
		t.Fatal()
	}
}

func TestMakeRequestWithProfileOverride(t *testing.T) {
	caseConverter := &TCaseConverter{
		tCase: &hrp.TCase{
			TestSteps: []*hrp.TStep{
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
	}, caseConverter.tCase.TestSteps[0].Request.Headers) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{
		"UserName": "debugtalk",
	}, caseConverter.tCase.TestSteps[0].Request.Cookies) {
		t.Fatal()
	}
}
