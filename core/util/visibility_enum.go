package util

import (
	"encoding/json"
	"fmt"
)

type VisibilityEnum int

const (
	PublicVisibility  VisibilityEnum = 0
	PrivateVisibility VisibilityEnum = 1
)

func NewVisibilityEnum(value int) (VisibilityEnum, error) {
	switch value {
	case 0:
		return PublicVisibility, nil
	case 1:
		return PrivateVisibility, nil
	default:
		return 0, fmt.Errorf("invalid visibility value: %d", value)
	}
}

// UnmarshalJSON unmarshals the visibility enum, also checking if the value is valid
func (v *VisibilityEnum) UnmarshalJSON(data []byte) error {
	var value int
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}

	switch value {
	case 0:
		*v = PublicVisibility
	case 1:
		*v = PrivateVisibility
	default:
		return fmt.Errorf("invalid visibility value: %d", value)
	}

	return nil
}
