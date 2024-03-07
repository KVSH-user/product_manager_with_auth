package good

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"inHouseAd/internal/entity"
	"inHouseAd/internal/http-server/handlers/auth/uidextractor"
	resp "inHouseAd/internal/lib/api/response"
	"inHouseAd/internal/storage/postgres"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type AdderGood interface {
	AddGood(goodName string, categoryId int) (int, string, error)
}

type UpdaterGood interface {
	UpdateGood(goodId, categoryIdToAdd int, goodName string) (int, []string, string, error)
}

type DeleterGood interface {
	DeleteGood(id int) error
}

type ListGood interface {
	GetGoodList(categoryId int) ([]entity.GoodList, error)
}

func Create(log *slog.Logger, adderGood AdderGood, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.good.Create"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.GoodAddRequest
		var response entity.GoodAddResponse

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("user unauthorized: authorization header is missing")
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		categoryId := chi.URLParam(r, "categoryId")
		if categoryId == "" {
			log.Info("category id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("category id parameter is required"))
			return
		}

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body: ", err)

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		_, err = uidextractor.ValidateToken(authHeader, secret)
		if err != nil {
			log.Error("user unauthorized: ", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		categoryIdInt, err := strconv.Atoi(categoryId)
		if err != nil {
			http.Error(w, "invalid category ID", http.StatusBadRequest)
			return
		}

		response.GoodId, response.CategoryName, err = adderGood.AddGood(req.GoodName, categoryIdInt)
		if err != nil {
			log.Error("failed to create category: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response.GoodName = req.GoodName
		response.GoodCategoryId = categoryIdInt

		log.Info("category created")

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response)
	}
}

func UpdateGood(log *slog.Logger, updaterGood UpdaterGood, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.good.UpdateGood"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.GoodUpdateRequest

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("user unauthorized: authorization header is missing")
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body: ", err)

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		_, err = uidextractor.ValidateToken(authHeader, secret)
		if err != nil {
			log.Error("user unauthorized: ", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		goodId, categoryNames, goodName, err := updaterGood.UpdateGood(req.GoodId, req.AddedCategoryId, req.GoodActualName)
		if err != nil {
			log.Error("failed to update good: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response := entity.GoodUpdateResponse{
			GoodId:       goodId,
			CategoryName: categoryNames,
			GoodName:     goodName,
		}

		log.Info("good updated")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}

func DeleteGood(log *slog.Logger, deleterGood DeleterGood, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.good.DeleteGood"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var response entity.GoodDeleteResponse

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("user unauthorized: authorization header is missing")
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		goodId := chi.URLParam(r, "id")
		if goodId == "" {
			log.Info("good id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("good id parameter is required"))
			return
		}

		_, err := uidextractor.ValidateToken(authHeader, secret)
		if err != nil {
			log.Error("user unauthorized: ", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		GoodIdInt, err := strconv.Atoi(goodId)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		err = deleterGood.DeleteGood(GoodIdInt)
		if err != nil {
			if err == postgres.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)

				return
			}
			log.Error("failed to delete good: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response.GoodId = GoodIdInt
		response.Deleted = true

		log.Info("good deleted")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}

func GetGoodList(log *slog.Logger, listGood ListGood) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.good.GetGoodList"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var response []entity.GoodList

		categoryId := chi.URLParam(r, "categoryId")
		if categoryId == "" {
			log.Info("category id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("category id parameter is required"))
			return
		}

		categoryIdInt, err := strconv.Atoi(categoryId)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		response, err = listGood.GetGoodList(categoryIdInt)
		if err != nil {
			if err == postgres.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)

				return
			}
			log.Error("failed to get good list: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("good list geted ")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}
