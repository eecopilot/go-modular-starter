package auth

import (
	"net/http"

	"github.com/eecopilot/userkit"
	"github.com/eecopilot/userkit/httpadapter"
)

const Prefix = "/api/v1"

type Module struct {
	handler *httpadapter.Handler
}

func New(service *userkit.Service) *Module {
	return &Module{handler: httpadapter.New(service)}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	sub := http.NewServeMux()
	m.handler.RegisterRoutes(sub)

	mux.Handle(Prefix+"/", http.StripPrefix(Prefix, sub))
}
