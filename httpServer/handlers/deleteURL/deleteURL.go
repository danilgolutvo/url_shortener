package deleteURL

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
)

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
	Info  string
}

type URLRemover interface {
	DeleteURL(alias string) (bool, error)
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
		ok, err := urlRemover.DeleteURL(alias)
		if !ok {
			log.Info("alias is not written correctly")
			render.JSON(w, r, resp.Error("case sensitivity problem"))
			return
		}
		if errors.Is(err, errors.New("alias not found")) {
			log.Info("alias not found", "alias", alias)
			render.JSON(w, r, resp.Error("alias not found"))
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("url successfully deleted", slog.String("alias", alias))
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
