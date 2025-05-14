package builtin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterface2Float64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{
			name:    "convert int",
			input:   42,
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert int32",
			input:   int32(42),
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert int64",
			input:   int64(42),
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert float32",
			input:   float32(42.5),
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert float64",
			input:   42.5,
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert string valid number",
			input:   "42.5",
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert string valid number",
			input:   "425",
			want:    425.0,
			wantErr: false,
		},
		{
			name:    "convert string invalid number",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "convert json.Number valid",
			input:   json.Number("42.5"),
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert json.Number invalid",
			input:   json.Number("invalid"),
			want:    0,
			wantErr: true,
		},
		{
			name:    "convert unsupported type",
			input:   []int{1, 2, 3},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interface2Float64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
