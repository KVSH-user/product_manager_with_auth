package signup

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"inHouseAd/internal/entity"
	resp "inHouseAd/internal/lib/api/response"
	"io"
	"log/slog"
	"net/http"
)

var ErrEmailTaken = errors.New("email already taken")

type Registration interface {
	Register(email string, passwordHashed []byte) (int, error)
}

func CreateUser(log *slog.Logger, registration Registration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.app.signup.CreateUser"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req entity.UserRegisterRequest
		var response entity.UserRegisterResponse

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

		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password hash: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		id, err := registration.Register(req.Email, passwordHashed)
		if err != nil {
			if errors.Is(err, ErrEmailTaken) {
				log.Error("email already taken: ", err)
				render.JSON(w, r, resp.Error("email already taken"))
				return
			}

			log.Error("failed to create user: ", err)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		response.Id = id

		log.Info("user created")

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response)
	}
}
