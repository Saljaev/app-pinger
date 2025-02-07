package utilapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Error struct {
	ErrorMessage string `json:"error_message"`
}

type APIContext struct {
	w      http.ResponseWriter
	r      *http.Request
	log    *slog.Logger
	next   HandlerFunc
	ctx    context.Context
	cancel context.CancelFunc
}

func newAPIContext(w http.ResponseWriter, req *http.Request, log *slog.Logger) *APIContext {
	ctx, cancel := context.WithCancel(req.Context())

	r := req.WithContext(ctx)

	return &APIContext{
		w:      w,
		r:      r,
		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ctx *APIContext) WithContext(c context.Context, next HandlerFunc) *APIContext {
	ctx.ctx = c
	ctx.next = next
	return ctx
}

func (ctx *APIContext) Error(msg string, err error) {
	ctx.log.Error(msg, slog.Any("error", err))
}

func (ctx *APIContext) Info(msg string, key string, value interface{}) {
	ctx.log.Info(msg, slog.Any(key, value))
}

func (ctx *APIContext) Debug(msg string, key string, value interface{}) {
	ctx.log.Debug(msg, slog.Any(key, value))
}

type validator interface {
	IsValid() bool
}

func (ctx *APIContext) Decode(dest validator) error {
	err := json.NewDecoder(ctx.r.Body).Decode(&dest)
	if err != nil || !dest.IsValid() {
		if err == nil {
			err = errors.New("invalid request")
		}

		return err
	}

	return nil
}

func (ctx *APIContext) WriteFailure(code int, msg string) {
	ctx.w.WriteHeader(code)

	data, _ := json.Marshal(Error{ErrorMessage: msg})

	_, err := ctx.w.Write(data)
	if err != nil {
		ctx.Error("response write error", err)
	}
	ctx.cancel()
}

func (ctx *APIContext) SuccessWithData(data interface{}) {
	jsonData, _ := json.Marshal(data)

	ctx.w.WriteHeader(http.StatusOK)
	_, _ = ctx.w.Write(jsonData)
}

func (ctx *APIContext) GetFromQuery(key string) string {
	return ctx.r.Header.Get(key)
}

func (ctx *APIContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

func (ctx *APIContext) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *APIContext) Err() error {
	return ctx.ctx.Err()
}

func (ctx *APIContext) Value(key any) any {
	return ctx.ctx.Value(key)
}
