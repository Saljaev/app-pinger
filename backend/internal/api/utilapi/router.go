package utilapi

import (
	"log/slog"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
	log *slog.Logger
}

func NewRouter(log *slog.Logger) *Router {
	return &Router{
		mux: http.NewServeMux(),
		log: log,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
