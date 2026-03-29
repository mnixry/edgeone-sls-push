package sls

import (
	"fmt"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/rs/zerolog"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Project         string
	LogStore        string
	Topic           string
	Source          string
	LingerMs        int64
	MaxBatchSize    int64
	MaxBatchCount   int
	Retries         int
	BaseRetryMs     int64
	MaxRetryMs      int64
}

type Forwarder struct {
	producer *producer.Producer
	client   sls.ClientInterface
	project  string
	logStore string
	topic    string
	source   string
	log      zerolog.Logger
}

type logCallback struct {
	log zerolog.Logger
}

func (c *logCallback) Success(result *producer.Result) {
	c.log.Debug().
		Str("request_id", result.GetRequestId()).
		Msg("SLS batch sent")
}

func (c *logCallback) Fail(result *producer.Result) {
	c.log.Error().
		Str("error_code", result.GetErrorCode()).
		Str("error_message", result.GetErrorMessage()).
		Str("request_id", result.GetRequestId()).
		Int("attempts", len(result.GetReservedAttempts())).
		Msg("SLS batch failed after retries")
}

func NewForwarder(cfg Config, log zerolog.Logger) (*Forwarder, error) {
	pc := producer.GetDefaultProducerConfig()
	pc.Endpoint = cfg.Endpoint
	pc.AccessKeyID = cfg.AccessKeyID
	pc.AccessKeySecret = cfg.AccessKeySecret
	pc.LingerMs = cfg.LingerMs
	if cfg.MaxBatchSize > 0 {
		pc.MaxBatchSize = cfg.MaxBatchSize
	}
	if cfg.MaxBatchCount > 0 {
		pc.MaxBatchCount = cfg.MaxBatchCount
	}
	if cfg.Retries > 0 {
		pc.Retries = cfg.Retries
	}
	pc.BaseRetryBackoffMs = cfg.BaseRetryMs
	pc.MaxRetryBackoffMs = cfg.MaxRetryMs

	p, err := producer.NewProducer(pc)
	if err != nil {
		return nil, fmt.Errorf("create SLS producer: %w", err)
	}
	p.Start()

	provider := sls.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, "")
	client := sls.CreateNormalInterfaceV2(cfg.Endpoint, provider)

	return &Forwarder{
		producer: p,
		client:   client,
		project:  cfg.Project,
		logStore: cfg.LogStore,
		topic:    cfg.Topic,
		source:   cfg.Source,
		log:      log.With().Str("component", "sls-forwarder").Logger(),
	}, nil
}

func (f *Forwarder) Enqueue(timestamp uint32, record map[string]string) error {
	cb := &logCallback{log: f.log}
	entry := producer.GenerateLog(timestamp, record)
	return f.producer.SendLogWithCallBack(
		f.project, f.logStore, f.topic, f.source, entry, cb,
	)
}

// Healthy verifies connectivity to SLS by fetching the logstore metadata.
func (f *Forwarder) Healthy() error {
	_, err := f.client.GetLogStore(f.project, f.logStore)
	return err
}

func (f *Forwarder) Close() {
	f.log.Info().Msg("draining SLS producer")
	f.producer.SafeClose()
}
