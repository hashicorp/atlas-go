package atlas

import (
	"fmt"
	"strings"
)

// ErrorDoc is the wrapper document for an erroneous response. This can
// be parsed directly as a JSON document (not a JSON API resource).
//
// This also implements the `error` interface so it can be used directly.
type ErrorDoc struct {
	Errors []*Error `json:"errors"`
}

// error impl.
func (e *ErrorDoc) Error() string {
	errs := make([]string, len(e.Errors))
	for i, e := range e.Errors {
		errs[i] = e.Error()
	}

	return strings.Join(errs, "\n")
}

// Error is a single error of an error document.
//
// This also implements the `error` interface so it can be used directly.
type Error struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
	Status string `json:"status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}
