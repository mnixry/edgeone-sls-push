package edgeone

import (
	"crypto/md5"
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AuthConfig struct {
	SecretID  string
	SecretKey string
	MaxSkew   time.Duration
}

type AuthVerifier struct {
	cfg AuthConfig
}

func NewAuthVerifier(cfg AuthConfig) *AuthVerifier {
	return &AuthVerifier{cfg: cfg}
}

// Verify checks the EdgeOne auth_key signature.
//
// Expected query parameters (extracted by the caller):
//
//	auth_key = timestamp-rand-md5hash
//	access_key = SecretId
//
// where md5hash = md5("path-timestamp-rand-SecretKey").
func (v *AuthVerifier) Verify(path, authKey, accessKey string) error {
	if authKey == "" || accessKey == "" {
		return fmt.Errorf("missing auth_key or access_key query parameter")
	}

	if accessKey != v.cfg.SecretID {
		return fmt.Errorf("unknown access_key")
	}

	parts := strings.SplitN(authKey, "-", 3)
	if len(parts) != 3 {
		return fmt.Errorf("malformed auth_key: expected timestamp-rand-md5hash")
	}

	timestamp, rand, md5hash := parts[0], parts[1], parts[2]

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp in auth_key: %w", err)
	}

	skew := time.Since(time.Unix(ts, 0)).Abs()
	if skew > v.cfg.MaxSkew {
		return fmt.Errorf("timestamp skew %s exceeds max %s", skew.Truncate(time.Second), v.cfg.MaxSkew)
	}

	signStr := fmt.Sprintf("%s-%s-%s-%s", path, timestamp, rand, v.cfg.SecretKey)
	expected := fmt.Sprintf("%x", md5.Sum([]byte(signStr)))

	if subtle.ConstantTimeCompare([]byte(expected), []byte(md5hash)) != 1 {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
