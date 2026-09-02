package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	models "github.com/lebendig13/metrics/internal/model"
)

func TestSendMetrics(t *testing.T) {
	// Присланные запросы
	receivedUrls := make([]string, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPost)

		contentType := r.Header.Get("Content-Type")
		assert.Equal(t, contentType, "text/plain")

		receivedUrls = append(receivedUrls, r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))

	client := &http.Client{}

	// Тестовые метрики
	gaugeMetricValue := 0.1
	gaugeMetric := models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &gaugeMetricValue}
	counterMetricValue := int64(1)
	counterMetric := models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &counterMetricValue}
	unknownMetric := models.Metrics{ID: "InvalidMetric", MType: "unknown", Delta: &counterMetricValue}

	tests := []struct {
		name         string
		metricsValue []*models.Metrics
		want         []string
	}{
		{
			name:         "counter metric",
			metricsValue: []*models.Metrics{&counterMetric},
			want:         []string{"/update/counter/PollCount/1"},
		},
		{
			name:         "gauge metric",
			metricsValue: []*models.Metrics{&gaugeMetric},
			want:         []string{"/update/gauge/Alloc/0.1000000000"},
		},
		{
			name:         "unknown metric",
			metricsValue: []*models.Metrics{&unknownMetric},
			want:         []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receivedUrls = []string{}
			SendMetrics(client, test.metricsValue, server.URL+"/update/")
			assert.Equal(t, receivedUrls, test.want)
		})
	}

	defer server.Close()
}

func TestSendUpdateRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	client := &http.Client{}

	tests := []struct {
		name     string
		urlValue string
		want     string
	}{
		{
			name:     "invalid url with space",
			urlValue: "http://local host:8080/update/gauge/testname/1",
			want:     "cannot create request",
		},
		{
			name:     "invalid url with wrong address",
			urlValue: "http://testlocalhost:8080/update/gauge/testname/1",
			want:     "cannot send request",
		},
		{
			name:     "error status",
			urlValue: server.URL + "/update/gauge/testname/1",
			want:     "got status 404",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SendUpdateRequest(client, test.urlValue)
			fmt.Println("Actual error: ", err.Error())
			assert.Contains(t, err.Error(), test.want)
		})
	}

	defer server.Close()
}
