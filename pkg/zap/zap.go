package zap

import (
	"flag"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Options contains all possible settings.
type Options struct {
	// Development configures the logger to use a Zap development config
	// (stacktraces on warnings, no sampling), otherwise a Zap production
	// config will be used (stacktraces on errors, sampling).
	Development bool `mapstructure:"zap-devel"`

	// Encoder configures how Zap will encode the output.  Defaults to
	// console when Development is true and JSON otherwise
	Encoder zapcore.Encoder

	// EncoderConfigOptions can modify the EncoderConfig needed to initialize an Encoder.
	// See https://pkg.go.dev/go.uber.org/zap/zapcore#EncoderConfig for the list of options
	// that can be configured.
	// Note that the EncoderConfigOptions are not applied when the Encoder option is already set.
	EncoderConfigOptions []crzap.EncoderConfigOption

	// NewEncoder configures Encoder using the provided EncoderConfigOptions.
	// Note that the NewEncoder function is not used when the Encoder option is already set.
	NewEncoder crzap.NewEncoderFunc `mapstructure:"zap-encoder"`

	// DestWriter controls the destination of the log output.  Defaults to
	// os.Stderr.
	DestWriter io.Writer

	// DestWritter controls the destination of the log output.  Defaults to
	// os.Stderr.
	//
	// Deprecated: Use DestWriter instead
	DestWritter io.Writer

	// Level configures the verbosity of the logging.
	// Defaults to Debug when Development is true and Info otherwise.
	// A zap log level should be multiplied by -1 to get the logr verbosity.
	// For example, to get logr verbosity of 3, set this field to zapcore.Level(-3).
	// See https://pkg.go.dev/github.com/go-logr/zapr for how zap level relates to logr verbosity.
	Level zapcore.LevelEnabler `mapstructure:"zap-log-level"`

	// StacktraceLevel is the level at and above which stacktraces will
	// be recorded for all messages. Defaults to Warn when Development
	// is true and Error otherwise.
	// See Level for the relationship of zap log level to logr verbosity.
	StacktraceLevel zapcore.LevelEnabler `mapstructure:"zap-stacktrace-level"`

	// ZapOpts allows passing arbitrary zap.Options to configure on the
	// underlying Zap logger.
	ZapOpts []zap.Option

	// TimeEncoder specifies the encoder for the timestamps in log messages.
	// Defaults to RFC3339TimeEncoder.
	TimeEncoder zapcore.TimeEncoder `mapstructure:"zap-time-encoding"`
}

func UseFlagOptions(in *Options) crzap.Opts {
	if err := viper.Unmarshal(in, viper.DecodeHook(
		zapHook(),
	)); err != nil {
		panic(fmt.Errorf("unmarshal zap config: %w", err))
	}

	return func(o *crzap.Options) {
		*o = crzap.Options(*in)
	}
}

func (o *Options) BindFlags(fs *flag.FlagSet) {
	zOpts := crzap.Options{}
	zOpts.BindFlags(fs)

	*o = Options(zOpts)
}

func New(opts ...crzap.Opts) logr.Logger {
	return zapr.NewLogger(crzap.NewRaw(opts...))
}

var (
	levelEnablerType   = reflect.TypeOf((*zapcore.LevelEnabler)(nil)).Elem()
	newEncoderFuncType = reflect.TypeOf((*crzap.NewEncoderFunc)(nil)).Elem()
)

/*
Following 2 encoder functions are copied from ControllerRuntime zap package.
We only set the EncoderFunc in the Hook function below and not initialize the Encoder.
This is done to ensure the TimeEncoder (passed in via env or the default) is used
while creating the Encoder in the ControllerRuntime code.
*/
func newConsoleEncoder(opts ...crzap.EncoderConfigOption) zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	for _, opt := range opts {
		opt(&encoderConfig)
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func newJSONEncoder(opts ...crzap.EncoderConfigOption) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	for _, opt := range opts {
		opt(&encoderConfig)
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

var levelStrings = map[string]zapcore.Level{
	"debug": zap.DebugLevel,
	"info":  zap.InfoLevel,
	"error": zap.ErrorLevel,
	"panic": zap.PanicLevel,
}

func zapHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		stringToLevelEnablerHookFunc(),
		stringToNewEncoderFuncHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
	)
}

func stringToLevelEnablerHookFunc() mapstructure.DecodeHookFuncType {
	return func(in reflect.Type, out reflect.Type, val interface{}) (interface{}, error) {
		if in.Kind() == reflect.String && out == levelEnablerType {
			sVal := val.(string)
			if sVal == "" {
				var v zapcore.LevelEnabler
				// return nil if level is not set; controller-runtime sets the default value
				return &v, nil
			}

			// level supports setting of integer value > 0 in addition to `info`, `error` or `debug`
			level, validLevel := levelStrings[strings.ToLower(sVal)]
			if !validLevel {
				logLevel, err := strconv.Atoi(sVal)
				if err != nil {
					return nil, fmt.Errorf("invalid log level \"%s\"", val)
				}

				if logLevel > 0 {
					intLevel := -1 * logLevel
					return zap.NewAtomicLevelAt(zapcore.Level(int8(intLevel))), nil
				} else {
					return nil, fmt.Errorf("invalid log level \"%s\"", val)
				}
			}

			return zap.NewAtomicLevelAt(level), nil
		}

		return val, nil
	}
}

func stringToNewEncoderFuncHookFunc() mapstructure.DecodeHookFuncType {
	return func(in reflect.Type, out reflect.Type, val interface{}) (interface{}, error) {
		if in.Kind() == reflect.String && out == newEncoderFuncType {
			// TODO: implement encoding.TextUnmarshaler interface for type NewEncoderFunc upstream
			var encoder crzap.NewEncoderFunc

			switch val.(string) {
			case "":
				// return nil if encoder is not set; controller-runtime sets the default value
				return encoder, nil
			case "console":
				encoder = newConsoleEncoder
			case "json":
				encoder = newJSONEncoder
			default:
				return nil, fmt.Errorf("invalid encoder value \"%s\"", val)
			}

			return encoder, nil
		}

		return val, nil
	}
}
