package opentracer

import (
	"testing"

	"github.com/entropyx/dd-trace-go/ddtrace"
	"github.com/entropyx/dd-trace-go/ddtrace/internal"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	assert := assert.New(t)
	ot := New()
	dd, ok := internal.GetGlobalTracer().(ddtrace.Tracer)
	assert.True(ok)
	ott, ok := ot.(*opentracer)
	assert.True(ok)
	assert.Equal(ott.Tracer, dd)
}
