package handler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

		if strings.Contains(r.URL.Path, "/update/") {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
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
		url          string
		want         []string
		err          string
	}{
		{
			name:         "counter metric",
			metricsValue: []*models.Metrics{&counterMetric},
			url:          server.URL + "/update/",
			want:         []string{"/update/counter/PollCount/1"},
		},
		{
			name:         "gauge metric",
			metricsValue: []*models.Metrics{&gaugeMetric},
			url:          server.URL + "/update/",
			want:         []string{"/update/gauge/Alloc/0.1000000000"},
		},
		{
			name:         "unknown metric",
			metricsValue: []*models.Metrics{&unknownMetric},
			url:          server.URL + "/update/",
			want:         []string{},
			err:          "couldn't update metrics",
		},
		{
			name:         "pollCount update error",
			metricsValue: []*models.Metrics{&counterMetric},
			url:          server.URL + "/unknownreq/",
			want:         []string{"/unknownreq/counter/PollCount/1"},
			err:          "couldn't update PollCount",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receivedUrls = []string{}
			err := SendMetrics(client, test.metricsValue, test.url)
			assert.Equal(t, test.want, receivedUrls)
			if err != nil {
				assert.Equal(t, test.err, err.Error())
			}
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
			log.Println("Actual error: ", err.Error())
			assert.Contains(t, err.Error(), test.want)
		})
	}

	defer server.Close()
}
