package handler

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	models "github.com/lebendig13/metrics/internal/model"
)

const (
	allMetricsPage = `<!DOCTYPE html>
	<html>
	<head><meta charset="utf-8"><title>Metrics</title></head>
	<body>
		<ul>
		{{range $name, $m := .}}
			<li>{{$name}}: {{$m}}</li>
		{{end}}
		</ul>
	</body>
	</html>`
)

type Storage interface {
	InsertOrUpdate(m models.Metrics) error
	Get(id string) (models.Metrics, bool)
	GetAllMetrics() map[string]string
}

type Server struct {
	storage Storage
}

func NewServer(stg Storage) *Server {
	return &Server{
		storage: stg,
	}
}

func MetricsRouter(server *Server) chi.Router {
	router := chi.NewRouter()
	router.Post("/update/{metric_type}/{metric_name}/{metric_value}", server.UpdateMetricHandler)
	router.Get("/", server.GetAllMetricsHandler)
	router.Get("/value/{metric_type}/{metric_name}", server.GetMetricHandler)

	return router
}

func (s *Server) UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	mType := chi.URLParam(req, "metric_type")
	if mType != models.Counter && mType != models.Gauge {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("mType: ", mType)

	mName := chi.URLParam(req, "metric_name")
	if mName == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	log.Println("mName: ", mName)

	mValue := chi.URLParam(req, "metric_value")
	if mValue == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("mValue: ", mValue)

	var metric models.Metrics
	metric.ID = mName
	metric.MType = mType

	switch mType {
	case models.Counter:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Delta = &value
	case models.Gauge:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Value = &value
	}

	err := s.storage.InsertOrUpdate(metric)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("Internal server error: cannot save metric to storage")
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func (s *Server) GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	var tmpl = template.Must(template.New("metrics").Parse(allMetricsPage))
	allMetrics := s.storage.GetAllMetrics()
	if err := tmpl.Execute(res, allMetrics); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("Internal server error: cannot generate HTML with all metrics")
	}
}

func (s *Server) GetMetricHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metric_type")
	if metricType != models.Counter && metricType != models.Gauge {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	metricName := chi.URLParam(req, "metric_name")
	metric, exists := s.storage.Get(metricName)
	if !exists {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if metric.MType != metricType {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	result := ""
	if metric.MType == models.Counter {
		result = strconv.FormatInt(*metric.Delta, 10)
	}
	if metric.MType == models.Gauge {
		result = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, result)
}
