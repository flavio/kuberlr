package logger

import (
	"context"
	"flag"
	"log/slog"
	"strconv"

	"k8s.io/klog/v2"
)

// clientGoSource is the value of the "source" attribute attached to messages
// coming from client-go.
const clientGoSource = "client-go"

// routeKlog makes the kubernetes client-go library, which logs through
// k8s.io/klog/v2, emit its messages through the given handler.
//
// client-go is chatty about details that are of no interest to kubectl users
// (e.g. warnings about a missing kubeconfig that kubectl itself is going to
// report anyway). Its messages are therefore capped at debug level: they are
// only visible when the verbosity is raised.
func routeKlog(handler slog.Handler, level slog.Level) {
	// klog gates klog.V(n) calls on its own global verbosity flag before the
	// message ever reaches the logger. slog levels below Info are negative
	// (Debug = -4, Trace = -8) and klog.V(n) is forwarded as slog level -n, so
	// -level is exactly the klog verbosity that lets through everything the
	// handler is going to accept.
	klogVerbosity := max(0, -int(level))
	fs := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(fs)
	// The flag is guaranteed to exist and the value is a valid integer.
	_ = fs.Set("v", strconv.Itoa(klogVerbosity))

	capped := &capLevelHandler{inner: handler, maxLevel: slog.LevelDebug}
	klog.SetSlogLogger(slog.New(capped).With("source", clientGoSource))
}

// capLevelHandler is a [slog.Handler] decorator that lowers the level of every
// record to at most maxLevel before forwarding it to the inner handler.
type capLevelHandler struct {
	inner    slog.Handler
	maxLevel slog.Level
}

func (h *capLevelHandler) cap(level slog.Level) slog.Level {
	return min(level, h.maxLevel)
}

// Enabled implements [slog.Handler].
func (h *capLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, h.cap(level))
}

// Handle implements [slog.Handler].
//
// The level check is repeated here because the logr -> slog bridge used by
// klog calls Handle for error records without consulting Enabled first.
func (h *capLevelHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Level = h.cap(record.Level)
	if !h.inner.Enabled(ctx, record.Level) {
		return nil
	}
	return h.inner.Handle(ctx, record)
}

// WithAttrs implements [slog.Handler].
func (h *capLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &capLevelHandler{inner: h.inner.WithAttrs(attrs), maxLevel: h.maxLevel}
}

// WithGroup implements [slog.Handler].
func (h *capLevelHandler) WithGroup(name string) slog.Handler {
	return &capLevelHandler{inner: h.inner.WithGroup(name), maxLevel: h.maxLevel}
}
