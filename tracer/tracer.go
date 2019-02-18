package tracer

import (
	"context"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"go.opencensus.io/trace"
)

// StartTraceEvent .. comment :)
func StartSpanTrace(ctx context.Context, message string, data log.Data) *trace.Span {
	_, span := trace.StartSpan(ctx, message)
	if span == nil {
		// TODO ....something!
	}

	if message != "" {
		span.AddAttributes(trace.StringAttribute("msg", message))
	}

	requestID := common.GetRequestId(ctx)
	span.AddAttributes(trace.StringAttribute("requestID", requestID))

	// Create span attributes from the data struct
	for k, v := range data {
		switch v := v.(type) {
		case string:
			span.AddAttributes(trace.StringAttribute(k, v))
		}
	}

	return span
}
