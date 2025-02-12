package db

import (
	cerr "AvitoTest/pkg/customErrors"
	"AvitoTest/pkg/models/apimodels"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type User struct {
	username string
	password string
}

var (
	db         *DB
	firstUser  User
	secondUser User
)

func TestDB(t *testing.T) {
	t.Log("Starting")
	cfg := Config{
		Host:     "0.0.0.0",
		Port:     "5434",
		User:     "user",
		Password: "123",
		DBName:   "avitodb",
	}
	var err error
	db, err = New(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Success.")
}

func TestAuth(t *testing.T) {
	t.Log("Testing authentication")
	firstUser = User{username: "abc", password: "123"}
	firstUserWrongPassword := User{username: "abc", password: "111"}
	secondUser = User{username: "cool_guy_1337", password: "123123"}

	t.Log("Authorizing first user")
	err := db.AuthorizeUser(context.Background(), firstUser.username, firstUser.password)
	if err != nil {
		t.Error(err)
	}

	t.Log("Authorizing second user")
	err = db.AuthorizeUser(context.Background(), secondUser.username, secondUser.password)
	if err != nil {
		t.Error(err)
	}

	t.Log("Authorizing first user with wrong password")
	err = db.AuthorizeUser(context.Background(), firstUserWrongPassword.username, firstUserWrongPassword.password)
	if err == nil {
		t.Error("Should be wrong password error")
	}
	require.ErrorIs(t, err, cerr.ErrWrongPassword, "Error while authorizing with wrong password should be %v, not %v", cerr.ErrWrongPassword, err)

	t.Log("Checkin if balance is 1000")
	balance, err := db.GetUserBalance(context.Background(), firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, balance, 1000, "New user's balance should be equal 1000, not %v", balance)
	t.Log("Success.")
}

func TestBuy(t *testing.T) {
	t.Log("Testing buying")
	var err error
	err = db.Buy(context.Background(), firstUser.username, "cup")
	if err != nil {
		t.Error(err)
	}
	t.Log("Buying again")
	err = db.Buy(context.Background(), firstUser.username, "cup")
	if err != nil {
		t.Error(err)
	}

	t.Log("Checking inventory after buying thing")
	inv, err := db.GetUserInventory(context.Background(), firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Item{{Type: "cup", Quantity: 2}}, inv, "Inventory should be %v, not %v", []apimodels.Item{{Type: "cup", Quantity: 1}}, inv)

	balance, err := db.GetUserBalance(context.Background(), firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, 960, balance, "New user's balance should be equal 960, not %v", balance)

	t.Log("Buying non existing item")
	err = db.Buy(context.Background(), firstUser.username, "non existent item")
	if err == nil {
		t.Error("Should be unreal item error")
	}
	require.ErrorIs(t, err, cerr.ErrItemNotExist, "Error should be %v, not %v", cerr.ErrItemNotExist, err)
	t.Log("Success.")
}

func TestCoinTransfer(t *testing.T) {
	t.Log("Testing coin transfering")
	var err error

	t.Log("Sending coins from first to second user")
	err = db.SendCoins(context.Background(), firstUser.username, secondUser.username, 10)
	if err != nil {
		t.Error(err)
	}

	t.Log("Sending coins from second to first user")
	err = db.SendCoins(context.Background(), secondUser.username, firstUser.username, 20)
	if err != nil {
		t.Error(err)
	}

	t.Log("Sending coins to non-existent user")
	err = db.SendCoins(context.Background(), firstUser.username, "", 1)
	if err == nil {
		t.Error(fmt.Errorf("should be no user error"))
	}
	require.EqualError(t, cerr.ErrRecieverNotExist, err.Error(), "Error should be: %v. Got: %v", cerr.ErrRecieverNotExist, err.Error())

	t.Log("Trying to send more than possible")
	err = db.SendCoins(context.Background(), firstUser.username, secondUser.username, 100000)
	if err == nil {
		t.Error(fmt.Errorf("should be no money error"))
	}
	require.EqualError(t, cerr.ErrNoMoney, err.Error(), "Error should be: %v. Got: %v", cerr.ErrNoMoney, err.Error())

	t.Log("Trying send to myself")
	err = db.SendCoins(context.Background(), firstUser.username, firstUser.username, 10)
	if err == nil {
		t.Error(fmt.Errorf("should self send error"))
	}
	require.EqualError(t, cerr.ErrSelfSend, err.Error(), "Error should be: %v. Got: %v", cerr.ErrSelfSend, err.Error())

	t.Log("Getting recieving history")
	rhistory, err := db.GetUserRecieveHistory(context.Background(), firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Recieving{{FromUser: secondUser.username, Amount: 20}}, rhistory, "History should be %v, not %v", []apimodels.Recieving{{FromUser: secondUser.username, Amount: 20}}, rhistory)

	t.Log("Getting sending history")
	shistory, err := db.GetUserSendHistory(context.Background(), firstUser.username)
	if err != nil {
		t.Error(err)
	}
	require.EqualValues(t, []apimodels.Sending{{ToUser: secondUser.username, Amount: 10}}, shistory, "History should be %v, not %v", []apimodels.Sending{{ToUser: secondUser.username, Amount: 10}}, shistory)

}

func TestClearing(t *testing.T) {
	t.Log("Clearing after everythin")

	_, err := db.db.Exec(context.Background(), `delete from public.history`)
	if err != nil {
		t.Log("Failed cleaning users", err.Error())
	}
	_, err = db.db.Exec(context.Background(), `delete from public.inventory`)
	if err != nil {
		t.Log("Failed cleaning users", err.Error())
	}
	_, err = db.db.Exec(context.Background(), `delete from public.users`)
	if err != nil {
		t.Log("Failed cleaning users:", err.Error())
	}
}
