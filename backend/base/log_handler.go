package base

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
)

// 实现 slog.Handler 接口，用于日志的染色输出
type ConsoleHandler struct {
	opts  slog.HandlerOptions
	out   io.Writer
	attrs []slog.Attr // 用于存储 With 注入的属性
	color bool        // 是否启用颜色输出
}

func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *ConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level
	levelStr := level.String()

	// 级别染色：保持 Level 颜色
	color := ColorReset
	switch {
	case level >= LevelFatal:
		color = ColorRed + ColorBold
		levelStr = "🔥 FATAL"
	case level >= LevelError:
		color = ColorRed
		levelStr = "ERROR"
	case level >= LevelWarn:
		color = ColorYellow
		levelStr = "WARN"
	case level >= LevelInfo:
		color = ColorGreen
		levelStr = "INFO"
	case level >= LevelDebug:
		color = ColorBlue
		levelStr = "DEBUG"
	}

	// 基础头部：时间 + 级别
	timeStr := r.Time.Format("2006-01-02 15:04:05")
	fmt.Fprintf(h.out, "%s %s ", timeStr, h.colorize(color, fmt.Sprintf("%-5s", levelStr)))

	// 行号/路径：使用青色 + 方括号
	if h.opts.AddSource {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			parts := strings.Split(f.File, "/")
			source := f.File
			if len(parts) > 2 {
				source = strings.Join(parts[len(parts)-2:], "/")
			}
			// 青色显示行号
			sourceText := fmt.Sprintf("[%s:%d]", source, f.Line)
			fmt.Fprintf(h.out, "%s ", h.colorize(ColorCyan, sourceText))
		}
	}
	// 收集所有属性 (With 的属性 + 当前行的属性)
	var allAttrs []slog.Attr
	allAttrs = append(allAttrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

	// 打印标题 (Message)：
	hasCurrentAttrs := r.NumAttrs() > 0
	if r.Message != "" {
		if hasCurrentAttrs || len(h.attrs) > 0 {
			// 标题用 浅紫色，箭头用 灰色
			msgPart := h.colorize(ColorLavender, r.Message)
			arrowPart := h.colorize(ColorDim, "»")
			fmt.Fprintf(h.out, "%s %s ", msgPart, arrowPart)
		} else {
			fmt.Fprintf(h.out, "%s ", r.Message)
		}
	}
	// 打印 Key-Value 参数
	for i, a := range allAttrs {
		// 如果有标题，或者这不是第一个参数，就补空格
		if r.Message != "" || i > 0 {
			fmt.Fprint(h.out, " ")
		}

		// 判定是否是背景参数（With 里的）
		isWith := i < len(h.attrs)
		c := ColorBlue
		if isWith {
			c = ColorDim
		}
		fmt.Fprintf(h.out, "%s %v", h.colorize(c, a.Key), a.Value.Any())
	}

	fmt.Fprint(h.out, "\n")
	return nil
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return &newHandler
}
func (h *ConsoleHandler) colorize(color string, text string) string {
	if !h.color || color == "" {
		return text
	}
	return color + text + ColorReset
}
func (h *ConsoleHandler) WithGroup(name string) slog.Handler { return h }
