package utils

import (
	"errors"
	"fmt"
)

func ErrorBuilder(errMsg string, err error) error {
	if err == nil {
		return errors.New(errMsg)
	}
	builderStr := fmt.Sprintf("%s : %s", errMsg, err.Error())
	return errors.New(builderStr)
}
