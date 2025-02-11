package cash

import (
	"AvitoTest/pkg/models/apimodels"
	"testing"

	"github.com/stretchr/testify/require"
)

type DB struct{}

func (d *DB) Buy(user, item string) error {
	return nil
}
func (d *DB) GetUserBalance(username string) (int, error) {
	return 1000, nil
}
func (d *DB) GetUserInventory(username string) ([]apimodels.Item, error) {
	return []apimodels.Item{}, nil
}
func (d *DB) GetUserRecieveHistory(username string) ([]apimodels.Recieving, error) {
	return []apimodels.Recieving{}, nil
}
func (d *DB) GetUserSendHistory(username string) ([]apimodels.Sending, error) {
	return []apimodels.Sending{}, nil
}
func (d *DB) SendCoins(sender, reciever string, amount int) error {
	return nil
}

func TestCash(t *testing.T) {
	cash, err := New(&DB{})
	require.NoError(t, err)

	username1 := "john1"
	username2 := "john2"
	item := "nice"
	amount := 2

	err = cash.BuyItem(username1, item)
	if err != nil {
		t.Error(err)
	}

	info, err := cash.UserInfo(username1)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, 1000, info.Coins)

	err = cash.SendCoins(username1, apimodels.SendCoinRequest{ToUser: username2, Amount: amount})
	if err != nil {
		t.Error(err)
	}
}
