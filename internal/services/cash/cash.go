package cash

import (
	cerr "AvitoTest/pkg/customErrors"
	"AvitoTest/pkg/models/apimodels"
	"context"
	"time"
)

type Cash struct {
	db IDB
}

type IDB interface {
	Buy(context.Context, string, string) error
	GetUserBalance(context.Context, string) (int, error)
	GetUserInventory(context.Context, string) ([]apimodels.Item, error)
	GetUserRecieveHistory(context.Context, string) ([]apimodels.Recieving, error)
	GetUserSendHistory(context.Context, string) ([]apimodels.Sending, error)
	SendCoins(context.Context, string, string, int) error
}

func New(db IDB) (*Cash, error) {
	return &Cash{db: db}, nil
}

func (c *Cash) BuyItem(username, item string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.db.Buy(ctx, username, item)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cash) UserInfo(username string) (apimodels.InfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	balance, err := c.db.GetUserBalance(ctx, username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	inventory, err := c.db.GetUserInventory(ctx, username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	rhistory, err := c.db.GetUserRecieveHistory(ctx, username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	shistory, err := c.db.GetUserSendHistory(ctx, username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	return apimodels.InfoResponse{Coins: balance, Inventory: inventory, CoinHistory: apimodels.CoinHistory{Recieved: rhistory, Sent: shistory}}, nil
}

func (c *Cash) SendCoins(username string, request apimodels.SendCoinRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	if username == request.ToUser {
		return cerr.ErrSelfSend
	}
	err := c.db.SendCoins(ctx, username, request.ToUser, request.Amount)
	if err != nil {
		return err
	}
	return nil
}
