package utilapi

import (
	"fmt"
	"log/slog"
	"net/http"
)

type HandlerFunc func(ctx *APIContext)

func (r *Router) Handle(pattern string, handlerFuncs ...HandlerFunc) {
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := newAPIContext(w, req, r.log)

		ctx.log = ctx.log.With(slog.String("pattern", fmt.Sprintf("%s", pattern)))
		ctx.w.Header().Set("Content-Type", "application/json; charset=utf-8")

		for _, h := range handlerFuncs {
			select {
			case <-ctx.Done():
				return
			default:
				h(ctx)
			}
		}
	})
}
