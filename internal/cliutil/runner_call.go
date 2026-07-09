package cliutil

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"agent-telegram/internal/ipc"
)

func (r *Runner) recordCall(method string) {
	r.lastMethod = method
	r.lastSafety = operationSafety(method)
}

func (r *Runner) ensureServerReady(method string) bool {
	if err := r.ensureServer(); err != nil {
		if r.agentMode {
			r.handleError(r.ensureErrorToRPC(err, method))
			return false
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}
	return true
}

func (r *Runner) callRPC(method string, params any) (any, *ipc.ErrorObject, time.Duration) {
	start := time.Now()
	client := r.Client()

	var result any
	var err *ipc.ErrorObject
	if optionAware, ok := client.(optionsRPCClient); ok {
		result, err = optionAware.CallWithOptions(method, params, ipc.CallOptions{
			TraceID: r.traceID,
			RunID:   r.runID,
			Confirm: r.confirm,
		})
	} else if runAware, ok := client.(runRPCClient); ok {
		result, err = runAware.CallWithTraceAndRun(method, params, r.traceID, r.runID)
	} else if traced, ok := client.(traceRPCClient); ok {
		result, err = traced.CallWithTrace(method, params, r.traceID)
	} else {
		result, err = client.Call(method, params)
	}
	return result, err, time.Since(start)
}

func (r *Runner) logCallError(
	log *slog.Logger,
	method string,
	params any,
	err *ipc.ErrorObject,
	duration time.Duration,
) {
	log.Info("cli: call",
		"run_id", r.runID,
		"trace_id", r.traceID,
		"method", method,
		"params", truncateAny(params),
		"duration_ms", duration.Milliseconds(),
		"error_code", err.Code,
		"error_type", errorType(err),
		"error", err.Message,
	)
	r.writeAudit(method, params, nil, err, duration)
}

func (r *Runner) logCallSuccess(log *slog.Logger, method string, params, result any, duration time.Duration) {
	log.Info("cli: call",
		"run_id", r.runID,
		"trace_id", r.traceID,
		"method", method,
		"params", truncateAny(params),
		"duration_ms", duration.Milliseconds(),
		"status", "ok",
	)
	r.writeAudit(method, params, result, nil, duration)
}

func (r *Runner) applyResultFilters(result any) any {
	if len(r.filterExprs) > 0 {
		return r.filterExprs.Apply(result)
	}
	return result
}
