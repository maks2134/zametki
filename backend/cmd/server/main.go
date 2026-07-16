package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"zametka/internal/auth"
	"zametka/internal/config"
	"zametka/internal/db"
	pgrepo "zametka/internal/repository/postgres"
	"zametka/internal/service"
	httptransport "zametka/internal/transport/http"
	"zametka/internal/transport/ws"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	issuer := auth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
	hub := ws.NewHub()
	go hub.Run(ctx)

	roomRepo := pgrepo.NewRoomRepository(pool)
	noteRepo := pgrepo.NewNoteRepository(pool)

	roomSvc := service.NewRoomService(roomRepo, issuer, hub)
	noteSvc := service.NewNoteService(noteRepo, hub)

	app := fiber.New(fiber.Config{
		ErrorHandler: httptransport.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return allowOrigin(origin, cfg.CORSOrigins)
		},
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type",
		AllowCredentials: false,
	}))

	httptransport.RegisterRoutes(app, roomSvc, noteSvc, issuer)
	ws.RegisterWS(app, hub, issuer)

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := app.Listen(cfg.HTTPAddr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	cancel()
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func allowOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	// Vercel preview + production URLs change; allow any https://*.vercel.app
	if strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".vercel.app") {
		return true
	}
	if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
		return true
	}
	return false
}
