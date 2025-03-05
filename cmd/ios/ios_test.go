//go:build localtest

package ios

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDevice(t *testing.T) {
	device, err := getDevice(udid)
	require.Nil(t, err)
	t.Logf("device: %v", device)
}
