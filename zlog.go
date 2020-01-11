package zlog

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/mulinbc/zerr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 表示 zap.Logger 初始化所需要的参数。
type Logger struct {
	DevMode    bool   `json:"dev_mode"`
	Level      string `json:"level" validate:"oneof=debug info warn error dpanic panic fatal"`
	Filename   string `json:"filename" validate:"required"`
	MaxSize    int    `json:"max_size" validate:"gte=0"`
	MaxAge     int    `json:"max_age" validate:"gte=0"`
	MaxBackups int    `json:"max_backups" validate:"gte=0"`
	LocalTime  bool   `json:"local_time"`
	Compress   bool   `json:"compress"`
}

var level = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

// New 返回一个 zap.Logger 实例。
func New(filename, level string, maxSize, maxAge, maxBackups, skip int, devMode, localTime, compress bool) (*zap.Logger, error) {
	l := Logger{
		DevMode:    devMode,
		Level:      level,
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		LocalTime:  localTime,
		Compress:   compress,
	}

	return l.New(skip)
}

// New 返回一个 zap.Logger 实例。
func (p *Logger) New(skip int) (*zap.Logger, error) {
	if err := validator.New().Struct(p); err != nil {
		return nil, zerr.Wrap(err)
	}
	return p.new(skip), nil
}

func (p *Logger) new(skip int) *zap.Logger {
	w := &lumberjack.Logger{
		Filename:   p.Filename,
		MaxSize:    p.MaxSize,
		MaxAge:     p.MaxAge,
		MaxBackups: p.MaxBackups,
		LocalTime:  p.LocalTime,
		Compress:   p.Compress,
	}

	if p.DevMode {
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(w)),
			zapcore.DebugLevel,
		)
		return zap.New(core, zap.Development(), zap.AddCaller(), zap.AddCallerSkip(skip), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(w)),
		level[p.Level],
	)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skip), zap.AddStacktrace(zapcore.ErrorLevel))
}
