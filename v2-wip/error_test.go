package atlas

import (
	"testing"
)

func TestError_impl(t *testing.T) {
	var _ error = new(ErrorDoc)
	var _ error = new(Error)
}
