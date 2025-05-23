package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestLoadParameters(t *testing.T) {
	testData := []struct {
		configParameters map[string]interface{}
		loadedParameters map[string]hrp.Parameters
	}{
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${parameterize(%s/$file)}", hrpExamplesDir),
			},
			map[string]hrp.Parameters{
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
				"user_agent":  []interface{}{"iOS/10.1", "iOS/10.2"},
				"app_version": []interface{}{4.0},
			},
			map[string]hrp.Parameters{
				"username-password": {
					{"username": "test1", "password": "111111"},
					{"username": "test2", "password": "222222"},
				},
				"user_agent": {
					{"user_agent": "iOS/10.1"},
					{"user_agent": "iOS/10.2"},
				},
				"app_version": {
					{"app_version": 4.0},
				},
			},
		},
		{
			map[string]interface{}{
				"username-password": []interface{}{
					[]interface{}{"test1", "111111"},
					[]interface{}{"test2", "222222"},
				},
			},
			map[string]hrp.Parameters{
				"username-password": {
					{"username": "test1", "password": "111111"},
					{"username": "test2", "password": "222222"},
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
	parser := hrp.NewParser()
	for _, data := range testData {
		value, err := parser.LoadParameters(data.configParameters, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.loadedParameters, value)
	}
}

func TestLoadParametersError(t *testing.T) {
	testData := []struct {
		configParameters map[string]interface{}
	}{
		{
			map[string]interface{}{
				"username_password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir),
				"user_agent":        []interface{}{"iOS/10.1", "iOS/10.2"},
			},
		},
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir),
				"user-agent":        []interface{}{"iOS/10.1", "iOS/10.2"},
			},
		},
		{
			map[string]interface{}{
				"username-password": fmt.Sprintf("${param(%s/account.csv)}", hrpExamplesDir),
				"user_agent":        []interface{}{"iOS/10.1", "iOS/10.2"},
			},
		},
	}
	parser := hrp.NewParser()
	for _, data := range testData {
		_, err := parser.LoadParameters(data.configParameters, map[string]interface{}{})
		assert.Error(t, err)
	}
}

func TestInitParametersIteratorCount(t *testing.T) {
	configParameters := map[string]interface{}{
		"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir), // 3
		"user_agent":        []interface{}{"iOS/10.1", "iOS/10.2"},                          // 2
		"app_version":       []interface{}{4.0},                                             // 1
	}
	testData := []struct {
		cfg         *hrp.TConfig
		expectLimit int
	}{
		// default, no parameters setting
		{
			&hrp.TConfig{
				Parameters:        configParameters,
				ParametersSetting: &hrp.TParamsConfig{},
			},
			6, // 3 * 2 * 1
		},
		{
			&hrp.TConfig{
				Parameters: configParameters,
			},
			6, // 3 * 2 * 1
		},
		// default equals to set overall parameters pick-order to "sequential"
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					PickOrder: "sequential",
				},
			},
			6, // 3 * 2 * 1
		},
		// default equals to set each individual parameters pick-order to "sequential"
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					Strategies: map[string]hrp.IteratorStrategy{
						"username-password": {Name: "user-info", PickOrder: "sequential"},
						"user_agent":        {Name: "user-identity", PickOrder: "sequential"},
						"app_version":       {Name: "app-version", PickOrder: "sequential"},
					},
				},
			},
			6, // 3 * 2 * 1
		},
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					Strategies: map[string]hrp.IteratorStrategy{
						"user_agent":  {Name: "user-identity", PickOrder: "sequential"},
						"app_version": {Name: "app-version", PickOrder: "sequential"},
					},
				},
			},
			6, // 3 * 2 * 1
		},

		// set overall parameters overall pick-order to "random"
		// each random parameters only select one item
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					PickOrder: "random",
				},
			},
			1, // 1 * 1 * 1
		},
		// set some individual parameters pick-order to "random"
		// this will override overall strategy
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					Strategies: map[string]hrp.IteratorStrategy{
						"user_agent": {Name: "user-identity", PickOrder: "random"},
					},
				},
			},
			3, // 3 * 1 * 1
		},
		{
			&hrp.TConfig{
				Parameters: configParameters,
				ParametersSetting: &hrp.TParamsConfig{
					Strategies: map[string]hrp.IteratorStrategy{
						"username-password": {Name: "user-info", PickOrder: "random"},
					},
				},
			},
			2, // 1 * 2 * 1
		},

		// set limit for parameters
		{
			&hrp.TConfig{
				Parameters: configParameters, // total: 6 = 3 * 2 * 1
				ParametersSetting: &hrp.TParamsConfig{
					Limit: 4, // limit could be less than total
				},
			},
			4,
		},
		{
			&hrp.TConfig{
				Parameters: configParameters, // total: 6 = 3 * 2 * 1
				ParametersSetting: &hrp.TParamsConfig{
					Limit: 9, // limit could also be greater than total
				},
			},
			9,
		},

		// no parameters
		// also will generate one empty item
		{
			&hrp.TConfig{
				Parameters:        nil,
				ParametersSetting: nil,
			},
			1,
		},
	}
	parser := hrp.NewParser()
	for _, data := range testData {
		iterator, err := parser.InitParametersIterator(data.cfg)
		assert.Nil(t, err)
		assert.Equal(t, data.expectLimit, iterator.Limit)

		for i := 0; i < data.expectLimit; i++ {
			assert.True(t, iterator.HasNext())
			iterator.Next() // consume next parameters
		}
		// should not have next
		assert.False(t, iterator.HasNext())
	}
}

