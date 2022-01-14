package shared

import (
	"fmt"
	"reflect"
)

// CallFunc calls function with arguments
// it is used when calling go plugin or builtin functions
func CallFunc(fn reflect.Value, args ...interface{}) (interface{}, error) {
	fnArgsNum := fn.Type().NumIn()
	if fnArgsNum > 0 && fn.Type().In(fnArgsNum-1).Kind() == reflect.Slice {
		// last argument is slice, do not check arguments number
		// e.g. ...interface{}
		// e.g. a, b string, c ...interface{}
	} else if fnArgsNum != len(args) {
		// function arguments not match
		return nil, fmt.Errorf("function arguments number not match, expect %d, got %d", fnArgsNum, len(args))
	}
	// arguments do not have slice, and arguments number matched

	argumentsValue := make([]reflect.Value, len(args))
	for index, argument := range args {
		if argument == nil {
			argumentsValue[index] = reflect.Zero(fn.Type().In(index))
			continue
		}

		argumentValue := reflect.ValueOf(argument)
		expectArgumentType := fn.Type().In(index)
		actualArgumentType := reflect.TypeOf(argument)

		// type match
		if expectArgumentType == actualArgumentType {
			argumentsValue[index] = argumentValue
			continue
		}

		// type not match, check if convertible
		if !actualArgumentType.ConvertibleTo(expectArgumentType) {
			// function argument type not match and not convertible
			err := fmt.Errorf("function argument %d's type is neither match nor convertible, expect %v, actual %v",
				index, expectArgumentType, actualArgumentType)
			return nil, err
		}
		// convert argument to expect type
		argumentsValue[index] = argumentValue.Convert(expectArgumentType)
	}

	resultValues := fn.Call(argumentsValue)
	if resultValues == nil {
		// no returns
		return nil, nil
	} else if len(resultValues) == 2 {
		// return two arguments: interface{}, error
		if resultValues[1].Interface() != nil {
			return resultValues[0].Interface(), resultValues[1].Interface().(error)
		} else {
			return resultValues[0].Interface(), nil
		}
	} else if len(resultValues) == 1 {
		// return one argument
		if err, ok := resultValues[0].Interface().(error); ok {
			// return error
			return nil, err
		} else {
			// return interface{}
			return resultValues[0].Interface(), nil
		}
	} else {
		// return more than 2 arguments, unexpected
		err := fmt.Errorf("function should return at most 2 arguments")
		return nil, err
	}
}
