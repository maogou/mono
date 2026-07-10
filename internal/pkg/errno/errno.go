package errno

import (
	"fmt"
)

type ErrNo struct {
	Code    int
	Message string
}

func (eo ErrNo) Error() string {
	return eo.Message
}

type Err struct {
	Code    int
	Message string
	Err     error
}

func NewErr(eo *ErrNo, err error) *Err {
	return &Err{
		Code:    eo.Code,
		Message: eo.Message,
		Err:     err,
	}
}

func (err *Err) Add(message string) error {
	err.Message += " " + message

	return err
}

func (err *Err) AddFormat(format string, args ...any) error {
	err.Message += " " + fmt.Sprintf(format, args...)

	return err
}

func (err *Err) Error() string {
	return fmt.Sprintf("Err - code: %d, message: %s, error: %s", err.Code, err.Message, err.Err)
}

func DecodeErr(err error) (int, string) {
	if err == nil {
		return Ok.Code, Ok.Message
	}

	switch e := err.(type) {
	case *Err:
		return e.Code, e.Message
	case *ErrNo:
		return e.Code, e.Message
	default:
		return InternalServerError.Code, InternalServerError.Message
	}
}
