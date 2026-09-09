package http

import "github.com/gofiber/fiber/v2"

// wires up the matrix routes
func RegisterRoutes(router fiber.Router, handler *MatrixHandler) {
	matrix := router.Group("/matrix")
	matrix.Post("/process", handler.Process)
}
