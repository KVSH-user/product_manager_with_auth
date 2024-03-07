package category

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

type CreatorCategory interface {
	Create(name string, uid int) (int, error)
}

type EditorCategory interface {
	EditCategory(id int, newName string) (int, error)
}

type DeleterCategory interface {
	DeleteCategory(id int) error
}

type ListCategory interface {
	GetCategoryList() ([]entity.CategoryList, error)
}

func Create(log *slog.Logger, creatorCategory CreatorCategory, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.category.Create"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.CategoryCreateRequest
		var response entity.CategoryCreateResponse

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

		uid, err := uidextractor.ValidateToken(authHeader, secret)
		if err != nil {
			log.Error("user unauthorized: ", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		response.CategoryId, err = creatorCategory.Create(req.CategoryName, uid)
		if err != nil {
			log.Error("failed to create category: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response.CategoryName = req.CategoryName

		log.Info("category created")

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response)
	}
}

func EditCategory(log *slog.Logger, editorCategory EditorCategory, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.category.EditCategory"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.CategoryEditRequest
		var response entity.CategoryEditResponse

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

		response.CategoryId, err = editorCategory.EditCategory(req.CategoryId, req.NewName)
		if err != nil {
			log.Error("failed to edit category: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response.NewName = req.NewName

		log.Info("category edited")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}

func DeleteCategory(log *slog.Logger, deleterCategory DeleterCategory, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.category.DeleteCategory"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var response entity.CategoryDeleteResponse

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("user unauthorized: authorization header is missing")
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		categoryId := chi.URLParam(r, "id")
		if categoryId == "" {
			log.Info("category id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("category id parameter is required"))
			return
		}

		_, err := uidextractor.ValidateToken(authHeader, secret)
		if err != nil {
			log.Error("user unauthorized: ", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		CategoryIdInt, err := strconv.Atoi(categoryId)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		err = deleterCategory.DeleteCategory(CategoryIdInt)
		if err != nil {
			if err == postgres.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)

				return
			}
			log.Error("failed to delete category: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		response.CategoryId = CategoryIdInt
		response.Deleted = true

		log.Info("category deleted")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}

func GetCategoryList(log *slog.Logger, listCategory ListCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goodsservice.category.GetCategoryList"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var response []entity.CategoryList

		response, err := listCategory.GetCategoryList()
		if err != nil {
			if err == postgres.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)

				return
			}
			log.Error("failed to get category list: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("category list geted")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}
