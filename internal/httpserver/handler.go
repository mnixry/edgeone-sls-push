package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mnixry/edgeone-sls-push/internal/edgeone"
	"github.com/rs/zerolog"
)

type LogForwarder interface {
	Enqueue(timestamp uint32, record map[string]string) error
}

type Response struct {
	ResultCode string `json:"result_code"`
	ResultDesc string `json:"result_desc"`
	Timestamp  int64  `json:"timestamp"`
}

type Handler struct {
	auth *edgeone.AuthVerifier
	fwd  LogForwarder
	log  zerolog.Logger
}

func NewHandler(
	auth *edgeone.AuthVerifier,
	fwd LogForwarder,
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
		ResultCode: "-1",
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

	body := c.Body()

	rec, err := edgeone.ParseRecord(body)
	if err != nil {
		h.log.Warn().Err(err).Msg("parse failed")
		return errResponse(c, fiber.StatusBadRequest, "malformed JSON body")
	}
	h.log.Debug().Str("ip", c.IP()).Interface("record", rec).Msg("record accepted")

	timestamp := uint32(time.Now().Unix())
	if parsed, err := time.Parse(time.RFC3339, rec.LogTime); err == nil {
		timestamp = uint32(parsed.Unix())
	}
	if err := h.fwd.Enqueue(timestamp, rec.Normalize()); err != nil {
		h.log.Error().Err(err).Msg("enqueue to SLS failed")
		return errResponse(c, fiber.StatusInternalServerError, "internal error")
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		ResultCode: "1",
		ResultDesc: "Success",
		Timestamp:  time.Now().Unix(),
	})
}
