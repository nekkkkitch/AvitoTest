package jwt

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	cfg := Config{
		AccessTokenExpiration:  3600,
		RefreshTokenExpiration: 3600,
	}
	jwt, err := New(cfg)
	if err != nil {
		t.Error(err)
	}

	username := "abc"
	token, err := jwt.CreateToken(username)
	if err != nil {
		t.Error(err)
	}

	ok, err := jwt.ValidateToken(&fiber.Ctx{}, token)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, true, ok, "Validation should be true, not %v", ok)

	usernameFromToken, err := jwt.GetUsernameFromToken(token)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, username, usernameFromToken, "Username from token should be %v, not %v", username, usernameFromToken)

}
