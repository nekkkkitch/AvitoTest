package auth

import (
	"AvitoTest/pkg/models/apimodels"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type DB struct{}
type JWT struct{}

func (d *DB) AuthorizeUser(_ context.Context, username, password string) error {
	return nil
}

func (j *JWT) CreateToken(username string) (string, error) {
	return "token", nil
}

func TestAuth(t *testing.T) {
	var err error
	auth, err := New(&DB{}, &JWT{})
	if err != nil {
		t.Error(err)
	}

	credentials := apimodels.AuthRequest{Username: "abc", Password: "123"}
	token, err := auth.AuthorizeUser(credentials)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, "token", token.Token, "Token should be token, not "+token.Token)
}
