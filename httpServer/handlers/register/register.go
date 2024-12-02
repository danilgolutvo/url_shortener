package register

import (
	"encoding/json"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	"url_shortener/httpServer/handlers/login"
	resp "url_shortener/internal/lib/api/response"
)

type RegistrationHandler interface {
	CreateUser(user login.User) error
}

func HandleRegistration(log *slog.Logger, handler RegistrationHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const info = "handlers.login.HandleRegistration"

		log := log.With(
			slog.String("info", info),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var loginReq login.User
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			log.Error("could not decode request body", slog.String("error", err.Error()))
			render.JSON(w, r, resp.Error("could not decode"))
			return
		}
		log.Info("here is the decoded request", loginReq)
		if loginReq.ID == "" {
			loginReq.ID = uuid.NewString()
		}
		err := handler.CreateUser(loginReq)
		if err != nil {
			log.Error("failed to create user", slog.String("error", err.Error()))
			render.JSON(w, r, resp.Error("server error"))
			return
		}
		log.Info("registration is done successfully", slog.String("username", loginReq.Username))

		responseOK(w, r)

	}
}
func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, "registration is done successfully JSON")
}
