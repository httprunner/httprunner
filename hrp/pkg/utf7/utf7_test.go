package utf7

import "testing"

func Test_Decode(t *testing.T) {
	str, err := Encoding.NewDecoder().String("&j71bgXcBbIiWM14CZbBsEV4CbBFlz4hX-36-4")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}
