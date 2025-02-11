package db

import (
	cerr "AvitoTest/pkg/customErrors"
	"AvitoTest/pkg/models/apimodels"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type User struct {
	username string
	password string
}

func TestDB(t *testing.T) {
	cfg := Config{
		Host:     "0.0.0.0",
		Port:     "5436",
		User:     "user",
		Password: "123",
		DBName:   "avitomockdb",
	}
	db, err := New(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	db.db.Exec(context.Background(), `delete from public.users`)
	db.db.Exec(context.Background(), `delete from public.inventory`)
	db.db.Exec(context.Background(), `delete from public.history`)
	t.Log("Testing authentication")
	firstUser := User{username: "abc", password: "123"}
	firstUserWrongPassword := User{username: "abc", password: "111"}
	secondUser := User{username: "cool_guy_1337", password: "123123"}

	err = db.AuthorizeUser(firstUser.username, firstUser.password)
	if err != nil {
		t.Error(err)
	}
	err = db.AuthorizeUser(secondUser.username, secondUser.password)
	if err != nil {
		t.Error(err)
	}

	err = db.AuthorizeUser(firstUserWrongPassword.username, firstUserWrongPassword.password)
	if err != nil {
		require.ErrorIs(t, err, cerr.ErrWrongPassword, "Error while authorizing with wrong password should be %v, not %v", cerr.ErrWrongPassword, err)
	}

	t.Log("testing balance machinations")
	balance, err := db.GetUserBalance(firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, balance, 1000, "New user's balance should be equal 1000, not %v", balance)

	err = db.Buy(firstUser.username, "cup")
	if err != nil {
		t.Error(err)
	}

	inv, err := db.GetUserInventory(firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Item{{Type: "cup", Quantity: 1}}, inv, "Inventory should be %v, not %v", []apimodels.Item{{Type: "cup", Quantity: 1}}, inv)

	balance, err = db.GetUserBalance(firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, balance, 980, "New user's balance should be equal 980, not %v", balance)

	err = db.Buy(firstUser.username, "non existent item")
	if err != nil {
		require.ErrorIs(t, err, cerr.ErrItemNotExist, "Error should be %v, not %v", cerr.ErrItemNotExist, err)
	}

	err = db.SendCoins(firstUser.username, secondUser.username, 10)
	if err != nil {
		t.Error(err)
	}

	err = db.SendCoins(secondUser.username, firstUser.username, 20)
	if err != nil {
		t.Error(err)
	}

	rhistory, err := db.GetUserRecieveHistory(firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Recieving{{FromUser: secondUser.username, Amount: 20}}, rhistory, "History should be %v, not %v", []apimodels.Recieving{{FromUser: secondUser.username, Amount: 20}}, rhistory)

	shistory, err := db.GetUserSendHistory(firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Sending{{ToUser: secondUser.username, Amount: 10}}, shistory, "History should be %v, not %v", []apimodels.Sending{{ToUser: secondUser.username, Amount: 10}}, shistory)

}
