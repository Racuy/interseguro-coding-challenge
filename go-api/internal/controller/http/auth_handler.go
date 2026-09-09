package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-api/internal/application"
)

type AuthHandler struct {
	usecase application.AuthUsecase
}

func NewAuthHandler(usecase application.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// POST /api/v1/auth/login: {username,password} -> {token}
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	token, err := h.usecase.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to log in",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
}
