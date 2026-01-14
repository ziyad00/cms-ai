package api

import (
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type Server struct {
	Config        Config
	Authenticator auth.Authenticator
	Store         store.Store
	Validator     spec.Validator
	Renderer      assets.Renderer
	ObjectStorage assets.ObjectStorage
	AIService     ai.AIServiceInterface
}
