package console

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

var bufferPool = &sync.Pool{
	New: func() any { return buffer{} },
}

var cwd, _ = os.Getwd()

// HandlerOptions are options for a ConsoleHandler.
// A zero HandlerOptions consists entirely of default values.
type HandlerOptions struct {
	// AddSource causes the handler to compute the source code position
	// of the log statement and add a SourceKey attribute to the output.
	AddSource bool

	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes LevelInfo.
	// The handler calls Level.Level for each record processed;
	// to adjust the minimum level dynamically, use a LevelVar.
	Level slog.Leveler
}

type ConsoleHandler struct {
	opts    *HandlerOptions
	out     io.Writer
	group   string
	context buffer
}

var _ slog.Handler = (*ConsoleHandler)(nil)

func NewHandler(out io.Writer, opts *HandlerOptions) *ConsoleHandler {
	if opts == nil {
		opts = new(HandlerOptions)
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	return &ConsoleHandler{
		opts:    opts,
		out:     out,
		group:   "",
		context: nil,
	}
}

// Enabled implements slog.Handler.
func (h *ConsoleHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.opts.Level.Level()
}

// Handle implements slog.Handler.
func (h *ConsoleHandler) Handle(_ context.Context, rec slog.Record) error {
	buf := bufferPool.Get().(buffer)

	buf.writeTimestamp(rec.Time)
	buf.writeLevel(rec.Level)
	if h.opts.AddSource && rec.PC > 0 {
		buf.writeSource(rec.PC, cwd)
	}
	buf.writeMessage(rec.Message)
	buf.copy(&h.context)
	rec.Attrs(func(a slog.Attr) bool {
		buf.writeAttr(a, h.group)
		return true
	})
	buf.NewLine()
	if _, err := buf.WriteTo(h.out); err != nil {
		buf.Reset()
		bufferPool.Put(buf)
		return err
	}
	bufferPool.Put(buf)
	return nil
}

// WithAttrs implements slog.Handler.
func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newCtx := h.context
	for _, a := range attrs {
		newCtx.writeAttr(a, h.group)
	}
	newCtx.Clip()
	return &ConsoleHandler{
		opts:    h.opts,
		out:     h.out,
		group:   h.group,
		context: newCtx,
	}
}

// WithGroup implements slog.Handler.
func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	if h.group != "" {
		name = h.group + "." + name
	}
	return &ConsoleHandler{
		opts:    h.opts,
		out:     h.out,
		group:   name,
		context: h.context,
	}
}
