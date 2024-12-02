package login

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	resp "url_shortener/internal/lib/api/response"
)

type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.49.1 --name=LoginHandler --output=url_shortener/test
type LoginHandler interface {
	GetUserByUsername(username string) (User, error)
}

func HandleLogin(log *slog.Logger, logingHandler LoginHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const info = "handlers.login.HandleLogin"

		log := log.With(
			slog.String("info", info),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var loginReq User
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			log.Error("could not decode request body", slog.String("error", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("could not decode"))
			return
		}
		user, err := logingHandler.GetUserByUsername(loginReq.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Error("found no rows", slog.String("error", err.Error()))
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("invalid username or password"))
				return
			} else {
				log.Info("server error")
				render.JSON(w, r, resp.Error("server error"))
			}
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
			log.Error("invalid username or password", slog.String("error", err.Error()))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, resp.Error("invalid username or password"))
			return
		}

		token, err := CreateToken(user.ID, user.Username)
		if err != nil {
			log.Error("server error", slog.String("error", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("invalid username or password"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}
