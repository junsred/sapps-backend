package middleware

import (
	"fmt"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

type Status struct {
	Code    int    `json:"-"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewStatus(code int, status string) Status {
	return Status{Code: code, Status: status}
}

var (
	StatusBadRequest          = NewStatus(fiber.StatusBadRequest, "BAD_REQUEST")
	StatusUnauthorized        = NewStatus(fiber.StatusUnauthorized, "UNAUTHORIZED")
	StatusForbidden           = NewStatus(fiber.StatusForbidden, "FORBIDDEN")
	StatusNotFound            = NewStatus(fiber.StatusNotFound, "NOT_FOUND")
	StatusConflict            = NewStatus(fiber.StatusConflict, "CONFLICT")
	StatusInternalServerError = NewStatus(fiber.StatusInternalServerError, "INTERNAL_ERROR")
)

type ErrorStatus struct {
	Error Status `json:"error"`
}

func (c *RequestContext) Error(status Status, message ...string) error {
	s := ErrorStatus{Error: status}
	if len(message) > 0 {
		s.Error.Message = message[0]
	}
	return c.Status(status.Code).JSON(s)
}

func (c *RequestContext) ErrorInternal(message ...string) error {
	return c.Error(StatusInternalServerError, message...)
}

func (c *RequestContext) LogErr(err error, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	critical := false
	for _, arg := range args {
		if arg == "critical" {
			critical = true
		}
	}
	criticalStr := "Critical Error"
	if !critical {
		criticalStr = "Error"
	}
	errStr := fmt.Sprintf("%v\n%s occurred at %s:%d\n", err, criticalStr, file, line)
	log.Print(errStr)
	/*if critical {
		gclog.Critical(errStr)
	} else {
		gclog.Error(errStr)
	}*/
}
