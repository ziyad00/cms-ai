package api

import (
	"context"
	"fmt"
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
	fmt.Println("🚀 STARTING SERVER INITIALIZATION...")
	config := LoadConfig()
	authenticator := auth.JWTAuthenticator{}
	validator := spec.DefaultValidator{}

	// Create object storage
	fmt.Println("📦 INITIALIZING OBJECT STORAGE...")
	factory := assets.NewStorageFactory()
	objectStorage, err := factory.CreateStorage(context.Background())
	if err != nil {
		fmt.Printf("❌ OBJECT STORAGE FAILED: %v\n", err)
		panic("failed to create object storage: " + err.Error())
	}

	var st store.Store
	dsn := os.Getenv("DATABASE_URL")
	fmt.Printf("🗄️ DATABASE_URL LENGTH: %d\n", len(dsn))
	
	if dsn != "" {
		fmt.Println("🔌 CONNECTING TO POSTGRES...")
		pg, err := postgres.New(dsn)
		if err != nil {
			fmt.Printf("❌ POSTGRES CONNECTION FAILED: %v\n", err)
			fmt.Println("⚠️ FALLING BACK TO IN-MEMORY STORAGE TO PREVENT PANIC")
			st = memory.New()
		} else {
			st = pg
			fmt.Println("✅ POSTGRES CONNECTED SUCCESS")
		}
	} else {
		fmt.Println("⚠️ USING IN-MEMORY STORAGE (NO DATABASE_URL)")
		st = memory.New()
	}

	// Create AI service
	fmt.Println("🤖 INITIALIZING AI SERVICE...")
	aiService := ai.NewAIService(st)

	fmt.Println("🎨 INITIALIZING RENDERER...")
	var renderer assets.Renderer
	if os.Getenv("HUGGINGFACE_API_KEY") != "" {
		renderer = assets.NewAIEnhancedRenderer(st)
	} else {
		renderer = assets.NewPythonPPTXRenderer("")
	}

	fmt.Println("✅ SERVER INITIALIZATION COMPLETE")
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
