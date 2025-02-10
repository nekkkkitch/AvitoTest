package apimodels

// Requests
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

// Responses
type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

// Essentials
type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Recieved []Recieving `json:"recieved"`
	Sent     []Sending   `json:"sent"`
}

type Recieving struct {
	FromUser string `json:"fromUser" db:"reciever"`
	Amount   int    `json:"amount" db:"amount"`
}

type Sending struct {
	ToUser string `json:"toUser" db:"sender"`
	Amount int    `json:"amount" db:"amount"`
}
