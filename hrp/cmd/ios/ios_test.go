//go:build localtest

package ios

func TestGetDevice(t *testing.T) {
	device, err := getDevice(udid)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("device: %v", device)
}
