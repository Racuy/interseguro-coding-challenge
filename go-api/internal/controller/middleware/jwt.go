// auth middleware
package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"go-api/internal/controller/bearer"
)

// rejects requests without a valid HS256 bearer token
func JWTProtected(secret string) fiber.Handler {
	key := []byte(secret)

	return func(c *fiber.Ctx) error {
		tokenString, ok := bearer.Parse(c.Get(fiber.HeaderAuthorization))
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}

		// RegisteredClaims checks exp/iat/nbf automatically
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
			// blocks alg:none and algorithm-confusion attacks
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return key, nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		return c.Next()
	}
}
