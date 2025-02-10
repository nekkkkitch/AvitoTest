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
		_, err = d.db.Exec(context.TODO(), `insert into public.users(username, password, amount) values($1, $2, $3)`, username, cryptedPassword, 1000)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		return nil
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (d *DB) Buy(username, itemTitle string) error {
	var item dbmodels.Item
	err := d.db.QueryRow(context.Background(), `select * from public.items where title = $1`, itemTitle).Scan(&item)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cerr.ErrItemNotExist
		}
		slog.Error(err.Error())
		return err
	}
	d.db.Exec(context.Background(), `begin transaction;`)
	d.db.Exec(context.Background(), `update public.users set balance = balance-$1 where username=$2;`, item.Price, username)
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
	var inventory []apimodels.Item
	rows, err := d.db.Query(context.Background(), `select * from public.inventory where username=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Item{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Item{}, err
	}
	defer rows.Close()
	err = rows.Scan(inventory)
	if err != nil {
		slog.Error(err.Error())
		return []apimodels.Item{}, err
	}
	return inventory, nil
}
func (d *DB) GetUserRecieveHistory(username string) ([]apimodels.Recieving, error) {
	var rhistory []apimodels.Recieving
	rows, err := d.db.Query(context.Background(), `select reciever, amount from public.history where reciever=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Recieving{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Recieving{}, err
	}
	defer rows.Close()
	err = rows.Scan(rhistory)
	if err != nil {
		slog.Error(err.Error())
		return []apimodels.Recieving{}, err
	}
	return rhistory, nil
}
func (d *DB) GetUserSendHistory(username string) ([]apimodels.Sending, error) {
	var shistory []apimodels.Sending
	rows, err := d.db.Query(context.Background(), `select sender, amount from public.history where sender=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Sending{}, nil
		}
		slog.Error(err.Error())
		return []apimodels.Sending{}, err
	}
	defer rows.Close()
	err = rows.Scan(shistory)
	if err != nil {
		slog.Error(err.Error())
		return []apimodels.Sending{}, err
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
	return nil
}

func (d *DB) getUser(username string) (dbmodels.User, error) {
	var user dbmodels.User
	err := d.db.QueryRow(context.Background(), `select * from public.users where username = $1`, username).Scan(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}
