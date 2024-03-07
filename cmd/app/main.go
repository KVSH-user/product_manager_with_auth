package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"inHouseAd/internal/config"
	"inHouseAd/internal/http-server/handlers/auth/signin"
	"inHouseAd/internal/http-server/handlers/auth/signup"
	"inHouseAd/internal/http-server/handlers/goodsservice/category"
	"inHouseAd/internal/http-server/handlers/goodsservice/good"
	"inHouseAd/internal/http-server/middleware/logger"
	"inHouseAd/internal/lib/goodgetter"
	"inHouseAd/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad("config/config.yaml")

	log := SetupLogger(cfg.Env)

	log.Info("App started", slog.String("env", cfg.Env))
	log.Debug("Debugging started")

	jwtSecret := cfg.Auth.JwtSecret

	storage, err := postgres.New(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
	)
	if err != nil {
		log.Error("failed to init storage: ", err)
		os.Exit(1)
	}

	log.Info("storage successfully initialized")

	go periodicGoodFetch(log, cfg.API.Url, storage)

	router := chi.NewRouter()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(corsHandler.Handler)

	router.Post("/user/signup", signup.CreateUser(log, storage))
	router.Post("/user/signin", signin.LoginUser(log, storage, jwtSecret))
	router.Post("/category/create", category.Create(log, storage, jwtSecret))
	router.Patch("/category/update", category.EditCategory(log, storage, jwtSecret))
	router.Delete("/category/delete/{id}", category.DeleteCategory(log, storage, jwtSecret))
	router.Post("/good/create/{categoryId}", good.Create(log, storage, jwtSecret))
	router.Patch("/good/update", good.UpdateGood(log, storage, jwtSecret))
	router.Delete("/good/delete/{id}", good.DeleteGood(log, storage, jwtSecret))
	router.Get("/category/list", category.GetCategoryList(log, storage))
	router.Get("/good/list/{categoryId}", good.GetGoodList(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func periodicGoodFetch(log *slog.Logger, apiURL string, adderGood good.AdderGood) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			goodgetter.GetGoodFromAPI(log, apiURL, adderGood)
		}
	}
}
