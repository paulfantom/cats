package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Cat struct {
	Id         string            `json:"id,omitempty"`
	Url        string            `json:"url"`
	Width      int               `json:"width,omitempty"`
	Height     int               `json:"height,omitempty"`
	Breeds     []string          `json:"breeds,omitempty"`
	Faviourite map[string]string `json:"favourite,omitempty"`
}

func getCat(apiKey string) (string, error) {
	url := "https://api.thecatapi.com/v1/images/search"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("x-api-key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var cats []Cat
	json.Unmarshal(body, &cats)

	return cats[0].Url, nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create a non-global registry.
	reg := prometheus.NewRegistry()
	requestsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		},
		[]string{"method", "code"},
	)

	// Add Go module build info.
	reg.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// Reset metrics to 0's. This allows for better alerting in a low-traffic environment.
	requestsTotal.WithLabelValues("GET", "2xx")
	requestDuration.WithLabelValues("GET", "2xx")

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestDuration.WithLabelValues(r.Method, "2xx").Observe(time.Since(start).Seconds())
		log.Info().Msgf("Received request from %s", r.RemoteAddr)
		apiKey := r.Header.Get("x-api-key")
		cat, err := getCat(apiKey)
		if err != nil {
			requestsTotal.WithLabelValues(r.Method, "5xx").Inc()
			requestDuration.WithLabelValues(r.Method, "5xx").Observe(time.Since(start).Seconds())
			log.Error().Err(err).Msg("Error getting a cat from the external API")
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		requestsTotal.WithLabelValues(r.Method, "2xx").Inc()
		http.Redirect(w, r, cat, http.StatusFound)
	})

	log.Info().Msg("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
