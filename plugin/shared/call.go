package shared

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
)

// CallFunc calls function with arguments
// it is used when calling go plugin or builtin functions
func CallFunc(fn reflect.Value, args ...interface{}) (interface{}, error) {
	fnArgsNum := fn.Type().NumIn()

	// function expect 0 argument
	if fnArgsNum == 0 {
		if len(args) > 0 {
			return nil, fmt.Errorf("function expect 0 argument, but got %d", len(args))
		}
		return call(fn, nil)
	}

	// fnArgsNum > 0: function expect more than 0 arguments
	// function last argument is not slice
	if fn.Type().In(fnArgsNum-1).Kind() != reflect.Slice {
		// function arguments should match exactly
		if fnArgsNum != len(args) {
			return nil, fmt.Errorf("function expect %d arguments, but got %d", fnArgsNum, len(args))
		}
		// prepare arguments
		argumentsValue, err := convertFirstNArgs(fn, fnArgsNum, args...)
		if err != nil {
			return nil, err
		}
		return call(fn, argumentsValue)
	}

	// function last argument is slice
	// e.g.
	// ...interface{}
	// a, b string, c ...interface{}
	// []int
	// a, b string, c []int

	// arguments number should >= fnArgsNum-1
	if len(args) < fnArgsNum-1 {
		return nil, fmt.Errorf("function expect at least %d arguments, but got %d", fnArgsNum-1, len(args))
	}

	// prepare first n-1 arguments
	argumentsValue, err := convertFirstNArgs(fn, fnArgsNum-1, args...)
	if err != nil {
		return nil, err
	}

	// prepare last slice argument
	fnLastSliceElemType := fn.Type().In(fnArgsNum - 1).Elem() // slice element type
	for i := fnArgsNum - 1; i < len(args); i++ {              // loop last len(args)-fnArgsNum+1 arguments
		argValue := reflect.ValueOf(args[i])
		argValueType := reflect.TypeOf(args[i])

		// check if slice element type match
		if fnLastSliceElemType == argValueType {
			argumentsValue = append(argumentsValue, argValue)
			continue
		}

		// last argument is also slice, e.g. []int
		if argValueType.Kind() == reflect.Slice {
			if argValueType.Elem() != fnLastSliceElemType {
				err := fmt.Errorf("function argument %d's slice element type is not match, expect %v, actual %v",
					i, fnLastSliceElemType, argValueType)
				return nil, err
			}
			argumentsValue = append(argumentsValue, argValue)
			continue
		}

		// type not match, check if convertible
		if !argValueType.ConvertibleTo(fnLastSliceElemType) {
			// function argument type not match and not convertible
			err := fmt.Errorf("function argument %d's type is neither match nor convertible, expect %v, actual %v",
				i, fnLastSliceElemType, argValueType)
			return nil, err
		}
		// convert argument to expect type
		argumentsValue = append(argumentsValue, argValue.Convert(fnLastSliceElemType))
	}

	return call(fn, argumentsValue)
}

func convertFirstNArgs(fn reflect.Value, n int, args ...interface{}) ([]reflect.Value, error) {
	argumentsValue := make([]reflect.Value, n)
	for index := 0; index < n; index++ {
		argument := args[index]
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
			log.Error().Err(err).Msg("convert arguments failed")
			return nil, err
		}
		// convert argument to expect type
		argumentsValue[index] = argumentValue.Convert(expectArgumentType)
	}
	return argumentsValue, nil
}

func call(fn reflect.Value, args []reflect.Value) (interface{}, error) {
	resultValues := fn.Call(args)
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
		err := fmt.Errorf("function should return at most 2 values")
		return nil, err
	}
}
