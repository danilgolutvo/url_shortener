package deleteURL

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	"url_shortener/httpServer/handlers"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"
)

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
	Info  string
}

type URLRemover interface {
	DeleteURL(alias, creator string) (bool, error)
}

func New(log *slog.Logger, urlRemover URLRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const info = "handlers.deleteURL.New"

		log := log.With(
			slog.String("info", info),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		params := mux.Vars(r)
		alias := params["alias"]

		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}
		creator, err := handlers.GetUserIDFromContext(r.Context())
		if err != nil {
			log.Error("could not get user id from context, unauthorized", sl.Err(err))
			render.JSON(w, r, resp.Error("could not ger user id from context, unauthorized"))
			return
		}
		ok, err := urlRemover.DeleteURL(alias, creator)
		if err != nil {
			// Check for specific error types or sentinel errors
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("Alias not found", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("alias not found"))
				return
			}

			// Handle other known errors if applicable
			if errors.Is(err, storage.ErrCaseMismatch) {
				log.Info("Alias has a case sensitivity issue", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("case sensitivity problem"))
				return
			}

			// Fallback for unexpected errors
			log.Error("Failed to delete URL", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		// Handle successful deletion
		if !ok {
			log.Info("Alias case mismatch prevented deletion", slog.String("alias", alias))
			render.JSON(w, r, resp.Error("case sensitivity problem"))
			return
		}

		log.Info("URL successfully deleted", slog.String("alias", alias))
		responseOK(w, r, alias)
	}
}
func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
		Info:     "Deleted",
	})
}
