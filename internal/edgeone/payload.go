package edgeone

import (
	"bytes"
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

// ParseRecords reads one or more JSON objects from data (single JSON or JSON Lines)
// using Go's streaming decoder, so both formats work without special handling.
func ParseRecords(data []byte) ([]*LogRecord, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var records []*LogRecord
	for dec.More() {
		var fields map[string]any
		if err := dec.Decode(&fields); err != nil {
			return nil, fmt.Errorf("json decode: %w", err)
		}
		rec := &LogRecord{Fields: fields}
		if err := tryCastAndPopulate(fields, "RequestID", &rec.RequestID); err != nil {
			return nil, err
		}
		if err := tryCastAndPopulate(fields, "LogTime", &rec.LogTime); err != nil {
			return nil, err
		}
		if err := tryCastAndPopulate(fields, "ContentID", &rec.ContentID); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	return records, nil
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
			if b, err := json.Marshal(v); err == nil {
				normalized[k] = string(b)
			}
		}
	}
	return normalized
}
