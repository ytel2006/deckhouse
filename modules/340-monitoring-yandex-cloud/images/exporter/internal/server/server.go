package server

import (
	"context"
	"exporter/internal/yandex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	api             *yandex.CloudApi
	servicesForBach []string
	logger          *log.Entry

	router chi.Router
}

func New(logger *log.Entry, api *yandex.CloudApi, servicesForBach []string) *Server {
	return &Server{
		logger:          logger,
		router:          chi.NewRouter(),
		api:             api,
		servicesForBach: servicesForBach,
	}
}

func (h *Server) Run(listenAddr string, stopCh chan struct{}) error {
	h.router.Route("/metrics/{service}", func(r chi.Router) {
		r.Get("/", h.getByService)
	})

	h.router.Get("/metrics", h.getByService)

	srv := http.Server{
		Addr:         listenAddr,
		Handler:      h.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  1 * time.Minute,
	}

	srv.RegisterOnShutdown(func() {
		close(stopCh)
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		h.logger.Infof("Signal received: %v. Exiting...", <-signalChan)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			h.logger.Fatalf("Error occurred while closing the server: %v", err)
		}
		os.Exit(0)
	}()

	h.logger.Infof("Start listening on %q", listenAddr)

	return srv.ListenAndServe()
}

func (h *Server) writeError(upErr error, w http.ResponseWriter) {
	h.logger.Errorf("cannot scrape metrics: %v", upErr)
	w.WriteHeader(http.StatusInternalServerError)
	response := "Cannot scrape metrics. see server logs for describe error"

	if _, err := w.Write([]byte(response)); err != nil {
		h.logger.Errorf("cannot write response: %v", err)
	}
}

func (h *Server) getByService(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")

	if !h.api.HasService(service) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Incorrect service"))
		if err != nil {
			h.logger.Errorf("Does not send response: %v", err)
		}
	}

	metrics, err := h.api.RequestMetrics(r.Context(), service)
	if err != nil {
		h.writeError(err, w)
		return
	}

	_, err = w.Write(metrics)
	if err != nil {
		h.logger.Errorf("cannot write response: %v", err)
	}
}

func (h *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	servicesLen := len(h.servicesForBach)

	if servicesLen == 0 {
		h.writeError(fmt.Errorf("cannot pas services for scrape"), w)
		return
	}

	resultsCh := make(chan []byte, servicesLen)

	for _, s := range h.servicesForBach {
		go func(service string) {
			metrics, err := h.api.RequestMetrics(r.Context(), service)
			if err != nil {
				h.logger.Errorf("ERROR: Cannot get metrics for service %s: %v\n", service, err)
				resultsCh <- nil
				return
			}

			resultsCh <- metrics
		}(s)
	}

	for i := 0; i < servicesLen; i++ {
		res := <-resultsCh
		if res != nil {
			_, err := w.Write(res)
			if err != nil {
				h.logger.Errorf("cannot write metrics %v", err)
			}
		}
	}
	_, err = w.Write(metrics)
	if err != nil {
		h.logger.Errorf("cannot write response: %v", err)
	}
}
