package storage

import (
	"errors"
)

func ValidateKey(key string) error {
	if key == "" {
		return errors.New("Empty key provided")
	}

	return nil
}
