package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/lebendig13/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
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
				code: http.StatusBadRequest,
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
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "empty value",
			method: http.MethodPost,
			path:   "/update/counter/PollCount/",
			want: want{
				code: http.StatusBadRequest,
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

			memStorage := models.NewMemStorage()
			server := NewServer(memStorage)
			server.UpdateMetricHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			if test.want.code == http.StatusOK {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