func TestInitParametersIteratorUnlimitedCount(t *testing.T) {
	configParameters := map[string]interface{}{
		"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir), // 3
		"user_agent":        []interface{}{"iOS/10.1", "iOS/10.2"},                          // 2
		"app_version":       []interface{}{4.0},                                             // 1
	}
	testData := []struct {
		cfg *hrp.TConfig
	}{
		// default, no parameters setting
		{
			&hrp.TConfig{
				Parameters:        configParameters,
				ParametersSetting: &hrp.TParamsConfig{},
			},
		},

		// no parameters
		// also will generate one empty item
		{
			&hrp.TConfig{
				Parameters:        nil,
				ParametersSetting: nil,
			},
		},
	}
	parser := hrp.NewParser()
	for _, data := range testData {
		iterator, err := parser.InitParametersIterator(data.cfg)
		assert.Nil(t, err)
		// set unlimited mode
		iterator.SetUnlimitedMode()
		assert.Equal(t, -1, iterator.Limit)

		for i := 0; i < 100; i++ {
			assert.True(t, iterator.HasNext())
			iterator.Next() // consume next parameters
		}
		assert.Equal(t, 100, iterator.Index)
		// should also have next
		assert.True(t, iterator.HasNext())
	}
}

func TestInitParametersIteratorContent(t *testing.T) {
	configParameters := map[string]interface{}{
		"username-password": fmt.Sprintf("${parameterize(%s/account.csv)}", hrpExamplesDir), // 3
		"user_agent":        []interface{}{"iOS/10.1", "iOS/10.2"},                          // 2
		"app_version":       []interface{}{4.0},                                             // 1
	}
	testData := []struct {
		cfg              *hrp.TConfig
		checkIndex       int
		expectParameters map[string]interface{}
	}{
		// default, no parameters setting
		{
			&hrp.TConfig{
				Parameters: configParameters,
			},
			0, // check first item
			map[string]interface{}{
				"username": "test1", "password": "111111", "user_agent": "iOS/10.1", "app_version": 4.0,
			},
		},

		// set limit for parameters
		{
			&hrp.TConfig{
				Parameters: map[string]interface{}{
					"username-password": []map[string]interface{}{ // 1
						{"username": "test1", "password": 111111, "other": "111"},
					},
					"user_agent": []string{"iOS/10.1", "iOS/10.2"}, // 2
				},
				ParametersSetting: &hrp.TParamsConfig{
					Limit: 5, // limit could also be greater than total
					Strategies: map[string]hrp.IteratorStrategy{
						"username-password": {Name: "user-info", PickOrder: "random"},
					},
				},
			},
			2, // check 3th item, equals to the first item
			map[string]interface{}{
				"username": "test1", "password": 111111, "user_agent": "iOS/10.1",
			},
		},

		// no parameters
		// also will generate one empty item
		{
			&hrp.TConfig{
				Parameters:        nil,
				ParametersSetting: nil,
			},
			0,
			map[string]interface{}{},
		},
	}
	parser := hrp.NewParser()
	for _, data := range testData {
		iterator, err := parser.InitParametersIterator(data.cfg)
		assert.Nil(t, err)

		// get expected parameters item
		for i := 0; i < data.checkIndex; i++ {
			assert.True(t, iterator.HasNext())
			iterator.Next() // consume next parameters
		}
		parametersItem := iterator.Next()

		assert.Equal(t, data.expectParameters, parametersItem)
	}
}

func TestGenCartesianProduct(t *testing.T) {
	testData := []struct {
		multiParameters []hrp.Parameters
		expect          hrp.Parameters
	}{
		{
			[]hrp.Parameters{
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
			hrp.Parameters{
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
			[]hrp.Parameters{},
			nil,
		},
	}

	for _, data := range testData {
		parameters := hrp.GenCartesianProduct(data.multiParameters)
		assert.Equal(t, data.expect, parameters)
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
		value, err := hrp.ConvertParameters(data.key, data.parametersRawList)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, value)
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
		_, err := hrp.ConvertParameters(data.key, data.parametersRawList)
		assert.Error(t, err)
	}
}
