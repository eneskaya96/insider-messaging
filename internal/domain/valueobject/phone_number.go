package valueobject

import (
	"fmt"
	"regexp"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(phone string) (*PhoneNumber, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone number cannot be empty")
	}

	if !phoneRegex.MatchString(phone) {
		return nil, fmt.Errorf("invalid phone number format: must start with + and contain country code")
	}

	return &PhoneNumber{value: phone}, nil
}

func (p *PhoneNumber) String() string {
	return p.value
}

func (p *PhoneNumber) Equals(other *PhoneNumber) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}
