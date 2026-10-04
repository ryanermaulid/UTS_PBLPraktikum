package helper

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

const (
	statusSuccess = "success"
	statusError   = "error"
)

// Meta dipakai untuk response berpagination.
type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

// successEnvelope adalah body response sukses. request_id sengaja tidak
// dimasukkan di sini karena SPEC.md hanya meminta request_id pada error.
type successEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// errorEnvelope adalah body response error. request_id selalu disertakan.
type errorEnvelope struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Errors    map[string][]string `json:"errors,omitempty"`
	RequestID string              `json:"request_id"`
}

func OK(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(successEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(successEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Paginated(c *fiber.Ctx, message string, data any, meta Meta) error {
	return c.Status(fiber.StatusOK).JSON(successEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Error menulis response error dengan format seragam. AppError.Fields akan
// dimasukan ke field "errors". Untuk error lain, status dibatasi sesuai
// SPEC.md (hanya 401, 403, 404, 409, 422, 429, 500) oleh pemanggil.
func Error(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "Terjadi kesalahan pada server"
	var fields map[string][]string

	if ae, ok := AsAppError(err); ok {
		status = ae.Status
		message = ae.Message
		fields = ae.Fields
	} else if fe, ok := err.(*fiber.Error); ok {
		status = fe.Code
		message = fe.Message
	}

	return c.Status(status).JSON(errorEnvelope{
		Success:   false,
		Message:   message,
		Errors:    fields,
		RequestID: requestIDFrom(c),
	})
}

func requestIDFrom(c *fiber.Ctx) string {
	if v := c.Locals(requestid.ConfigDefault.ContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
