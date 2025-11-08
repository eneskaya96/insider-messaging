package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhoneNumber(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		wantError bool
	}{
		{
			name:      "valid turkish phone",
			phone:     "+905551234567",
			wantError: false,
		},
		{
			name:      "valid us phone",
			phone:     "+15551234567",
			wantError: false,
		},
		{
			name:      "empty phone",
			phone:     "",
			wantError: true,
		},
		{
			name:      "missing plus",
			phone:     "905551234567",
			wantError: true,
		},
		{
			name:      "invalid format",
			phone:     "+0551234567",
			wantError: true,
		},
		{
			name:      "too short",
			phone:     "+90",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhoneNumber(tt.phone)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, phone)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, phone)
				assert.Equal(t, tt.phone, phone.String())
			}
		})
	}
}

func TestPhoneNumberEquals(t *testing.T) {
	phone1, _ := NewPhoneNumber("+905551234567")
	phone2, _ := NewPhoneNumber("+905551234567")
	phone3, _ := NewPhoneNumber("+905559876543")

	assert.True(t, phone1.Equals(phone2))
	assert.False(t, phone1.Equals(phone3))
	assert.False(t, phone1.Equals(nil))
}
