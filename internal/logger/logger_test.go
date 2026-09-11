package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"
)

func TestLevelFromVerbosity(t *testing.T) {
	t.Parallel()

	tests := map[int]slog.Level{
		-1: slog.LevelInfo,
		0:  slog.LevelInfo,
		1:  slog.LevelDebug,
		2:  LevelTrace,
		9:  LevelTrace,
	}
	for verbosity, want := range tests {
		assert.Equal(t, want, LevelFromVerbosity(verbosity), "verbosity %d", verbosity)
	}
}

func TestHandlerFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{name: "info has no tag", level: slog.LevelInfo, want: "kuberlr: hello\n"},
		{name: "warn", level: slog.LevelWarn, want: "kuberlr: warning: hello\n"},
		{name: "error", level: slog.LevelError, want: "kuberlr: error: hello\n"},
		{name: "debug", level: slog.LevelDebug, want: "kuberlr: debug: hello\n"},
		{name: "trace", level: LevelTrace, want: "kuberlr: debug: hello\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := slog.New(NewHandler(&buf, LevelTrace, false))
			logger.Log(context.Background(), tt.level, "hello")
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestHandlerAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, slog.LevelInfo, false))

	logger.Info("downloading",
		"version", "1.31.4",
		"count", 3,
		"error", errors.New("boom happened"),
		"empty", "")

	assert.Equal(t,
		"kuberlr: downloading version=1.31.4 count=3 error=\"boom happened\" empty=\"\"\n",
		buf.String())
}

func TestHandlerWithAttrsAndGroup(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, slog.LevelInfo, false)).
		With("source", "client-go").
		WithGroup("req")

	logger.Info("done", "status", 200)

	assert.Equal(t, "kuberlr: done source=client-go req.status=200\n", buf.String())
}

func TestHandlerLevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, slog.LevelWarn, false))

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	assert.Equal(t, "kuberlr: warning: warn\nkuberlr: error: error\n", buf.String())
}

func TestHandlerColors(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, LevelTrace, true))

	logger.Warn("careful")
	assert.Equal(t, "kuberlr: "+text.FgYellow.EscapeSeq()+"warning:"+text.EscapeReset+" careful\n", buf.String())

	buf.Reset()
	logger.Error("bad")
	assert.Equal(t, "kuberlr: "+text.FgRed.EscapeSeq()+"error:"+text.EscapeReset+" bad\n", buf.String())

	buf.Reset()
	logger.Debug("detail")
	assert.Equal(t, "kuberlr: "+text.Faint.EscapeSeq()+"debug:"+text.EscapeReset+" detail\n", buf.String())

	buf.Reset()
	logger.Info("plain")
	assert.Equal(t, "kuberlr: plain\n", buf.String(), "info lines are never colored")
}

func TestLoggerShowProgress(t *testing.T) {
	t.Parallel()

	assert.True(t, newLogger(slog.DiscardHandler, true).ShowProgress())
	assert.False(t, newLogger(slog.DiscardHandler, false).ShowProgress())
	assert.False(t, Discard().ShowProgress())
}

func TestLoggerTrace(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	newLogger(NewHandler(&buf, LevelTrace, false), false).Trace("deep", "k", "v")
	assert.Equal(t, "kuberlr: debug: deep k=v\n", buf.String())

	buf.Reset()
	newLogger(NewHandler(&buf, slog.LevelDebug, false), false).Trace("hidden")
	assert.Empty(t, buf.String())
}

func TestFromContext(t *testing.T) {
	t.Parallel()

	log := Discard()
	assert.Same(t, log, FromContext(NewContext(context.Background(), log)))
	assert.NotNil(t, FromContext(context.Background()), "falls back to a discarding logger")
	assert.NotNil(t, FromContext(nil)) //nolint:staticcheck // nil context is exactly the case under test
}

func TestParseColorMode(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"auto", "Auto", " AUTO ", ""} {
		mode, err := ParseColorMode(input)
		require.NoError(t, err, input)
		assert.Equal(t, ColorAuto, mode, input)
	}

	mode, err := ParseColorMode("always")
	require.NoError(t, err)
	assert.Equal(t, ColorAlways, mode)

	mode, err = ParseColorMode("never")
	require.NoError(t, err)
	assert.Equal(t, ColorNever, mode)

	mode, err = ParseColorMode("rainbow")
	require.Error(t, err)
	assert.Equal(t, ColorAuto, mode, "falls back to auto")
}

func TestResolveColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	assert.True(t, ResolveColor(ColorAlways, false))
	assert.False(t, ResolveColor(ColorNever, true))
	assert.True(t, ResolveColor(ColorAuto, true))
	assert.False(t, ResolveColor(ColorAuto, false), "auto is off when stderr is not a tty")

	t.Setenv("NO_COLOR", "1")
	assert.False(t, ResolveColor(ColorAuto, true), "NO_COLOR disables auto")
	assert.True(t, ResolveColor(ColorAlways, true), "always ignores NO_COLOR")
}

// TestKlogRouting mutates klog's global logger, hence it is neither parallel
// nor split into subtests.
func TestKlogRouting(t *testing.T) { //nolint:paralleltest // mutates global klog state
	t.Cleanup(klog.ClearLogger)

	tests := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{
			name:  "default verbosity hides everything from client-go",
			level: slog.LevelInfo,
			want:  "",
		},
		{
			name:  "debug verbosity shows warnings and low V levels, capped at debug",
			level: slog.LevelDebug,
			want: "kuberlr: debug: something odd source=client-go\n" +
				"kuberlr: debug: failed source=client-go err=boom\n" +
				"kuberlr: debug: v1 detail source=client-go\n",
		},
		{
			name:  "trace verbosity also shows high V levels",
			level: LevelTrace,
			want: "kuberlr: debug: something odd source=client-go\n" +
				"kuberlr: debug: failed source=client-go err=boom\n" +
				"kuberlr: debug: v1 detail source=client-go\n" +
				"kuberlr: debug: v6 detail source=client-go\n",
		},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		routeKlog(NewHandler(&buf, tt.level, false), tt.level)

		klog.Warning("something odd")
		klog.ErrorS(errors.New("boom"), "failed")
		klog.V(1).Info("v1 detail")
		klog.V(6).Info("v6 detail")
		klog.Flush()

		assert.Equal(t, tt.want, buf.String(), tt.name)
	}
}
