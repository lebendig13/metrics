package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/lebendig13/metrics/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandler(t *testing.T) {
	memStorage := models.NewMemStorage()
	server := NewServer(memStorage)
	router := MetricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		path   string
		want   want
	}{
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/update/counter/PollCount/1",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodGet,
			path:   "/update/counter/PollCount/1",
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "no metric",
			method: http.MethodPost,
			path:   "/update",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "bad type",
			method: http.MethodPost,
			path:   "/update/unknown_type/PollCount/1",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "no metric type",
			method: http.MethodPost,
			path:   "/update/counter",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "empty metric name",
			method: http.MethodPost,
			path:   "/update/counter/",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "invalid path",
			method: http.MethodPost,
			path:   "/update/counter/PollCount/1/test/1/1",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "empty value",
			method: http.MethodPost,
			path:   "/update/counter/PollCount/",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "bad counter value",
			method: http.MethodPost,
			path:   "/update/counter/PollCount/test",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "bad gauge value",
			method: http.MethodPost,
			path:   "/update/gauge/Alloc/test",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			if test.want.code == http.StatusOK {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	memStorage := models.NewMemStorage()
	dvalue := int64(1)
	vvalue := 0.1
	memStorage.InsertOrUpdate(models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &dvalue})
	memStorage.InsertOrUpdate(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &vvalue})
	server := NewServer(memStorage)
	ts := httptest.NewServer(MetricsRouter(server))
	defer ts.Close()

	type want struct {
		code        int
		contentType string
		body        string
	}
	tests := []struct {
		name   string
		method string
		path   string
		want   want
	}{
		{
			name:   "success",
			method: http.MethodGet,
			path:   "/value/counter/PollCount",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        "1",
			},
		},
		{
			name:   "unknown metric type",
			method: http.MethodGet,
			path:   "/value/counter_test/PollCount",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "unknown metric name",
			method: http.MethodGet,
			path:   "/value/counter/PollCount_test",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "incorrect metric type",
			method: http.MethodGet,
			path:   "/value/gauge/PollCount",
			want: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, test := range tests {
		request, err := http.NewRequest(test.method, ts.URL+test.path, nil)
		require.NoError(t, err)

		resp, err := ts.Client().Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, test.want.code, resp.StatusCode)

		if test.want.code == http.StatusOK {
			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, test.want.body, string(resBody))
		}
	}
}

func TestGetAllMetricsHandler(t *testing.T) {
	memStorage := models.NewMemStorage()
	dvalue := int64(1)
	vvalue := 0.1
	memStorage.InsertOrUpdate(models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &dvalue})
	memStorage.InsertOrUpdate(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &vvalue})
	server := NewServer(memStorage)
	ts := httptest.NewServer(MetricsRouter(server))
	defer ts.Close()

	request, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(request)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	wantBody := `<!DOCTYPE html>
	<html>
	<head><meta charset="utf-8"><title>Metrics</title></head>
	<body>
		<ul>
		
			<li>Alloc: 0.1000000000</li>
		
			<li>PollCount: 1</li>
		
		</ul>
	</body>
	</html>`
	assert.Equal(t, wantBody, string(resBody))
}
