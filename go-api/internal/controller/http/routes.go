package http

import "github.com/gofiber/fiber/v2"

// wires up the matrix routes
func RegisterRoutes(router fiber.Router, handler *MatrixHandler) {
	matrix := router.Group("/matrix")
	matrix.Post("/process", handler.Process)
}

// wires up the login route, public, no JWT required
func RegisterAuthRoutes(router fiber.Router, handler *AuthHandler) {
	auth := router.Group("/auth")
	auth.Post("/login", handler.Login)
}
