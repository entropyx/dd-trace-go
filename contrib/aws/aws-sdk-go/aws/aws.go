// Package aws provides functions to trace aws/aws-sdk-go (https://github.com/aws/aws-sdk-go).
package aws // import "github.com/entropyx/dd-trace-go/contrib/aws/aws-sdk-go/aws"

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/entropyx/dd-trace-go/ddtrace/ext"
	"github.com/entropyx/dd-trace-go/ddtrace/tracer"
)

const (
	tagAWSAgent     = "aws.agent"
	tagAWSOperation = "aws.operation"
	tagAWSRegion    = "aws.region"
)

type handlers struct {
	cfg *config
}

// WrapSession wraps a session.Session, causing requests and responses to be traced.
func WrapSession(s *session.Session, opts ...Option) *session.Session {
	cfg := new(config)
	for _, opt := range opts {
		opt(cfg)
	}
	h := &handlers{cfg: cfg}
	s = s.Copy()
	s.Handlers.Send.PushFrontNamed(request.NamedHandler{
		Name: "github.com/entropyx/dd-trace-go/contrib/aws/aws-sdk-go/aws/handlers.Send",
		Fn:   h.Send,
	})
	s.Handlers.Complete.PushBackNamed(request.NamedHandler{
		Name: "github.com/entropyx/dd-trace-go/contrib/aws/aws-sdk-go/aws/handlers.Complete",
		Fn:   h.Complete,
	})
	return s
}

func (h *handlers) Send(req *request.Request) {
	_, ctx := tracer.StartSpanFromContext(req.Context(), h.operationName(req),
		tracer.SpanType(ext.SpanTypeHTTP),
		tracer.ServiceName(h.serviceName(req)),
		tracer.ResourceName(h.resourceName(req)),
		tracer.Tag(tagAWSAgent, h.awsAgent(req)),
		tracer.Tag(tagAWSOperation, h.awsOperation(req)),
		tracer.Tag(tagAWSRegion, h.awsRegion(req)),
		tracer.Tag(ext.HTTPMethod, req.Operation.HTTPMethod),
		tracer.Tag(ext.HTTPURL, req.HTTPRequest.URL.String()),
	)
	req.SetContext(ctx)
}

func (h *handlers) Complete(req *request.Request) {
	span, ok := tracer.SpanFromContext(req.Context())
	if !ok {
		return
	}
	if req.HTTPResponse != nil {
		span.SetTag(ext.HTTPCode, strconv.Itoa(req.HTTPResponse.StatusCode))
	}
	span.Finish(tracer.WithError(req.Error))
}

func (h *handlers) operationName(req *request.Request) string {
	return h.awsService(req) + ".command"
}

func (h *handlers) resourceName(req *request.Request) string {
	return h.awsService(req) + "." + req.Operation.Name
}

func (h *handlers) serviceName(req *request.Request) string {
	if h.cfg.serviceName != "" {
		return h.cfg.serviceName
	}
	return "aws." + h.awsService(req)
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

func (h *handlers) awsOperation(req *request.Request) string {
	return req.Operation.Name
}

func (h *handlers) awsRegion(req *request.Request) string {
	return req.ClientInfo.SigningRegion
}

func (h *handlers) awsService(req *request.Request) string {
	return req.ClientInfo.ServiceName
}
