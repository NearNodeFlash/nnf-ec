package elementcontroller

import (
	"fmt"
	"net/http"
)

type ControllerError struct {
	StatusCode  int
	ErrorString string
}

func NewControllerError(sc int, es string) *ControllerError {
	return &ControllerError{StatusCode: sc, ErrorString: es}
}

func (err *ControllerError) Error() string {
	return fmt.Sprintf("Error %d: %s", err.StatusCode, err.ErrorString)
}

var (
	ErrNotFound = NewControllerError(http.StatusNotFound, "Not Found")
)
