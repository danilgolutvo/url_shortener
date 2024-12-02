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
			log.Info("could not decode body")
			render.JSON(w, r, resp.Error("could not decode"))
			return
		}
		user, err := logingHandler.GetUserByUsername(loginReq.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Info("did not find user")
				render.JSON(w, r, resp.Error("did not find user"))
				http.Error(w, "invalid username or password", http.StatusUnauthorized)
			} else {
				log.Info("server error")
				render.JSON(w, r, resp.Error("server error"))
			}
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}

		token, err := createToken(user.ID, user.Username)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}
