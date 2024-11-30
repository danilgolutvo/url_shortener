package save

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"io"
	"log/slog"
	"net/http"
	"url_shortener/cmd/middleware"
	"url_shortener/httpServer/handlers/url/random"
	"url_shortener/internal/config"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ErrorValidator(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.RandomString(config.MustLoad().AliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

//type Request struct {
//	URL   string `json:"url" validate:"required,url"`
//	Alias string `json:"alias,omitempty"`
//}
//
//type Response struct {
//	response.Response
//	Alias string `json:"alias,omitempty"`
//}
//
////go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
//type URLSaver interface {
//	SaveURL(urlToSave string, alias string) (int64, error)
//	//AliasExists(alias string) (bool, error)
//}
//
//func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		const info = "httpServer.handlers.url.save.New"
//
//		log := log.With(
//			slog.String("info", info),
//			slog.String("request_id", middleware.GetReqID(r.Context())),
//		)
//
//		var req Request
//
//		err := render.DecodeJSON(r.Body, &req)
//		if err != nil {
//			log.Error("failed to decode request body", sl.Err(err))
//
//			render.JSON(w, r, response.Error("failed to decode request"))
//
//			return
//		}
//
//		log.Info("request body's been decoded", slog.Any("request", req))
//
//		// validate request that has come
//		err = validator.New().Struct(req)
//		if err != nil {
//			validateErr := err.(validator.ValidationErrors)
//
//			log.Error("invalid request", sl.Err(err))
//
//			render.JSON(w, r, response.ErrorValidator(validateErr))
//		}
//
//		alias := req.Alias
//		if alias == "" {
//			alias = random.RandomString(config.MustLoad().AliasLength)
//		}
//		//ok, err := urlSaver.AliasExists(alias)
//		//if ok == true {
//		//	log.Info("alias exists", slog.String("alias", req.Alias))
//		//
//		//	render.JSON(w, r, response.Error("alias exists"))
//		//
//		//	return
//		//}
//		//if err != nil {
//		//	log.Error("failed to check if alias exists", sl.Err(err))
//		//
//		//	render.JSON(w, r, response.Error("failed to check if alias exists"))
//		//
//		//	return
//		//}
//
//		id, err := urlSaver.SaveURL(req.URL, alias)
//		if errors.Is(err, storage.ErrURLExists) {
//			log.Info("url already exists", slog.String("url", req.URL))
//
//			render.JSON(w, r, response.Error("url already exists"))
//
//			return
//		}
//		if err != nil {
//			log.Error("failed to add url", sl.Err(err))
//
//			render.JSON(w, r, response.Error("failed to add url"))
//
//			return
//		}
//
//		log.Info("url added", slog.Int64("id", id))
//
//		responseOK(w, r, alias)
//	}
//}
//
//func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
//	render.JSON(w, r, Response{
//		Response: response.OK(),
//		Alias:    alias,
//	})
//}
