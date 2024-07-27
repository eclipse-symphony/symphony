package http

import (
	"time"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/valyala/fasthttp"
)

type Log struct {
	Observability observability.Observability
}

func (l Log) Log(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {

		actCtx := contexts.ParseActivityLogContextFromHttpRequestHeader(reqCtx)
		diagCtx := contexts.ParseDiagnosticLogContextFromHttpRequestHeader(reqCtx)
		ctx := composeCOARequestContext(reqCtx, actCtx, diagCtx)

		startTime := time.Now().UTC()

		logger.GetUserAuditsLogger().InfofCtx(ctx, "Request received: Method: %s URL: %s", reqCtx.Method(), reqCtx.Path())

		next(reqCtx)

		latency := time.Since(startTime).Seconds()

		logger.GetUserAuditsLogger().InfofCtx(ctx, "Request completed in %f seconds: Method: %s URL: %s StatusCode: %d", latency, reqCtx.Method(), reqCtx.Path(), reqCtx.Response.StatusCode())
	}
}
