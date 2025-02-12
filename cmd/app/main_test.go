package main

import (
	"AvitoTest/pkg/models/apimodels"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	token    string
}

func TestAll(t *testing.T) {
	url := "http://localhost:8080/api"
	users := []User{
		{
			Username: "perviy",
			Password: "password",
		},
		{
			Username: "vtoroy",
			Password: "password",
		},
	}
	client := &http.Client{}

	t.Log("Registering first user")
	first_user, err := json.Marshal(users[0])
	if err != nil {
		t.Error(err)
	}
	req, err := http.NewRequest("POST", url+"/auth", bytes.NewBuffer(first_user))
	if err != nil {
		t.Error(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	body, _ := io.ReadAll(resp.Body)
	authResp := apimodels.AuthResponse{}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		t.Error(err)
	}
	users[0].token = authResp.Token

	t.Log("Registering second user")
	second_user, err := json.Marshal(users[1])
	if err != nil {
		t.Error(err)
	}
	req, err = http.NewRequest("POST", url+"/auth", bytes.NewBuffer(second_user))
	if err != nil {
		t.Error(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	body, _ = io.ReadAll(resp.Body)
	authResp = apimodels.AuthResponse{}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		t.Error(err)
	}
	users[1].token = authResp.Token

	t.Log("Buying item as first user")
	req, err = http.NewRequest("GET", url+"/buy/cup", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Authorization", "Bearer "+users[0].token)
	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		errBody := apimodels.ErrorResponse{}
		body, _ = io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &errBody)
		if err != nil {
			t.Error(err)
		}
		t.Log("Buying error: " + errBody.Errors)
	}

	t.Log("Sending money to another user")
	sendRequest, err := json.Marshal(apimodels.SendCoinRequest{ToUser: users[1].Username, Amount: 100})
	if err != nil {
		t.Error(err)
	}
	req, err = http.NewRequest("POST", url+"/sendCoin", bytes.NewBuffer(sendRequest))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Authorization", "Bearer "+users[0].token)
	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		errBody := apimodels.ErrorResponse{}
		body, _ = io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &errBody)
		if err != nil {
			t.Error(err)
		}
		t.Log("Sending error: " + errBody.Errors)
	}

	t.Log("Getting info")
	req, err = http.NewRequest("GET", url+"/info", bytes.NewBuffer(sendRequest))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Authorization", "Bearer "+users[0].token)
	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		errBody := apimodels.ErrorResponse{}
		body, _ = io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &errBody)
		if err != nil {
			t.Error(err)
		}
		t.Log("Sending error: " + errBody.Errors)
	}
	info := apimodels.InfoResponse{}
	body, _ = io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &info)
	if err != nil {
		t.Error(err)
	}
	t.Log("Got log:", info)
	t.Log("Done. No errors means all routs work fine, right? (￢_￢;)")
}
