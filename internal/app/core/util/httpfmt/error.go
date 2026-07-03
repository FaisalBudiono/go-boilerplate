package httpfmt

import (
	"fmt"
	"strings"

	"komdigi-immigration/internal/app/core/util/app"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/domain/errcode"

	"go.opentelemetry.io/otel/trace"
)

type ErrResponse struct {
	Msg     string       `json:"message"`
	ErrCode errcode.Code `json:"errorCode"`
	TraceID string       `json:"traceID,omitempty"`
}

type errOpts struct {
	TraceID string
}

func newErrOpts() *errOpts {
	return &errOpts{}
}

type Option func(*errOpts)

func WithTraceID(trace trace.Span) Option {
	return func(o *errOpts) {
		o.TraceID = trace.SpanContext().TraceID().String()
	}
}

func NewError(ec errcode.Code, msg string, opts ...Option) ErrResponse {
	cfg := newErrOpts()
	for _, opt := range opts {
		opt(cfg)
	}

	return ErrResponse{
		Msg:     msg,
		ErrCode: ec,
		TraceID: cfg.TraceID,
	}
}

func NewErrorGeneric(err error, opts ...Option) ErrResponse {
	cfg := newErrOpts()
	for _, opt := range opts {
		opt(cfg)
	}

	if app.ENV().Log.Level == app.LogLevelDebug {
		return ErrResponse{
			Msg:     err.Error(),
			ErrCode: errcode.Generic,
			TraceID: cfg.TraceID,
		}
	}

	return ErrResponse{
		Msg:     "Something wrong in the server",
		ErrCode: errcode.Generic,
		TraceID: cfg.TraceID,
	}
}

type errMetaValue struct {
	Code invalid.Code `json:"code"`
	Msg  string       `json:"message,omitempty"`
}

type errMeta map[string][]errMetaValue

func NewUnprocessableErr(opts ...Option) *UnprocessableErr {
	cfg := newErrOpts()
	for _, opt := range opts {
		opt(cfg)
	}

	return &UnprocessableErr{
		ErrResponse: ErrResponse{
			Msg:     "Structure body/param might be invalid.",
			ErrCode: errcode.InvalidParam,
			TraceID: cfg.TraceID,
		},
		Meta: make(errMeta),
	}
}

type UnprocessableErr struct {
	ErrResponse

	Meta errMeta `json:"meta"`
}

func (e *UnprocessableErr) Error() string {
	return fmt.Sprintf("Invalid param request, meta: %#v", e.Meta)
}

func (e *UnprocessableErr) IsError() bool {
	return len(e.Meta) > 0
}

// Add meta to error
func (e *UnprocessableErr) Add(key string, code invalid.Code, msg string) {
	e.Meta[key] = append(e.Meta[key], errMetaValue{Code: code, Msg: msg})
}

func EnumOptions[T ~string](raws []T) string {
	strs := make([]string, len(raws))
	for i, raw := range raws {
		strs[i] = string(raw)
	}
	return strings.Join(strs, ",")
}
