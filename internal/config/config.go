package config

import (
	"time"

	"github.com/rs/zerolog"
)

type CLI struct {
	HTTP    HTTPConfig    `embed:"" prefix:"http-" envprefix:"HTTP_"`
	EdgeOne EdgeOneConfig `embed:"" prefix:"edgeone-" envprefix:"EDGEONE_"`
	SLS     SLSConfig     `embed:"" prefix:"sls-" envprefix:"SLS_"`
	Log     LogConfig     `embed:"" prefix:"log-" envprefix:"LOG_"`
}

type HTTPConfig struct {
	Addr         string        `name:"addr" env:"ADDR" default:":8080" help:"HTTP server listen address."`
	Path         string        `name:"path" env:"PATH" default:"/edgeone-logs" help:"URL path to receive log pushes."`
	ReadTimeout  time.Duration `name:"read-timeout" env:"READ_TIMEOUT" default:"30s" help:"HTTP read timeout."`
	WriteTimeout time.Duration `name:"write-timeout" env:"WRITE_TIMEOUT" default:"30s" help:"HTTP write timeout."`
	MaxBodyBytes int           `name:"max-body-bytes" env:"MAX_BODY_BYTES" default:"10485760" help:"Max request body size in bytes."`
}

type EdgeOneConfig struct {
	SecretID  string        `name:"secret-id" env:"SECRET_ID" required:"" help:"EdgeOne SecretId for signature verification."`
	SecretKey string        `name:"secret-key" env:"SECRET_KEY" required:"" help:"EdgeOne SecretKey for signature verification."`
	MaxSkew   time.Duration `name:"max-skew" env:"MAX_SKEW" default:"300s" help:"Max allowed clock skew for auth_key timestamp."`
}

type SLSConfig struct {
	Endpoint        string `name:"endpoint" env:"ENDPOINT" required:"" help:"SLS endpoint (e.g. cn-hangzhou.log.aliyuncs.com)."`
	AccessKeyID     string `name:"access-key-id" env:"ACCESS_KEY_ID" required:"" help:"Alibaba Cloud AccessKey ID."`
	AccessKeySecret string `name:"access-key-secret" env:"ACCESS_KEY_SECRET" required:"" help:"Alibaba Cloud AccessKey Secret."`
	Project         string `name:"project" env:"PROJECT" required:"" help:"SLS project name."`
	LogStore        string `name:"logstore" env:"LOGSTORE" required:"" help:"SLS logstore name."`
	Topic           string `name:"topic" env:"TOPIC" default:"" help:"SLS log topic."`
	Source          string `name:"source" env:"SOURCE" default:"" help:"SLS log source."`
	LingerMs        int64  `name:"linger-ms" env:"LINGER_MS" default:"2000" help:"Producer linger time in ms before flushing a batch."`
	MaxBatchSize    int64  `name:"max-batch-size" env:"MAX_BATCH_SIZE" default:"524288" help:"Max batch size in bytes."`
	MaxBatchCount   int    `name:"max-batch-count" env:"MAX_BATCH_COUNT" default:"4096" help:"Max number of logs per batch."`
	Retries         int    `name:"retries" env:"RETRIES" default:"10" help:"Max retry attempts for failed batches."`
	BaseRetryMs     int64  `name:"base-retry-ms" env:"BASE_RETRY_MS" default:"100" help:"Base retry backoff in ms."`
	MaxRetryMs      int64  `name:"max-retry-ms" env:"MAX_RETRY_MS" default:"50000" help:"Max retry backoff in ms."`
}

type LogFormat string

const (
	LogFormatJSON    LogFormat = "json"
	LogFormatConsole LogFormat = "console"
)

type LogConfig struct {
	Level      zerolog.Level `name:"level" env:"LEVEL" default:"info" help:"Log level (trace, debug, info, warn, error, fatal, panic)."`
	Output     string        `name:"output" env:"OUTPUT" default:"stdout" help:"Log output: stdout, stderr, or a file path."`
	Format     LogFormat     `name:"format" env:"FORMAT" default:"json" enum:"json,console" help:"Log format: json or console."`
	MaxSize    int           `name:"max-size" env:"MAX_SIZE" default:"100" help:"Max size in MB before log rotation (0 disables)."`
	MaxAge     int           `name:"max-age" env:"MAX_AGE" default:"30" help:"Max age in days to retain old log files."`
	MaxBackups int           `name:"max-backups" env:"MAX_BACKUPS" default:"10" help:"Max number of old log files to retain."`
	Compress   bool          `name:"compress" env:"COMPRESS" default:"true" help:"Compress rotated log files with gzip."`
}
