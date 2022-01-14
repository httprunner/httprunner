package shared

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallFunc(t *testing.T) {
	type data struct {
		f      interface{}
		args   []interface{}
		expVal interface{}
		expErr error
	}
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
		// one argument, return one value
		{f: func(n int) int { return n * n }, args: []interface{}{2}, expVal: 4},
		{f: func(c string) string { return c + c }, args: []interface{}{"p"}, expVal: "pp"},
		{f: func(arg interface{}) interface{} { return fmt.Sprintf("%v", arg) }, args: []interface{}{1.23}, expVal: "1.23"},
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
