package cerr

import "fmt"

var (
	ErrUserNotExist     = fmt.Errorf("no user with such login")
	ErrNoToken          = fmt.Errorf("no authorization token")
	ErrRecieverNotExist = fmt.Errorf("no recievers with such username")
	ErrItemNotExist     = fmt.Errorf("no such items")
	ErrWrongPassword    = fmt.Errorf("wrong password")
	ErrNoMoney          = fmt.Errorf("not enough credits to buy this item")
	ErrSelfSend         = fmt.Errorf("can't send cash to yourself")
)
