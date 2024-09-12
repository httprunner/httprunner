package uixt

import (
	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp/code"
)

func (dExt *DriverExt) Input(text string) (err error) {
	err = dExt.Driver.Input(text)
	if err != nil {
		return errors.Wrap(code.MobileUIInputError, err.Error())
	}
	return nil
}
