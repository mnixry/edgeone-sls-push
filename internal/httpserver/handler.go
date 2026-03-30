package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mnixry/edgeone-sls-push/internal/edgeone"
	"github.com/mnixry/edgeone-sls-push/internal/sls"
	"github.com/rs/zerolog"
)

type ResultCode string

const (
	ResultCodeSuccess ResultCode = "1"
	ResultCodeError   ResultCode = "-1"
)

type Response struct {
	ResultCode ResultCode `json:"result_code"`
	ResultDesc string     `json:"result_desc"`
	Timestamp  int64      `json:"timestamp"`
}

type Handler struct {
	auth *edgeone.AuthVerifier
	fwd  *sls.Forwarder
	log  zerolog.Logger
}

func NewHandler(
	auth *edgeone.AuthVerifier,
	fwd *sls.Forwarder,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		auth: auth,
		fwd:  fwd,
		log:  log.With().Str("component", "handler").Logger(),
	}
}

func errResponse(c fiber.Ctx, status int, desc string) error {
	return c.Status(status).JSON(Response{
		ResultCode: ResultCodeError,
		ResultDesc: desc,
		Timestamp:  time.Now().Unix(),
	})
}

func (h *Handler) Handle(c fiber.Ctx) error {
	path := c.Path()
	authKey := fiber.Query[string](c, "auth_key")
	accessKey := fiber.Query[string](c, "access_key")

	if err := h.auth.Verify(path, authKey, accessKey); err != nil {
		h.log.Warn().Err(err).Str("remote", c.IP()).Msg("auth failed")
		return errResponse(c, fiber.StatusUnauthorized, "unauthorized")
	}

	records, err := edgeone.ParseRecords(c.Body())
	if err != nil {
		h.log.Warn().Err(err).Msg("parse failed")
		return errResponse(c, fiber.StatusBadRequest, "malformed JSON body")
	}

	h.log.Debug().
		Str("remote", c.IP()).
		Interface("records", records).
		Msg("forwarding batch")

	entries := make([]sls.LogEntry, 0, len(records))
	for _, rec := range records {
		ts := uint32(time.Now().Unix())
		if parsed, parseErr := time.Parse(time.RFC3339, rec.LogTime); parseErr == nil {
			ts = uint32(parsed.Unix())
		}
		entries = append(entries, sls.LogEntry{Timestamp: ts, Fields: rec.Normalize()})
	}

	if err := h.fwd.Enqueue(entries); err != nil {
		h.log.Error().Err(err).Msg("enqueue to SLS failed")
		return errResponse(c, fiber.StatusInternalServerError, "internal error")
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		ResultCode: ResultCodeSuccess,
		ResultDesc: "Success",
		Timestamp:  time.Now().Unix(),
	})
}
