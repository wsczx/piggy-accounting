package base

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志配置
type LogConfig struct {
	Level      string `yaml:"level" mapstructure:"level" comment:"日志级别: debug, info, warn, error, fatal"`
	Output     string `yaml:"output" mapstructure:"output" comment:"输出方式: stdout 或 file"`
	File       string `yaml:"file" mapstructure:"file" comment:"日志文件路径（当 output=file 时生效）"`
	Color      bool   `yaml:"color" mapstructure:"color" comment:"日志是否启用颜色"`
	Format     string `yaml:"format" mapstructure:"format" comment:"日志格式: text 或 json"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size" comment:"单个日志文件最大大小(MB)"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups" comment:"保留的历史日志文件数量"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age" comment:"日志保留天数"`
	Compress   bool   `yaml:"compress" mapstructure:"compress" comment:"是否压缩旧日志文件"`
	AddSource  bool   `yaml:"add_source" mapstructure:"add_source" comment:"是否输出日志调用源文件与行号"`
}

type Logger struct {
	handler slog.Handler
	level   *slog.LevelVar
	file    *lumberjack.Logger
}

var (
	logger = &Logger{
		handler: slog.NewTextHandler(os.Stdout, nil),
		level:   &slog.LevelVar{},
	}
)

// 初始化或重新配置日志
func InitLog(c LogConfig) {
	newLogger := NewLoggerWithConfig(c)
	logger = newLogger
	slog.SetDefault(slog.New(newLogger.handler))
}
func NewLoggerWithConfig(cfg LogConfig) *Logger {
	lVar := &slog.LevelVar{}
	lVar.Set(parseLevel(cfg.Level))

	var out io.Writer
	var ljFile *lumberjack.Logger

	// 确定输出源
	if strings.ToLower(cfg.Output) == "file" {
		ljFile = &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		out = ljFile
	} else {
		out = os.Stdout
	}

	opts := slog.HandlerOptions{
		Level:     lVar,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 当 Key 是 source 时，精简路径
			if a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if !ok {
					return a
				}
				// 只保留最后两级
				parts := strings.Split(source.File, "/")
				if len(parts) > 2 {
					source.File = strings.Join(parts[len(parts)-2:], "/")
				}

				// 清空function 字段
				source.Function = ""
				return slog.Any(a.Key, source)
			}
			return a
		},
	}
	var handler slog.Handler
	// Text格式美化显示
	if strings.ToLower(cfg.Format) == FormatText {
		// color := true
		// // 如果输出目标是文件，则禁用颜色
		// if strings.ToLower(cfg.Output) == "file" {
		// 	color = cfg.FileColor
		// }
		handler = &ConsoleHandler{
			opts:  opts,
			out:   out,
			color: cfg.Color,
		}
	} else {
		// json格式使用标准json输出
		handler = slog.NewJSONHandler(out, &opts)
	}

	return &Logger{handler: handler, level: lVar, file: ljFile}
}

func internalLog(ctx context.Context, level slog.Level, skip int, msg string, args ...any) {
	// 支持 With 属性
	if !logger.handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// 使用传入的 skip
	runtime.Callers(skip, pcs[:])

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	if len(args) > 0 {
		if len(args)%2 != 0 {
			// 如果剩下的参数是奇数，默认Key为空
			r.Add(args[:len(args)-1]...)
			r.Add("", args[len(args)-1]) // 最后一个参数Key为空
		} else {
			r.Add(args...)
		}
	}
	logger.handler.Handle(ctx, r)
}
func handleSmartLog(level slog.Level, args ...any) {
	if len(args) == 0 {
		return
	}

	// var msg string
	// var attrs []any

	// // 如果参数个数是偶数，说明是全 key-value 模式，补一个空标题
	// // Handle 里可以通过判断 msg == "" 来跳过紫色标题渲染
	// if len(args)%2 == 0 {
	// 	msg = ""
	// 	attrs = args
	// } else {
	// 	// 如果参数个数是奇数个，则第一个是标题
	// 	msg = fmt.Sprint(args[0])
	// 	attrs = args[1:]
	// }
	var msg string
	var attrs []any

	if len(args) == 1 {
		// 只有一个参数：视为普通消息，不带特殊渲染
		msg = fmt.Sprint(args[0])
	} else {
		// 两个及以上：第一个是标题（紫色），其余是 KV
		msg = fmt.Sprint(args[0])
		attrs = args[1:]
	}

	internalLog(context.Background(), level, 4, msg, attrs...)
}

func Debug(args ...any) { handleSmartLog(LevelDebug, args...) }
func Info(args ...any)  { handleSmartLog(LevelInfo, args...) }
func Warn(args ...any)  { handleSmartLog(LevelWarn, args...) }
func Error(args ...any) { handleSmartLog(LevelError, args...) }
func Fatal(args ...any) {
	handleSmartLog(LevelFatal, args...)
	os.Exit(1)
}

func Debugf(f string, a ...any) {
	internalLog(context.Background(), LevelDebug, 3, fmt.Sprintf(f, a...))
}
func Infof(f string, a ...any) {
	internalLog(context.Background(), LevelInfo, 3, fmt.Sprintf(f, a...))
}
func Warnf(f string, a ...any) {
	internalLog(context.Background(), LevelWarn, 3, fmt.Sprintf(f, a...))
}
func Errorf(f string, a ...any) {
	internalLog(context.Background(), LevelError, 3, fmt.Sprintf(f, a...))
}

func Fatalf(f string, a ...any) {
	internalLog(context.Background(), LevelFatal, 3, fmt.Sprintf(f, a...))
	os.Exit(1)
}

func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

func SetLevel(levelStr string) {
	logger.level.Set(parseLevel(levelStr))
}

func Close() error {
	if logger != nil && logger.file != nil {
		return logger.file.Close()
	}
	return nil
}

func parseLevel(l string) slog.Level {
	switch strings.ToUpper(l) {
	case "DEBUG":
		return LevelDebug
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}
