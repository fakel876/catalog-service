package rhealth

import (
	"net/http"

	rhandler "github.com/fakel876/catalog-service/internal/app/handler/http"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (*handler) LastCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
