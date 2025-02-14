package auth

import (
	"AvitoTest/pkg/models/apimodels"
	"context"
	"time"
)

type Auth struct {
	db  IDB
	jwt IJWT
}

type IDB interface {
	AuthorizeUser(ctx context.Context, username, password string) error
}

type IJWT interface {
	CreateToken(username string) (string, error)
}

func New(db IDB, jwt IJWT) (*Auth, error) {
	return &Auth{db: db, jwt: jwt}, nil
}

func (a *Auth) AuthorizeUser(credentials apimodels.AuthRequest) (apimodels.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := a.db.AuthorizeUser(ctx, credentials.Username, credentials.Password)
	if err != nil {
		return apimodels.AuthResponse{}, err
	}
	token, err := a.jwt.CreateToken(credentials.Username)
	if err != nil {
		return apimodels.AuthResponse{}, err
	}
	return apimodels.AuthResponse{Token: token}, nil
}
