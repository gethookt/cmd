package errors // import "hookt.dev/cmd/pkg/errors"

import (
	"errors"
	"fmt"
	"runtime"
)

type E struct {
	Err error
	pc  []uintptr
}

func (e *E) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *E) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(format string, args ...any) error {
	e := &E{
		Err: fmt.Errorf(format, args...),
		pc:  make([]uintptr, 10),
	}

	runtime.Callers(2, e.pc)

	return e
}

func Join(err ...error) error {
	return errors.Join(err...)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
