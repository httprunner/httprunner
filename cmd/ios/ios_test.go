//go:build localtest

package ios

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDevice(t *testing.T) {
	t.Skip("Skip iOS test - requires physical iOS device")
	device, err := getDevice(udid)
	require.Nil(t, err)
	t.Logf("device: %v", device)
}
