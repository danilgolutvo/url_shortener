package redirect

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.49.1 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const info = "handlers.redirect.New"

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

		resultURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("url gotten", slog.String("url", resultURL))
		http.Redirect(w, r, resultURL, http.StatusFound)
	}
}
