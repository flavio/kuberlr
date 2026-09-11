package logger

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/text"
)

// Prefix is prepended to every line written by the Handler.
const Prefix = "kuberlr: "

// Handler is a [slog.Handler] that renders records in a CLI-friendly format:
//
//	kuberlr: [<tag>: ]<message>[ key=value ...]
//
// where <tag> is "warning:", "error:" or "debug:" depending on the record level.
// Informational messages carry no tag. Tags are colored when colors are enabled.
type Handler struct {
	w     io.Writer
	mu    *sync.Mutex
	level slog.Leveler
	color bool

	// attrs and groups are inherited from WithAttrs/WithGroup calls.
	attrs  []slog.Attr
	groups []string
}

// NewHandler returns a Handler writing to w. Only records with a level greater
// than or equal to level are emitted. When color is true, tags are colored using
// ANSI escape sequences.
func NewHandler(w io.Writer, level slog.Leveler, color bool) *Handler {
	return &Handler{
		w:     w,
		mu:    &sync.Mutex{},
		level: level,
		color: color,
	}
}

// Enabled implements [slog.Handler].
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements [slog.Handler].
func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	var sb strings.Builder
	sb.WriteString(Prefix)
	sb.WriteString(h.tag(record.Level))
	sb.WriteString(record.Message)

	for _, attr := range h.attrs {
		h.writeAttr(&sb, attr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		h.writeAttr(&sb, h.qualify(attr))
		return true
	})
	sb.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, sb.String())
	return err
}

// WithAttrs implements [slog.Handler].
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	clone := *h
	clone.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	clone.attrs = append(clone.attrs, h.attrs...)
	for _, attr := range attrs {
		clone.attrs = append(clone.attrs, h.qualify(attr))
	}
	return &clone
}

// WithGroup implements [slog.Handler].
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.groups = make([]string, 0, len(h.groups)+1)
	clone.groups = append(clone.groups, h.groups...)
	clone.groups = append(clone.groups, name)
	return &clone
}

// tag returns the colored, trailing-space terminated tag for the given level.
// Info records have no tag.
func (h *Handler) tag(level slog.Level) string {
	var label string
	var color text.Color

	switch {
	case level >= slog.LevelError:
		label, color = "error:", text.FgRed
	case level >= slog.LevelWarn:
		label, color = "warning:", text.FgYellow
	case level >= slog.LevelInfo:
		return ""
	default:
		label, color = "debug:", text.Faint
	}

	if h.color {
		label = color.EscapeSeq() + label + text.EscapeReset
	}
	return label + " "
}

// qualify prefixes the attribute key with the currently open groups.
func (h *Handler) qualify(attr slog.Attr) slog.Attr {
	if len(h.groups) == 0 {
		return attr
	}
	attr.Key = strings.Join(h.groups, ".") + "." + attr.Key
	return attr
}

func (h *Handler) writeAttr(sb *strings.Builder, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		for _, sub := range attr.Value.Group() {
			if attr.Key != "" {
				sub.Key = attr.Key + "." + sub.Key
			}
			h.writeAttr(sb, sub)
		}
		return
	}

	sb.WriteByte(' ')
	sb.WriteString(attr.Key)
	sb.WriteByte('=')
	sb.WriteString(formatValue(attr.Value))
}

// formatValue renders a value, quoting strings that contain whitespace or
// quotes so that key=value pairs stay unambiguous.
func formatValue(value slog.Value) string {
	s := value.String()
	if s == "" || strings.ContainsAny(s, " \t\n\"") {
		return strconv.Quote(s)
	}
	return s
}
