package edgeone

import (
	"encoding/json"
	"fmt"
)

// LogRecord represents a single EdgeOne access-log event.
// Required fields: RequestID, LogTime, ContentID.
// All other fields from https://cloud.tencent.com/document/product/1552/105791
// are preserved in Fields for forwarding to SLS.
type LogRecord struct {
	RequestID string
	LogTime   string
	ContentID string
	Fields    map[string]any
}

func tryCastAndPopulate[T any](m map[string]any, key string, t *T) (err error) {
	if v, ok := m[key]; ok {
		*t, ok = v.(T)
		if !ok {
			return fmt.Errorf("key %s is not of type %T", key, t)
		}
		delete(m, key)
	}
	return nil
}

func ParseRecord(data []byte) (*LogRecord, error) {
	record := &LogRecord{}
	if err := json.Unmarshal(data, &record.Fields); err != nil {
		return nil, err
	}
	if err := tryCastAndPopulate(record.Fields, "RequestID", &record.RequestID); err != nil {
		return nil, err
	}
	if err := tryCastAndPopulate(record.Fields, "LogTime", &record.LogTime); err != nil {
		return nil, err
	}
	if err := tryCastAndPopulate(record.Fields, "ContentID", &record.ContentID); err != nil {
		return nil, err
	}
	return record, nil
}

func (r *LogRecord) Normalize() map[string]string {
	normalized := map[string]string{
		"RequestID": r.RequestID,
		"LogTime":   r.LogTime,
		"ContentID": r.ContentID,
	}
	for k, v := range r.Fields {
		switch v := v.(type) {
		case string:
			normalized[k] = v
		default:
			if json, err := json.Marshal(v); err == nil {
				normalized[k] = string(json)
			}
		}
	}
	return normalized
}
