package hrp

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLoadParameters(t *testing.T) {
	testData := []struct {
		configParameters map[string]interface{}
		loadedParameters map[string]Parameters
	}{
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${parameterize(%s/$file)}", hrpExamplesDir),
			},
			map[string]Parameters{
				"username-password": {
					{"username": "test1", "password": "111111"},
					{"username": "test2", "password": "222222"},
					{"username": "test3", "password": "333333"},
				},
			},
		},
		{
			map[string]interface{}{
				"username-password": [][]interface{}{
					{"test1", "111111"},
					{"test2", "222222"},
				},
				"user_agent":  []interface{}{"IOS/10.1", "IOS/10.2"},
				"app_version": []interface{}{4.0},
			},
			map[string]Parameters{
				"username-password": {
					{"username": "test1", "password": "111111"},
					{"username": "test2", "password": "222222"},
				},
				"user_agent": {
					{"user_agent": "IOS/10.1"},
					{"user_agent": "IOS/10.2"},
				},
				"app_version": {
					{"app_version": 4.0},
				},
			},
		},
		{
			map[string]interface{}{},
			nil,
		},
		{
			nil,
			nil,
		},
	}

	variablesMapping := map[string]interface{}{
		"file": "account.csv",
	}
	for _, data := range testData {
		value, err := loadParameters(data.configParameters, variablesMapping)
		if !assert.Nil(t, err) {
			t.Fatal()
		}
		if !assert.Equal(t, data.loadedParameters, value) {
			t.Fatal()
		}
	}
}

func TestLoadParametersError(t *testing.T) {
	testData := []struct {
		configParameters map[string]interface{}
	}{
		{
			map[string]interface{}{
				"username_password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir),
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir),
				"user-agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${param(%s/account.csv)}", hrpExamplesDir),
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
	}
	for _, data := range testData {
		_, err := loadParameters(data.configParameters, map[string]interface{}{})
		if !assert.Error(t, err) {
			t.Fatal()
		}
	}
}

func TestInitParametersIterator(t *testing.T) {
	configParameters := map[string]interface{}{
		"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir), // 3
		"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"},
		"app_version":       []interface{}{4.0},
	}
	testData := []struct {
		cfg         *TConfig
		expectLimit int
	}{
		{
			&TConfig{
				Parameters:        configParameters,
				ParametersSetting: &TParamsConfig{},
			},
			6,
		},
		{
			&TConfig{
				Parameters: configParameters,
				ParametersSetting: &TParamsConfig{
					Strategy: "random",
				},
			},
			1,
		},
		{
			&TConfig{
				Parameters: configParameters,
				ParametersSetting: &TParamsConfig{
					Strategies: map[string]iteratorStrategy{
						"username-password": "random",
					},
				},
			},
			2,
		},
	}
	for _, data := range testData {
		iterator, err := initParametersIterator(data.cfg)
		if !assert.Nil(t, err) {
			t.Fatal()
		}
		if !assert.Equal(t, data.expectLimit, iterator.limit) {
			t.Fatal()
		}

		for i := 0; i < data.expectLimit; i++ {
			if !assert.True(t, iterator.HasNext()) {
				t.Fatal()
			}
			log.Info().Interface("next", iterator.Next()).Msg("get next parameters")
		}
		// should not have next
		if !assert.False(t, iterator.HasNext()) {
			t.Fatal()
		}
	}
}

func TestGenCartesianProduct(t *testing.T) {
	testData := []struct {
		multiParameters []Parameters
		expect          Parameters
	}{
		{
			[]Parameters{
				{
					{"app_version": 4.0},
				},
				{
					{"username": "test1", "password": "111111"},
					{"username": "test2", "password": "222222"},
				},
				{
					{"user_agent": "iOS/10.1"},
					{"user_agent": "iOS/10.2"},
				},
			},
			Parameters{
				{"app_version": 4.0, "password": "111111", "user_agent": "iOS/10.1", "username": "test1"},
				{"app_version": 4.0, "password": "111111", "user_agent": "iOS/10.2", "username": "test1"},
				{"app_version": 4.0, "password": "222222", "user_agent": "iOS/10.1", "username": "test2"},
				{"app_version": 4.0, "password": "222222", "user_agent": "iOS/10.2", "username": "test2"},
			},
		},
		{
			nil,
			nil,
		},
		{
			[]Parameters{},
			nil,
		},
	}

	for _, data := range testData {
		parameters := genCartesianProduct(data.multiParameters)
		if !assert.Equal(t, data.expect, parameters) {
			t.Fatal()
		}
	}
}

func TestConvertParameters(t *testing.T) {
	testData := []struct {
		key               string
		parametersRawList interface{}
		expect            []map[string]interface{}
	}{
		{
			"username-password",
			[]map[string]interface{}{
				{"username": "test1", "password": 111111, "other": "111"},
				{"username": "test2", "password": 222222, "other": "222"},
			},
			[]map[string]interface{}{
				{"username": "test1", "password": 111111},
				{"username": "test2", "password": 222222},
			},
		},
		{
			"username-password",
			[][]string{
				{"test1", "111111"},
				{"test2", "222222"},
			},
			[]map[string]interface{}{
				{"username": "test1", "password": "111111"},
				{"username": "test2", "password": "222222"},
			},
		},
		{
			"app_version",
			[]float64{3.1, 3.0},
			[]map[string]interface{}{
				{"app_version": 3.1},
				{"app_version": 3.0},
			},
		},
		{
			"user_agent",
			[]string{"iOS/10.1", "iOS/10.2"},
			[]map[string]interface{}{
				{"user_agent": "iOS/10.1"},
				{"user_agent": "iOS/10.2"},
			},
		},
	}

	for _, data := range testData {
		value, err := convertParameters(data.key, data.parametersRawList)
		if !assert.Nil(t, err) {
			t.Fatal()
		}
		if !assert.Equal(t, data.expect, value) {
			t.Fatal()
		}
	}
}

func TestConvertParametersError(t *testing.T) {
	testData := []struct {
		key               string
		parametersRawList interface{}
	}{
		{
			"app_version",
			123, // not slice
		},
		{
			"app_version",
			"123", // not slice
		},
		{
			"username-password",
			[]map[string]interface{}{ // parameter names not match
				{"username": "test1", "other": "111"},
				{"username": "test2", "other": "222"},
			},
		},
		{
			"username-password",
			[][]string{ // parameter names length not match
				{"test1"},
				{"test2"},
			},
		},
	}

	for _, data := range testData {
		_, err := convertParameters(data.key, data.parametersRawList)
		if !assert.Error(t, err) {
			t.Fatal()
		}
	}
}
