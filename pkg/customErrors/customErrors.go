package cerr

import "fmt"

var (
	ErrNoToken          = fmt.Errorf("no authorization token")
	ErrRecieverNotExist = fmt.Errorf("no recievers with such username")
	ErrItemNotExist     = fmt.Errorf("no such items")
)
