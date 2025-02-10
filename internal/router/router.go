package router

import (
	"AvitoTest/pkg/models"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Config struct {
	Port string `yaml:"port"`
}

type Router struct {
	App  *fiber.App
	jwt  IJWT
	auth IAuth
	cash ICash
}

type IJWT interface {
	GetUsernameFromToken(token string) (string, error)
	ValidateToken(c *fiber.Ctx, key string) (bool, error)
	AuthFilter(c *fiber.Ctx) bool
	RefreshFilter(c *fiber.Ctx) bool
}
type IAuth interface{}

type ICash interface {
	BuyItem(string, string) error
	UserInfo(string) (int, []models.Item, models.CoinHistory, error)
	SendCoins(string, string, int) error
}

func New(cfg Config, jwt IJWT, auth IAuth, cash ICash) *Router {
	app := fiber.New()
	router := Router{App: app, jwt: jwt, auth: auth, cash: cash}
	router.App.Post("/api/auth", router.Auth())
	router.App.Get("/api/info", router.Info())
	router.App.Get("/api/buy/:item", router.Buy())
	router.App.Post("/api/sendCoin", router.SendCoin())
	return &router
}

func (r *Router) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {

		return nil
	}
}

func (r *Router) Info() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := c.GetReqHeaders()["Authorization"][0]
		bearerToken := strings.Split(bearer, " ")[1]
		userID, err := r.jwt.GetUsernameFromToken(bearerToken)
		r.cash.UserInfo()
		return nil
	}
}

func (r *Router) SendCoin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func (r *Router) Buy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func (r *Router) ErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		slog.Info("Wrong jwts: " + err.Error())
		return err
	}
}
