package e

import (
	"fmt"
)

func Wrap(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}

func WrapIfError(message string, err error) error {
	if err == nil {
		return nil
	}

	return Wrap(message, err)
}
