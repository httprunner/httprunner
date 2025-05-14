package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// IntOrString supports int or string
type IntOrString struct {
	IntValue    *int    // e.g 513
	StringValue *string // e.g "513", "$var"
}

// Value returns the int value, converting from string if necessary
func (ios *IntOrString) Value() (int, error) {
	if ios == nil {
		return 0, nil
	}

	if ios.IntValue != nil {
		return *ios.IntValue, nil
	}
	if ios.StringValue != nil {
		if *ios.StringValue == "" {
			return 0, nil
		}
		n, err := strconv.Atoi(*ios.StringValue)
		if err != nil {
			// variable expression, e.g. "$var"
			return 0, err
		}
		return n, nil
	}

	// IntValue and StringValue are both nil
	return 0, nil
}

// UnmarshalJSON implements custom JSON unmarshalling for IntOrString
func (ios *IntOrString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		ios.IntValue = &i
		ios.StringValue = nil
		return nil
	}
	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		ios.StringValue = &s
		ios.IntValue = nil
		return nil
	}
	return fmt.Errorf("invalid IntOrString data: %s", string(data))
}

// FloatOrString supports float64 or string
type FloatOrString struct {
	FloatValue  *float64 // e.g 5.13
	StringValue *string  // e.g "5.13", "$var"
}

// Value returns the float value, converting from string if necessary
func (ios *FloatOrString) Value() (float64, error) {
	if ios == nil {
		return 0, nil
	}

	if ios.FloatValue != nil {
		return *ios.FloatValue, nil
	}
	if ios.StringValue != nil {
		if *ios.StringValue == "" {
			return 0, nil
		}
		n, err := strconv.ParseFloat(*ios.StringValue, 64)
		if err != nil {
			// variable expression, e.g. "$var"
			return 0, err
		}
		return n, nil
	}

	// IntValue and StringValue are both nil
	return 0, nil
}

// UnmarshalJSON implements custom JSON unmarshalling for IntOrString
func (ios *FloatOrString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as float
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		ios.FloatValue = &f
		ios.StringValue = nil
		return nil
	}
	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		ios.StringValue = &s
		ios.FloatValue = nil
		return nil
	}
	return fmt.Errorf("invalid FloatOrString data: %s", string(data))
}
