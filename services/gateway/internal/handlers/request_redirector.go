package handlers

import (
	"fmt"
	"net/http"

	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/handlers/middleware"
)

type Handler struct {
	service *config.Service
}

func NewHandler(service *config.Service) *Handler {
	return &Handler{service}
}

func (h *Handler) RegisterEndpoints(mux *http.ServeMux) {
	prefix := fmt.Sprintf("/api/%s", h.service.ApiVersion)
	addPrefix := func(method, path string) string {
		return fmt.Sprintf("%s %s%s", method, prefix, path)
	}

	switch h.service.Name {
	case "inventory service":
		handleInventory(addPrefix, h.service, mux)
	case "orders service":
		handleOrders(addPrefix, h.service, mux)
	}
}

func handleInventory(addPrefix func(method, path string) string, svc *config.Service, mux *http.ServeMux) {
	mux.HandleFunc(addPrefix("POST", "/inventory"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("POST", "/inventory/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("GET", "/inventory"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("GET", "/inventory/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("GET", "/inventory/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("GET", "/inventory/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("PUT", "/inventory/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("PUT", "/inventory/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("DELETE", "/inventory/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("DELETE", "/inventory/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("POST", "/discounts"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("DELETE", "/discounts/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("DELETE", "/discounts/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("DELETE", "/discounts/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))
}

func handleOrders(addPrefix func(method, path string) string, svc *config.Service, mux *http.ServeMux) {
	mux.HandleFunc(addPrefix("POST", "/orders"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("POST", "/orders/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("GET", "/orders"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("GET", "/orders/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("GET", "/orders/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("GET", "/orders/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("PUT", "/orders/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("PUT", "/orders/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))

	mux.HandleFunc(addPrefix("DELETE", "/orders/{id}"), middleware.ReverseProxyHandler(svc, svc.URLBase))
	mux.HandleFunc(addPrefix("DELETE", "/orders/{id}/"), middleware.ReverseProxyHandler(svc, svc.URLBase))
}
