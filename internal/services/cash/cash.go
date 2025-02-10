package cash

import (
	"AvitoTest/pkg/models/apimodels"
)

type Cash struct {
	db IDB
}

type IDB interface {
	Buy(string, string) error
	GetUserBalance(string) (int, error)
	GetUserInventory(string) ([]apimodels.Item, error)
	GetUserRecieveHistory(string) ([]apimodels.Recieving, error)
	GetUserSendHistory(string) ([]apimodels.Sending, error)
	SendCoins(string, string, int) error
}

func New(db IDB) (*Cash, error) {
	return &Cash{db: db}, nil
}

func (c *Cash) BuyItem(username, item string) error {
	err := c.db.Buy(username, item)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cash) UserInfo(username string) (apimodels.InfoResponse, error) {
	balance, err := c.db.GetUserBalance(username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	inventory, err := c.db.GetUserInventory(username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	rhistory, err := c.db.GetUserRecieveHistory(username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	shistory, err := c.db.GetUserSendHistory(username)
	if err != nil {
		return apimodels.InfoResponse{}, err
	}
	return apimodels.InfoResponse{Coins: balance, Inventory: inventory, CoinHistory: apimodels.CoinHistory{Recieved: rhistory, Sent: shistory}}, nil
}

func (c *Cash) SendCoins(username string, request apimodels.SendCoinRequest) error {
	err := c.db.SendCoins(username, request.ToUser, request.Amount)
	if err != nil {
		return err
	}
	return nil
}
