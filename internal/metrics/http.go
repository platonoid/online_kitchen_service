package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "kitchen_http_requests_total", Help: "Total HTTP requests handled by the kitchen API."}, []string{"method", "route", "status"})
	once     sync.Once
)

func Register() {
	once.Do(func() {
		prometheus.MustRegister(requests)
	})
}

func Middleware(next http.Handler) http.Handler {
	Register()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		requests.WithLabelValues(r.Method, r.URL.Path, http.StatusText(recorder.status)).Inc()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
