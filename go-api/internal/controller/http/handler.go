// turns HTTP requests into use case calls, using Fiber
package http

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-api/internal/application"
	"go-api/internal/controller/bearer"
	"go-api/internal/domain"
)

// caps the whole rotate+QR+node-api pipeline
const matrixProcessTimeout = 5 * time.Second

type MatrixHandler struct {
	usecase application.MatrixUsecase
}

func NewMatrixHandler(usecase application.MatrixUsecase) *MatrixHandler {
	return &MatrixHandler{usecase: usecase}
}

// POST /api/v1/matrix/process: {matrix:[[...]]} -> {max,min,average,sum,isDiagonal}
func (h *MatrixHandler) Process(c *fiber.Ctx) error {
	var req domain.MatrixRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// attach token and a deadline so a slow node-api can't hang the request
	token, _ := bearer.Parse(c.Get(fiber.HeaderAuthorization))
	ctx, cancel := context.WithTimeout(application.ContextWithToken(c.Context(), token), matrixProcessTimeout)
	defer cancel()

	result, err := h.usecase.Process(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidMatrix) || errors.Is(err, domain.ErrMatrixTooLarge) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process matrix",
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
