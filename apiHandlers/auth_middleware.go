package apiHandlers

import (
	"context"
	"lawyerSL-Backend/dto"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

type AuthMiddleware struct {
    config      *dto.AuthConfig
    cache       *cache.Cache
    firebaseApp *firebase.App
}

func NewAuthMiddleware(config dto.AuthConfig, firebaseApp *firebase.App) *AuthMiddleware {
    return &AuthMiddleware{
        config:      &config,
        cache:       cache.New(5*time.Minute, 10*time.Minute),
        firebaseApp: firebaseApp,
    }
}


func (a *AuthMiddleware) ValidateToken(c *fiber.Ctx) error {
    ctx := context.Background()

    client, err := a.firebaseApp.Auth(ctx)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to init Firebase"})
    }

    authHeader := c.Get("Authorization")
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing/invalid auth header"})
    }

    idToken := parts[1]
    token, err := client.VerifyIDToken(ctx, idToken)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token"})
    }

    c.Locals("uid", token.UID)
    if email, ok := token.Claims["email"].(string); ok {
        c.Locals("email", email)
    }

    return c.Next()
}
