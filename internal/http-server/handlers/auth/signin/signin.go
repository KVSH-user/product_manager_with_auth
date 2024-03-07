package signin

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"inHouseAd/internal/entity"
	"inHouseAd/internal/lib/accesstoken"
	resp "inHouseAd/internal/lib/api/response"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var ErrInvalidEmail = errors.New("invalid email")

type Authorization interface {
	Authorizate(email string) ([]byte, int, error)
}

func LoginUser(log *slog.Logger, authorization Authorization, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.app.signin.LoginUser"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.UserAuthRequest
		var response entity.UserAuthResponse

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

		passwordHashed, id, err := authorization.Authorizate(req.Email)
		if err != nil {
			if errors.Is(err, ErrInvalidEmail) {
				log.Error("incorrect email: ", err)

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("incorrect credentials"))

				return
			}
			log.Error("failed to get password: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		if err := bcrypt.CompareHashAndPassword(passwordHashed, []byte(req.Password)); err != nil {
			log.Error("invalid password: ", err)

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid credential"))

			return
		}

		token, err := accesstoken.Generate(secret, id, time.Hour*24)

		response.Token = token

		log.Info("user successful login")

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response)
	}
}
