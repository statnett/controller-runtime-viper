package zap

import (
	"bytes"
	"encoding/json"
	"flag"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zap Log Suite")
}

const testMessage = "This is a test message"

var _ = Describe("Zap log level flag options setup", func() {
	var (
		fs             flag.FlagSet
		pfs            pflag.FlagSet
		logInfoLevel0  = "info text"
		logDebugLevel1 = "debug 1 text"
		logDebugLevel2 = "debug 2 text"
		logDebugLevel3 = "debug 3 text"
		opts           = Options{}
	)

	BeforeEach(func() {
		fs = *flag.NewFlagSet("read from env", flag.ExitOnError)
		pfs = *pflag.NewFlagSet("read from env", pflag.ExitOnError)

		opts.BindFlags(&fs)

		pfs.AddGoFlagSet(&fs)
		err := viper.BindPFlags(&pfs)
		Expect(err).ToNot(HaveOccurred())

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	})

	Context("with  zap-log-level options provided", func() {
		It("Should output logs for info and debug zap-log-level.", Label("loglevel"), func() {
			GinkgoT().Setenv("ZAP_LOG_LEVEL", "debug")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.V(0).Info(logInfoLevel0)
			logger.V(1).Info(logDebugLevel1)

			outRaw := logOut.Bytes()

			Expect(string(outRaw)).Should(ContainSubstring(logInfoLevel0))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel1))
		})

		It("Should output only error logs, otherwise empty logs", Label("loglevel"), func() {
			GinkgoT().Setenv("ZAP_LOG_LEVEL", "error")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.V(0).Info(logInfoLevel0)
			logger.V(1).Info(logDebugLevel1)

			outRaw := logOut.Bytes()

			Expect(outRaw).To(BeEmpty())
		})
	})

	Context("with  zap-log-level  with increased verbosity.", func() {
		It("Should output debug and info log, with default production mode.", Label("loglevel"), func() {
			GinkgoT().Setenv("ZAP_LOG_LEVEL", "1")
			GinkgoT().Setenv("ZAP_DEVEL", "false")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.V(0).Info(logInfoLevel0)
			logger.V(1).Info(logDebugLevel1)

			outRaw := logOut.Bytes()

			Expect(string(outRaw)).Should(ContainSubstring(logInfoLevel0))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel1))
		})

		It("Should output info and debug logs, with development mode.", func() {
			GinkgoT().Setenv("ZAP_LOG_LEVEL", "1")
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.V(0).Info(logInfoLevel0)
			logger.V(1).Info(logDebugLevel1)

			outRaw := logOut.Bytes()

			Expect(string(outRaw)).Should(ContainSubstring(logInfoLevel0))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel1))
		})

		It("Should output info, and debug logs with increased verbosity, and with development mode set to true.", Label("level3"), func() {
			GinkgoT().Setenv("ZAP_LOG_LEVEL", "3")
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.V(0).Info(logInfoLevel0)
			logger.V(1).Info(logDebugLevel1)
			logger.V(2).Info(logDebugLevel2)
			logger.V(3).Info(logDebugLevel3)

			outRaw := logOut.Bytes()

			Expect(string(outRaw)).Should(ContainSubstring(logInfoLevel0))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel1))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel2))
			Expect(string(outRaw)).Should(ContainSubstring(logDebugLevel3))
		})
	})

	Context("with zap-time-encoding flag provided", Label("timeencoder"), func() {
		It("Should set time encoder in options", func() {
			GinkgoT().Setenv("ZAP_TIME_ENCODING", "rfc3339")

			opt := crzap.Options{}
			UseFlagOptions(&opts)(&opt)

			optVal := reflect.ValueOf(opt.TimeEncoder)
			expVal := reflect.ValueOf(zapcore.RFC3339TimeEncoder)

			Expect(optVal.Pointer()).To(Equal(expVal.Pointer()))
		})

		It("Should default to 'iso8061' time encoding", func() {
			GinkgoT().Setenv("ZAP_TIME_ENCODING", "")

			opt := crzap.Options{}
			UseFlagOptions(&opts)(&opt)

			optVal := reflect.ValueOf(opt.TimeEncoder)
			expVal := reflect.ValueOf(zapcore.EpochTimeEncoder)

			Expect(optVal.Pointer()).To(Equal(expVal.Pointer()))
		})

		It("Should return epochTimeEncoder for unknown time-encoding", func() {
			GinkgoT().Setenv("ZAP_TIME_ENCODING", "unknown")

			opt := crzap.Options{}
			UseFlagOptions(&opts)(&opt)

			optVal := reflect.ValueOf(opt.TimeEncoder)
			expVal := reflect.ValueOf(zapcore.EpochTimeEncoder)

			Expect(optVal.Pointer()).To(Equal(expVal.Pointer()))
		})

		It("Should propagate time encoder to logger", func() {
			// zaps ISO8601TimeEncoder uses 2006-01-02T15:04:05.000Z0700 as pattern for iso8601 encoding
			iso8601Pattern := `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{3}([-+][0-9]{4}|Z)`

			GinkgoT().Setenv("ZAP_TIME_ENCODING", "iso8601")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.Bytes()

			res := map[string]any{}
			Expect(json.Unmarshal(outRaw, &res)).To(Succeed())
			Expect(res["ts"]).Should(MatchRegexp(iso8601Pattern))
		})
	})

	Context("with zap-encoding flag provided", Label("encoder"), func() {
		It("Should default to console encoder when not set (in development mode)", func() {
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.String()
			expectedPattern := `.+\tINFO\tThis is a test message\n`
			Expect(outRaw).Should(MatchRegexp(expectedPattern))
		})

		It("Should set json encoder in options", func() {
			GinkgoT().Setenv("ZAP_ENCODER", "json")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.Bytes()

			Expect(json.Valid(outRaw)).To(BeTrue())
		})

		It("Should default to json encoder when not set", func() {
			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.Bytes()

			Expect(json.Valid(outRaw)).To(BeTrue())
		})

		It("should PANIC when invalid encoder is supplied", func() {
			GinkgoT().Setenv("ZAP_ENCODER", "invalid")

			panicFunc := func() {
				UseFlagOptions(&opts)
			}

			Expect(panicFunc).To(Panic())
		})

		It("Should propagate console encoder to logger", func() {
			GinkgoT().Setenv("ZAP_ENCODER", "console")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.String()
			expectedPattern := `.+\tINFO\tThis is a test message\n`
			Expect(outRaw).Should(MatchRegexp(expectedPattern))
		})

		It("Should propagate json encoder to logger", func() {
			GinkgoT().Setenv("ZAP_ENCODER", "json")
			GinkgoT().Setenv("ZAP_TIME_ENCODING", "iso8601")

			logOut := new(bytes.Buffer)

			logger := New(UseFlagOptions(&opts), crzap.WriteTo(logOut))
			logger.Info(testMessage)

			outRaw := logOut.String()
			expectedPattern := `{\"level\":\"info\",\"ts\":\".+\",\"msg\":\"This is a test message\"}\n`
			Expect(outRaw).Should(MatchRegexp(expectedPattern))
		})
	})

	Context("with  zap-stacktrace-level options provided", func() {
		It("Should output stacktrace at info level, with development mode set to true.", func() {
			GinkgoT().Setenv("ZAP_STACKTRACE_LEVEL", "info")
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			out := crzap.Options{}
			UseFlagOptions(&opts)(&out)

			Expect(out.StacktraceLevel.Enabled(zapcore.InfoLevel)).To(BeTrue())
		})

		It("Should output stacktrace at error level, with development mode set to true.", func() {
			GinkgoT().Setenv("ZAP_STACKTRACE_LEVEL", "error")
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			out := crzap.Options{}
			UseFlagOptions(&opts)(&out)

			Expect(out.StacktraceLevel.Enabled(zapcore.ErrorLevel)).To(BeTrue())
		})

		It("Should output stacktrace at panic level, with development mode set to true.", func() {
			GinkgoT().Setenv("ZAP_STACKTRACE_LEVEL", "panic")
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			out := crzap.Options{}
			UseFlagOptions(&opts)(&out)

			Expect(out.StacktraceLevel.Enabled(zapcore.PanicLevel)).To(BeTrue())
			Expect(out.StacktraceLevel.Enabled(zapcore.ErrorLevel)).To(BeFalse())
			Expect(out.StacktraceLevel.Enabled(zapcore.InfoLevel)).To(BeFalse())
		})
	})

	Context("with only -zap-devel flag provided", func() {
		It("Should set dev=true.", Label("onlydev"), func() {
			GinkgoT().Setenv("ZAP_DEVEL", "true")

			out := crzap.Options{}

			UseFlagOptions(&opts)(&out)

			Expect(out.Development).To(BeTrue())
			Expect(out.Encoder).To(BeNil())
			Expect(out.Level).To(BeNil())
			Expect(out.StacktraceLevel).To(BeNil())
			Expect(out.EncoderConfigOptions).To(BeNil())
		})

		It("Should set dev=false", func() {
			GinkgoT().Setenv("ZAP_DEVEL", "false")

			out := crzap.Options{}

			UseFlagOptions(&opts)(&out)

			Expect(out.Development).To(BeFalse())
			Expect(out.Encoder).To(BeNil())
			Expect(out.Level).To(BeNil())
			Expect(out.StacktraceLevel).To(BeNil())
			Expect(out.EncoderConfigOptions).To(BeNil())
		})
	})
})
