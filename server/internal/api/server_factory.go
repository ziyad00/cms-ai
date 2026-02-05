package api

import (
	"context"
	"log"
	"os"

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
	config := LoadConfig()
	// Use JWT authenticator if JWT_SECRET is set, otherwise fall back to header auth for dev
	var authenticator auth.Authenticator
	if os.Getenv("JWT_SECRET") != "" {
		authenticator = auth.JWTAuthenticator{}
	} else {
		authenticator = auth.HeaderAuthenticator{}
	}
	validator := spec.DefaultValidator{}

	// Create object storage
	factory := assets.NewStorageFactory()
	objectStorage, err := factory.CreateStorage(context.Background())
	if err != nil {
		panic("failed to create object storage: " + err.Error())
	}

	var st store.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		log.Printf("Using PostgreSQL storage with DSN: %s", dsn[:20]+"...")
		pg, err := postgres.New(dsn)
		if err != nil {
			panic("failed to connect to postgres: " + err.Error())
		}
		st = pg
		log.Printf("PostgreSQL connection successful")
	} else {
		log.Printf("Using in-memory storage (no DATABASE_URL)")
		st = memory.New()
	}

	// Use AI-enhanced Python renderer as default (with Olama backgrounds) - force rebuild v2
	var renderer assets.Renderer
	if os.Getenv("HUGGINGFACE_API_KEY") != "" {
		log.Printf("Using AI-enhanced Python renderer with Hugging Face (default)")
		renderer = assets.NewAIEnhancedRenderer(st)
	} else {
		log.Printf("Using Python PPTX renderer (no AI key)")
		renderer = &assets.PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: "/app/tools/renderer/render_pptx.py",
		}
	}

	// Create AI service
	aiService := ai.NewAIService(st)

	return &Server{
		Config:        config,
		Authenticator: authenticator,
		Store:         st,
		Validator:     validator,
		Renderer:      renderer,
		ObjectStorage: objectStorage,
		AIService:     aiService,
	}
}

func NewServerWithWorker() (*Server, *worker.Worker) {
	srv := NewServer()
	// Note: Worker still uses old Storage interface, need to update worker separately
	w := worker.New(srv.Store, srv.Renderer, nil) // TODO: Update worker to use ObjectStorage
	return srv, w
}
