package pluginUtils

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type data struct {
	f      interface{}
	args   []interface{}
	expVal interface{}
	expErr error
}

func TestCallFuncBasic(t *testing.T) {
	params := []data{
		// zero argument, zero return
		{f: func() {}, args: []interface{}{}, expVal: nil, expErr: nil},
		// zero argument, return one value
		{f: func() int { return 1 }, args: []interface{}{}, expVal: 1, expErr: nil},
		{f: func() string { return "a" }, args: []interface{}{}, expVal: "a", expErr: nil},
		{f: func() interface{} { return 1.23 }, args: []interface{}{}, expVal: 1.23, expErr: nil},
		// zero argument, return error
		{f: func() error { return errors.New("xxx") }, args: []interface{}{}, expVal: nil, expErr: errors.New("xxx")},
		// zero argument, return one value and error
		{f: func() (int, error) { return 1, errors.New("xxx") }, args: []interface{}{}, expVal: 1, expErr: errors.New("xxx")},
		{f: func() (interface{}, error) { return 1.23, errors.New("xxx") }, args: []interface{}{}, expVal: 1.23, expErr: errors.New("xxx")},

		// one argument, zero return
		{f: func(n int) {}, args: []interface{}{1}, expVal: nil, expErr: nil},
		// one argument, return one value
		{f: func(n int) int { return n * n }, args: []interface{}{2}, expVal: 4},
		{f: func(c string) string { return c + c }, args: []interface{}{"p"}, expVal: "pp"},
		{f: func(arg interface{}) interface{} { return fmt.Sprintf("%v", arg) }, args: []interface{}{1.23}, expVal: "1.23"},
		// one argument, return one value and error
		{f: func(arg interface{}) (interface{}, error) { return 1.23, errors.New("xxx") }, args: []interface{}{"a"}, expVal: 1.23, expErr: errors.New("xxx")},

		// two arguments in same type
		{f: func(a, b int) int { return a * b }, args: []interface{}{2, 3}, expVal: 6},
		// two arguments in different type
		{
			f: func(n int, c string) string {
				var s string
				for i := 0; i < n; i++ {
					s += c
				}
				return s
			},
			args:   []interface{}{3, "p"},
			expVal: "ppp",
		},

		// variable arguments list: ...int, ...interface{}
		{
			f: func(n ...int) int {
				var sum int
				for _, arg := range n {
					sum += arg
				}
				return sum
			},
			args:   []interface{}{1, 2, 3},
			expVal: 6,
		},
		{
			f: func(args ...interface{}) (interface{}, error) {
				var result string
				for _, arg := range args {
					result += fmt.Sprintf("%v", arg)
				}
				return result, nil
			},
			args:   []interface{}{1, 2.3, "4.5", "p"},
			expVal: "12.34.5p",
		},
		{
			f: func(a, b int8, n ...int) int {
				var sum int
				for _, arg := range n {
					sum += arg
				}
				sum += int(a) + int(b)
				return sum
			},
			args:   []interface{}{1, 2, 3, 4.5},
			expVal: 10,
		},
		{
			f: func(a, b int8, n ...int) int {
				sum := int(a) + int(b)
				for _, arg := range n {
					sum += arg
				}
				return sum
			},
			args:   []interface{}{1, 2},
			expVal: 3,
		},

		{
			f: func(a []int, n ...int) int {
				var sum int
				for _, arg := range a {
					sum += arg
				}
				for _, arg := range n {
					sum += arg
				}
				return sum
			},
			args:   []interface{}{[]int{1, 2}, 3, 4},
			expVal: 10,
		},
	}

	for _, p := range params {
		fn := reflect.ValueOf(p.f)
		val, err := CallFunc(fn, p.args...)
		if !assert.Equal(t, p.expErr, err) {
			t.Fatal(err)
		}
		if !assert.Equal(t, p.expVal, val) {
			t.Fatal()
		}
	}

}

func TestCallFuncComplex(t *testing.T) {
	params := []data{
		// arguments include slice
		{
			f: func(a int, n []int, b int) int {
				sum := a
				for _, arg := range n {
					sum += arg
				}
				sum += b
				return sum
			},
			args:   []interface{}{1, []int{2, 3}, 4},
			expVal: 10,
		},
		// last argument is slice
		{
			f: func(n []int) int {
				var sum int
				for _, arg := range n {
					sum += arg
				}
				return sum
			},
			args:   []interface{}{[]int{1, 2, 3}},
			expVal: 6,
		},
		{
			f: func(a, b int, n []int) int {
				sum := a + b
				for _, arg := range n {
					sum += arg
				}
				return sum
			},
			args:   []interface{}{1, 2, []int{3, 4}},
			expVal: 10,
		},
	}

	for _, p := range params {
		fn := reflect.ValueOf(p.f)
		val, err := CallFunc(fn, p.args...)
		if !assert.Equal(t, p.expErr, err) {
			t.Fatal(err)
		}
		if !assert.Equal(t, p.expVal, val) {
			t.Fatal()
		}
	}

}

func TestCallFuncAbnormal(t *testing.T) {
	params := []data{
		// return more than 2 values
		{
			f:      func() (int, int, error) { return 1, 2, nil },
			args:   []interface{}{},
			expVal: nil,
			expErr: fmt.Errorf("function should return at most 2 values"),
		},
	}

	for _, p := range params {
		fn := reflect.ValueOf(p.f)
		val, err := CallFunc(fn, p.args...)
		if !assert.Equal(t, p.expErr, err) {
			t.Fatal(err)
		}
		if !assert.Equal(t, p.expVal, val) {
			t.Fatal()
		}
	}

}
