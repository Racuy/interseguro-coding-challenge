// gentoken: prints a JWT signed with JWT_SECRET, for testing go-api locally
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-api/internal/controller/auth"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	issuer := auth.NewJWTIssuer(secret)
	token, err := issuer.IssueToken(context.Background(), "local-test-user")
	if err != nil {
		log.Fatalf("sign token: %v", err)
	}

	fmt.Println(token)
}
