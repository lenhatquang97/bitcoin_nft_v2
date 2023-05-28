package utils

import "errors"

func WrapperError(errStr string) error {
	return errors.New(errStr)
}
