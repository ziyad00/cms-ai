package api

import (
	"context"
	"log"
	"os"

	lib_validator "github.com/go-playground/validator/v10"
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
	"github.com/ziyad/cms-ai/server/internal/store/postgres"
	"github.com/ziyad/cms-ai/server/internal/worker"
)

func NewServer() *Server {
	log.Println("Starting server initialization...")
	config := LoadConfig()
	authenticator := auth.JWTAuthenticator{}
	validator := spec.DefaultValidator{}

	// Create object storage (fall back to local if cloud fails)
	factory := assets.NewStorageFactory()
	objectStorage, err := factory.CreateStorage(context.Background())
	if err != nil {
		log.Printf("Object storage init failed (%v), using local storage", err)
		objectStorage, _ = assets.NewLocalStorage(assets.StorageConfig{Type: "local", BasePath: "/tmp/cms-ai-assets"})
	}

	var st store.Store
	dsn := os.Getenv("DATABASE_URL")

	if dsn != "" {
		pg, err := postgres.New(dsn)
		if err != nil {
			log.Printf("Postgres connection failed: %v. Falling back to in-memory store.", err)
			st = memory.New()
		} else {
			st = pg
			log.Println("Connected to PostgreSQL")
		}
	} else {
		log.Println("No DATABASE_URL set, using in-memory store")
		st = memory.New()
	}

	aiService := ai.NewAIService(st)

	var renderer assets.Renderer
	if os.Getenv("HUGGINGFACE_API_KEY") != "" {
		renderer = assets.NewAIEnhancedRenderer(st)
	} else {
		renderer = assets.NewPythonPPTXRenderer("")
	}

	log.Println("Server initialization complete")
	return &Server{
		Config:        config,
		Authenticator: authenticator,
		Store:         st,
		Validator:     validator,
		Renderer:      renderer,
		ObjectStorage: objectStorage,
		AIService:     aiService,
		validate:      lib_validator.New(),
	}
}

func NewServerWithWorker() (*Server, *worker.Worker) {
	srv := NewServer()
	// Create worker with the same object storage as the server
	w := worker.New(srv.Store, srv.Renderer, srv.ObjectStorage, srv.AIService)
	return srv, w
}
