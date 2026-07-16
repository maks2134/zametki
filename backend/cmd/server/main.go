package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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
		AllowOrigins:     joinOrigins(cfg.CORSOrigins),
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

func joinOrigins(origins []string) string {
	if len(origins) == 0 {
		return "http://localhost:3000"
	}
	result := origins[0]
	for i := 1; i < len(origins); i++ {
		result += "," + origins[i]
	}
	return result
}
