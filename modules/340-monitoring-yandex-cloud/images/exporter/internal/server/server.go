package server

import (
	"exporter/internal/yandex"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"exporter/internal/config"
)

type Server struct {
	router chi.Router
	logger *log.Logger

	config  config.Server
	metrics *yandex.Metrics
}

func New(config config.Server, logger *log.Logger, metrics *yandex.Metrics) *Server {
	return &Server{
		config:  config,
		logger:  logger,
		router:  chi.NewRouter(),
		metrics: metrics,
	}
}

func (h *Server) Run() {
	h.router.Route("/metrics/{service}", func(r chi.Router) {
		r.Get("/", h.getMetrics)
	})
}

func (h *Server) getMetrics(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")

	if !h.metrics.HasService(service) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Incorrect service"))
		if err != nil {
			h.logger.Errorf("Does not send response: %v", err)
		}
	}

	w.Write()

}

type metric struct {
	metrics []*string
	mu      sync.Mutex
}

func (h *Server) MetricsServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var response metric

	response.metrics = make([]*string, 0)
	wg.Add(len(h.req))

	for _, inst := range h.req {
		go func(inst data.InstanceRequest) {
			defer wg.Done()
			str, err := getMetrics(inst, r.Context())
			if err != nil {
				log.Printf("ERROR: Cannot get metrics for url '%s': %s\n", inst.Url, err)
				return
			}
			response.mu.Lock()
			response.metrics = append(response.metrics, &str)
			response.mu.Unlock()
		}(inst)
	}
	wg.Wait()

	response.mu.Lock()
	result := buildString(response.metrics)
	response.mu.Unlock()

	if result == "" {
		w.WriteHeader(http.StatusNotFound)
		_, sendErr := io.WriteString(w, "404 cannot get metrics")
		if sendErr != nil {
			log.Fatalf("ERROR: Internal error. Cannot send '404 Not found' response to client: %s", sendErr)
		}
	}
	_, err := io.WriteString(w, result)
	if err != nil {
		log.Fatalf("ERROR: Internal error. Cannot send '200 OK' response to client: %s", err)
	}
}

func buildString(resp []*string) string {
	var builder strings.Builder
	for _, s := range resp {
		builder.WriteString(*s)
	}
	return builder.String()
}
