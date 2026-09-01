package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

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

type Server struct {
	memStorage *models.MemStorage
}

func NewServer(memStorage *models.MemStorage) *Server {
	return &Server{
		memStorage: memStorage,
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
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := req.URL.Path[1:]
	fmt.Println("Path: ", path)
	pathSegments := strings.Split(path, "/")
	pathSegmentsLen := len(pathSegments)

	if pathSegmentsLen < 2 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Println("Bad request: no metric type")
		return
	}

	mType := pathSegments[1]
	if mType != models.Counter && mType != models.Gauge {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Println("Bad request: bad type")
		return
	}
	fmt.Println("mType: ", mType)

	if pathSegmentsLen < 3 {
		res.WriteHeader(http.StatusNotFound)
		fmt.Println("Bad request: no metric type")
		return
	}

	mName := pathSegments[2]
	// TODO: Проверять имя метрики
	// ...
	if mName == "" {
		res.WriteHeader(http.StatusNotFound)
		fmt.Println("Bad request: empty metric name")
		return
	}
	fmt.Println("mName: ", mName)

	if pathSegmentsLen != 4 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Println("Bad request: pathSegmentsLen = ", pathSegmentsLen)
		return
	}

	mValue := pathSegments[3]
	if mValue == "" {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Println("Bad request: empty value")
		return
	}
	fmt.Println("mValue: ", mValue)

	var metric models.Metrics
	metric.ID = mName
	metric.MType = mType

	switch mType {
	case models.Counter:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Println("Bad request: cannot parse counter value")
			return
		}
		metric.Delta = &value
	case models.Gauge:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Println("Bad request: cannot parse gauge value")
			return
		}
		metric.Value = &value
	}

	err := s.memStorage.InsertOrUpdate(metric)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Internal server error: cannot save metric to storage")
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func (s *Server) GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	var tmpl = template.Must(template.New("metrics").Parse(allMetricsPage))
	allMetrics := s.memStorage.GetAllMetrics()
	if err := tmpl.Execute(res, allMetrics); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) GetMetricHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metric_type")
	if metricType != models.Counter && metricType != models.Gauge {
		fmt.Println("Unknown metric type: ", metricType)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	metricName := chi.URLParam(req, "metric_name")
	metric, exists := s.memStorage.Get(metricName)
	if !exists {
		fmt.Printf("Value of %s not found\r\n", metricName)
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if metric.MType != metricType {
		fmt.Printf("Metric %s has type %s\r\n", metricName, metric.MType)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	result := ""
	if metric.MType == models.Counter {
		result = strconv.FormatInt(*metric.Delta, 10)
	}
	if metric.MType == models.Gauge {
		result = strconv.FormatFloat(*metric.Value, 'f', 10, 64)
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, result)
}
