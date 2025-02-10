package dbmodels

type User struct {
	Username string `db:"username"`
	Password []byte `db:"password"`
	Balance  int32  `db:"balance"`
}

type Item struct {
	Title string `db:"title"`
	Price int    `db:"price"`
}
