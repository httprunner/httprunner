package hrp

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func TestRun(t *testing.T) {
	err := BuildPlugin(tmpl("plugin/debugtalk.go"), "./debugtalk.bin")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	genDebugTalkPyPath := filepath.Join(tmpl("plugin/"), PluginPySourceGenFile)
	err = BuildPlugin(tmpl("plugin/debugtalk.py"), genDebugTalkPyPath)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	contentBytes, err := builtin.ReadFile(genDebugTalkPyPath)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	content := string(contentBytes)
	if !assert.Contains(t, content, "import funppy") {
		t.Fatal()
	}

	if !assert.Contains(t, content, "funppy.register") {
		t.Fatal()
	}

	reg, _ := regexp.Compile(`funppy\.register`)
	matchedSlice := reg.FindAllStringSubmatch(content, -1)
	if !assert.Len(t, matchedSlice, 10) {
		t.Fatal()
	}
}

func TestFindAllPythonFunctionNames(t *testing.T) {
	content := `
def test_1():	# exported function
    pass

def _test_2():	# exported function
    pass

def __test_3():	# private function
    pass

# def test_4():	# commented out function
#    pass

def Test5():	# exported function
    pass
`
	names, err := regexPyFunctionName.findAllFunctionNames(content)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	if !assert.Contains(t, names, "test_1") {
		t.FailNow()
	}
	if !assert.Contains(t, names, "Test5") {
		t.FailNow()
	}
	if !assert.Contains(t, names, "_test_2") {
		t.FailNow()
	}
	if !assert.NotContains(t, names, "__test_3") {
		t.FailNow()
	}
	// commented out function
	if !assert.NotContains(t, names, "test_4") {
		t.FailNow()
	}
}

func TestFindAllGoFunctionNames(t *testing.T) {
	content := `
func Test1() {	// exported function
	return
}

func testFunc2() {	// exported function
	return
}

func init() {	// private function
	return
}

func _Test3() { // exported function
	return
}

// func Test4() {	// commented out function
// 	return
// }
`
	names, err := regexGoFunctionName.findAllFunctionNames(content)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	if !assert.Contains(t, names, "Test1") {
		t.FailNow()
	}
	if !assert.Contains(t, names, "testFunc2") {
		t.FailNow()
	}
	if !assert.NotContains(t, names, "init") {
		t.FailNow()
	}
	if !assert.Contains(t, names, "_Test3") {
		t.FailNow()
	}
	// commented out function
	if !assert.NotContains(t, names, "Test4") {
		t.FailNow()
	}
}

func TestFindAllGoFunctionNamesAbnormal(t *testing.T) {
	content := `
func init() {	// private function
	return
}

func main() {	// should not define main() function
	return
}
`
	_, err := regexGoFunctionName.findAllFunctionNames(content)
	if !assert.NotNil(t, err) {
		t.FailNow()
	}
}
