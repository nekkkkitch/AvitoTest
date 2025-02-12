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
	"github.com/jackc/pgx/v5/pgxpool"
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
	db     *pgxpool.Pool
}

func New(cfg Config) (*DB, error) {
	d := &DB{config: cfg}
	connection := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := pgxpool.New(context.Background(), connection)
	slog.Info("Connecting to: " + connection)
	if err != nil {
		return nil, err
	}
	d.db = db
	return d, nil
}

func (d *DB) AuthorizeUser(ctx context.Context, username, password string) error {
	user, err := d.getUser(ctx, username)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Finding user error: " + "DB: " + err.Error())
			return err
		}
		cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("DB: " + err.Error())
			return err
		}
		_, err = d.db.Exec(ctx, `insert into public.users(username, password, balance) values($1, $2, $3)`, username, cryptedPassword, 1000)
		if err != nil {
			slog.Error("DB: " + err.Error())
			return err
		}
		return nil
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return cerr.ErrWrongPassword
		}
		slog.Error("DB: " + err.Error())
		return err
	}
	return nil
}

func (d *DB) Buy(ctx context.Context, username, itemTitle string) error {
	slog.Info(fmt.Sprintf("DB: %v trying to buy %v", username, itemTitle))
	var item dbmodels.Item
	err := d.db.QueryRow(ctx, `select * from public.merch where title = $1`, itemTitle).Scan(&item.Title, &item.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cerr.ErrItemNotExist
		}
		slog.Error("DB: " + err.Error())
		return err
	}
	conn, err := d.db.Acquire(ctx)
	if err != nil {
		slog.Error("DB: " + err.Error())
		return err
	}
	defer conn.Release()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("DB: " + err.Error())
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	_, err = tx.Exec(ctx, `update public.users set balance = balance-$1 where username=$2;`, item.Price, username)
	if err != nil {
		slog.Error("DB: " + err.Error())
		return cerr.ErrNoMoney
	}
	_, err = tx.Exec(ctx, `call add_item($1, $2);`, username, item.Title)
	if err != nil {
		slog.Error("DB: " + err.Error())
		return err
	}
	return nil
}
func (d *DB) GetUserBalance(ctx context.Context, username string) (int, error) {
	slog.Info(fmt.Sprintf("DB: %v trying to get balance", username))
	user, err := d.getUser(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, cerr.ErrUserNotExist
		}
		slog.Error("DB: " + err.Error())
		return -1, err
	}
	return int(user.Balance), nil
}
func (d *DB) GetUserInventory(ctx context.Context, username string) ([]apimodels.Item, error) {
	slog.Info(fmt.Sprintf("DB: %v trying to get inventory", username))
	rows, err := d.db.Query(ctx, `select item, amount from public.inventory where username=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Item{}, nil
		}
		slog.Error("DB: " + err.Error())
		return []apimodels.Item{}, err
	}
	defer rows.Close()
	inventory := make([]apimodels.Item, 0, 10)
	for rows.Next() {
		var part apimodels.Item
		err := rows.Scan(&part.Type, &part.Quantity)
		if err != nil {
			slog.Error("error while scanning inventory: " + "DB: " + err.Error())
			return []apimodels.Item{}, err
		}
		inventory = append(inventory, part)
	}
	return inventory, nil
}
func (d *DB) GetUserRecieveHistory(ctx context.Context, username string) ([]apimodels.Recieving, error) {
	slog.Info(fmt.Sprintf("DB: %v trying to get rhistory", username))
	rows, err := d.db.Query(ctx, `select sender, amount from public.history where reciever=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Recieving{}, nil
		}
		slog.Error("DB: " + err.Error())
		return []apimodels.Recieving{}, err
	}
	defer rows.Close()
	rhistory := make([]apimodels.Recieving, 0, 20)
	for rows.Next() {
		var part apimodels.Recieving
		err := rows.Scan(&part.FromUser, &part.Amount)
		if err != nil {
			slog.Error("error while scanning rhistory: " + "DB: " + err.Error())
			continue
		}
		rhistory = append(rhistory, part)
	}
	return rhistory, nil
}
func (d *DB) GetUserSendHistory(ctx context.Context, username string) ([]apimodels.Sending, error) {
	slog.Info(fmt.Sprintf("DB: %v trying to get shistory", username))
	rows, err := d.db.Query(ctx, `select reciever, amount from public.history where sender=$1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []apimodels.Sending{}, nil
		}
		slog.Error("DB: " + err.Error())
		return []apimodels.Sending{}, err
	}
	defer rows.Close()
	shistory := make([]apimodels.Sending, 0, 20)
	for rows.Next() {
		var part apimodels.Sending
		err := rows.Scan(&part.ToUser, &part.Amount)
		if err != nil {
			slog.Error("error while scanning rhistory: " + "DB: " + err.Error())
			continue
		}
		shistory = append(shistory, part)
	}
	return shistory, nil
}
func (d *DB) SendCoins(ctx context.Context, sender, reciever string, amount int) error {
	slog.Info(fmt.Sprintf("DB: %v trying to send %v to %v", sender, amount, reciever))
	if sender == reciever {
		return cerr.ErrSelfSend
	}
	if _, err := d.getUser(ctx, reciever); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cerr.ErrRecieverNotExist
		}
		slog.Error("DB:" + err.Error())
		return err
	}
	conn, err := d.db.Acquire(ctx)
	if err != nil {
		slog.Error("DB: " + "DB: " + err.Error())
		return err
	}
	defer conn.Release()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("DB: " + err.Error())
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	_, err = tx.Exec(ctx, `update public.users set balance = balance-$1 where username=$2;`, amount, sender)
	if err != nil {
		slog.Error("DB: " + err.Error())
		return cerr.ErrNoMoney
	}
	_, err = tx.Exec(ctx, `update public.users set balance = balance+$1 where username=$2;`, amount, reciever)
	if err != nil {
		slog.Error("DB: " + err.Error())
		return err
	}
	d.updateHistory(ctx, sender, reciever, amount)
	return nil
}

func (d *DB) updateHistory(ctx context.Context, sender, reciever string, amount int) {
	_, err := d.db.Exec(ctx, `insert into public.history values($1,$2,$3)`, sender, reciever, amount)
	if err != nil {
		slog.Error("error while updating history: " + "DB: " + err.Error())
	}
}

func (d *DB) getUser(ctx context.Context, username string) (dbmodels.User, error) {
	var user dbmodels.User
	err := d.db.QueryRow(ctx, `select * from public.users where username = $1`, username).Scan(&user.Username, &user.Password, &user.Balance)
	if err != nil {
		return user, err
	}
	return user, nil
}
