package router

import (
	cerr "AvitoTest/pkg/customErrors"
	"AvitoTest/pkg/models/apimodels"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

type Config struct {
	Port string `yaml:"port"`
}

type Router struct {
	App  *fiber.App
	cfg  Config
	jwt  IJWT
	auth IAuth
	cash ICash
}

type IJWT interface {
	GetUsernameFromToken(token string) (string, error)
	ValidateToken(c *fiber.Ctx, key string) (bool, error)
	AuthFilter(c *fiber.Ctx) bool
}
type IAuth interface {
	AuthorizeUser(apimodels.AuthRequest) (apimodels.AuthResponse, error)
}

type ICash interface {
	BuyItem(string, string) error
	UserInfo(string) (apimodels.InfoResponse, error)
	SendCoins(string, apimodels.SendCoinRequest) error
}

func New(cfg Config, jwt IJWT, auth IAuth, cash ICash) *Router {
	app := fiber.New()
	router := Router{App: app, jwt: jwt, auth: auth, cash: cash, cfg: cfg}
	router.App.Use(cors.New(cors.Config{
		AllowHeaders: "X-Access-Token, X-Refresh-Token",
	}))
	router.App.Use(keyauth.New(keyauth.Config{
		Next:         router.jwt.AuthFilter,
		KeyLookup:    "header:Authorization",
		Validator:    router.jwt.ValidateToken,
		ErrorHandler: router.ErrorHandler(),
	}))
	router.App.Post("/api/auth", router.Auth())
	router.App.Get("/api/info", router.Info())
	router.App.Get("/api/buy/:item", router.Buy())
	router.App.Post("/api/sendCoin", router.SendCoin())
	return &router
}

func (r *Router) Listen() error {
	return r.App.Listen(r.cfg.Port)
}

func (r *Router) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Info("ROUTER: got auth")
		var authCredentials apimodels.AuthRequest
		err := json.Unmarshal(c.Body(), &authCredentials)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		resp, err := r.auth.AuthorizeUser(authCredentials)
		if err != nil {
			if errors.Is(err, cerr.ErrWrongPassword) {
				c.Status(401)
				return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
			}
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		err = c.JSON(resp)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		c.Status(200)
		return nil
	}
}

func (r *Router) Info() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Info("ROUTER: got info")
		bearer := c.GetReqHeaders()["Authorization"]
		bearerToken := strings.Split(bearer[0], " ")[1]
		username, err := r.jwt.GetUsernameFromToken(bearerToken)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		slog.Info(fmt.Sprintf("ROUTER: got request from %v to get info", username))
		info, err := r.cash.UserInfo(username)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		err = c.JSON(info)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		c.Status(200)
		return nil
	}
}

func (r *Router) SendCoin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Info("ROUTER: got sendCoin")
		bearer := c.GetReqHeaders()["Authorization"]
		bearerToken := strings.Split(bearer[0], " ")[1]
		username, err := r.jwt.GetUsernameFromToken(bearerToken)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		var to apimodels.SendCoinRequest
		err = json.Unmarshal(c.Body(), &to)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		slog.Info(fmt.Sprintf("ROUTER: got request from %v to send %v coins to %v", username, to.Amount, to.ToUser))
		err = r.cash.SendCoins(username, to)
		if err != nil {
			if errors.Is(err, cerr.ErrRecieverNotExist) || errors.Is(err, cerr.ErrSelfSend) {
				c.Status(400)
				return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
			}
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		c.Status(200)
		return nil
	}
}

func (r *Router) Buy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Info("ROUTER: got buy")
		bearer := c.GetReqHeaders()["Authorization"]
		bearerToken := strings.Split(bearer[0], " ")[1]
		username, err := r.jwt.GetUsernameFromToken(bearerToken)
		if err != nil {
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		item := c.Params("item")
		slog.Info(fmt.Sprintf("ROUTER: got request from %v to buy %v", username, item))
		err = r.cash.BuyItem(username, item)
		if err != nil {
			if errors.Is(err, cerr.ErrItemNotExist) {
				c.Status(400)
				return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
			}
			c.Status(500)
			return c.JSON(apimodels.ErrorResponse{Errors: err.Error()})
		}
		c.Status(200)
		return nil
	}
}

func (r *Router) ErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		slog.Info("Wrong jwts: " + err.Error())
		c.Status(401)
		return err
	}
}
