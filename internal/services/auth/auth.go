package auth

import (
	"AvitoTest/pkg/models/apimodels"
)

type Auth struct {
	db  IDB
	jwt IJWT
}

type IDB interface {
	AuthorizeUser(username, password string) error
}

type IJWT interface {
	CreateToken(username string) (string, error)
}

func New(db IDB, jwt IJWT) (*Auth, error) {
	return &Auth{db: db, jwt: jwt}, nil
}

func (a *Auth) AuthorizeUser(credentials apimodels.AuthRequest) (apimodels.AuthResponse, error) {
	err := a.db.AuthorizeUser(credentials.Username, credentials.Password)
	if err != nil {
		return apimodels.AuthResponse{}, err
	}
	token, err := a.jwt.CreateToken(credentials.Username)
	if err != nil {
		return apimodels.AuthResponse{}, err
	}
	return apimodels.AuthResponse{Token: token}, nil
}
