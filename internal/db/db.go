package db

import (
	cerr "AvitoTest/pkg/customErrors"
	"AvitoTest/pkg/models/apimodels"
	"AvitoTest/pkg/models/dbmodels"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type DB struct {
	config Config
	db     *pgx.Conn
}

func New(cfg Config) (*DB, error) {
	d := &DB{config: cfg}
	connection := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := pgx.Connect(context.Background(), connection)
	slog.Info("Connecting to: " + connection)
	if err != nil {
		return nil, err
	}
	d.db = db
	return d, nil
}

func (d *DB) AuthorizeUser(username, password string) error {
	user, err := d.getUser(username)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Finding user error: " + err.Error())
			return err
		}
		cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		_, err = d.db.Exec(context.TODO(), `insert into public.users(username, password, balance) values($1, $2, $3)`, username, cryptedPassword, 1000)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		return nil
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return cerr.ErrWrongPassword
		}
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (d *DB) Buy(username, itemTitle string) error {
	var item dbmodels.Item
	err := d.db.QueryRow(context.Background(), `select * from public.merch where title = $1`, itemTitle).Scan(&item.Title, &item.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cerr.ErrItemNotExist
		}
		slog.Error(err.Error())
		return err
	}
	d.db.Exec(context.Background(), `begin transaction;`)
	d.db.Exec(context.Background(), `update public.users set balance = balance-$1 where username=$2;`, item.Price, username)
	d.db.Exec(context.Background(), `call add_item($1, $2);`, username, item.Title)
	_, err = d.db.Exec(context.Background(), `commit;`)
	if err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return cerr.ErrNoMoney
		}
		slog.Error(err.Error())
		return err
	}
	return nil
}
func (d *DB) GetUserBalance(username string) (int, error) {
	user, err := d.getUser(username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, cerr.ErrUserNotExist
		}
		slog.Error(err.Error())
		return -1, err
	}
	return int(user.Balance), nil
}
func (d *DB) GetUserInventory(username string) ([]apimodels.Item, error) {
	rows, err := d.db.Query(context.Background(), `select item, amount from public.inventory where username=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Item{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Item{}, err
	}
	defer rows.Close()
	inventory := make([]apimodels.Item, 0, 10)
	for rows.Next() {
		var part apimodels.Item
		err := rows.Scan(&part.Type, &part.Quantity)
		if err != nil {
			slog.Error("error while scanning inventory: " + err.Error())
			return []apimodels.Item{}, err
		}
		inventory = append(inventory, part)
	}
	return inventory, nil
}
func (d *DB) GetUserRecieveHistory(username string) ([]apimodels.Recieving, error) {
	rows, err := d.db.Query(context.Background(), `select sender, amount from public.history where reciever=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Recieving{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Recieving{}, err
	}
	defer rows.Close()
	rhistory := make([]apimodels.Recieving, 0, 20)
	for rows.Next() {
		var part apimodels.Recieving
		err := rows.Scan(&part.FromUser, &part.Amount)
		if err != nil {
			slog.Error("error while scanning rhistory: " + err.Error())
			continue
		}
		rhistory = append(rhistory, part)
	}
	return rhistory, nil
}
func (d *DB) GetUserSendHistory(username string) ([]apimodels.Sending, error) {
	rows, err := d.db.Query(context.Background(), `select reciever, amount from public.history where sender=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Sending{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Sending{}, err
	}
	defer rows.Close()
	shistory := make([]apimodels.Sending, 0, 20)
	for rows.Next() {
		var part apimodels.Sending
		err := rows.Scan(&part.ToUser, &part.Amount)
		if err != nil {
			slog.Error("error while scanning rhistory: " + err.Error())
			continue
		}
		shistory = append(shistory, part)
	}
	return shistory, nil
}
func (d *DB) SendCoins(sender, reciever string, amount int) error {
	d.db.Exec(context.Background(), `begin transaction;`)
	d.db.Exec(context.Background(), `update public.users set balance = balance-$1 where username=$2;`, amount, sender)
	d.db.Exec(context.Background(), `update public.users set balance = balance+$1 where username=$2;`, amount, reciever)
	_, err := d.db.Exec(context.Background(), `commit;`)
	if err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return cerr.ErrNoMoney
		}
		slog.Error(err.Error())
		return err
	}
	d.updateHistory(sender, reciever, amount)
	return nil
}

func (d *DB) updateHistory(sender, reciever string, amount int) {
	_, err := d.db.Exec(context.Background(), `insert into public.history values($1,$2,$3)`, sender, reciever, amount)
	if err != nil {
		slog.Error("error while updating history: " + err.Error())
	}
}

func (d *DB) getUser(username string) (dbmodels.User, error) {
	var user dbmodels.User
	err := d.db.QueryRow(context.Background(), `select * from public.users where username = $1`, username).Scan(&user.Username, &user.Password, &user.Balance)
	if err != nil {
		return user, err
	}
	return user, nil
}
