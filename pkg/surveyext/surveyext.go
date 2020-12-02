package surveyext

import (
	"github.com/AlecAivazis/survey/v2"
)

// AskConfirm asks for confirmation on the terminal.
func AskConfirm(msg string, def bool, opts ...survey.AskOpt) (res bool, err error) {
	err = survey.AskOne(&survey.Confirm{
		Message: msg,
		Default: def,
	}, &res, opts...)
	return
}
