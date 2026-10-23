package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	models "github.com/lebendig13/metrics/internal/model"
)

func SendMetrics(client *http.Client, m []*models.Metrics, baseURL string) error {
	successRequestCounter := len(m) // оптимистично предполагаем, что все метрики будут успешно отправлены
	for _, v := range m {
		metricValue := ""
		if v.MType == models.Counter {
			metricValue = strconv.FormatInt(*v.Delta, 10)
		}
		if v.MType == models.Gauge {
			metricValue = strconv.FormatFloat(*v.Value, 'f', 10, 64)
		}
		if metricValue == "" {
			log.Println("Cannot get metricValue: ", v.ID)
			successRequestCounter--
			continue
		}
		log.Printf("metricValue %s = %s\r\n", v.ID, metricValue)

		body, err := json.Marshal(v)
		if err != nil {
			log.Printf("Cannot marshal metric %s = %s: %v\r\n", v.ID, metricValue, err)
			continue
		}
		err = SendUpdateWithJSONRequest(client, baseURL, bytes.NewReader(body))
		if err != nil {
			log.Println(err)
			successRequestCounter--

			if v.ID == "PollCount" {
				return fmt.Errorf("couldn't update PollCount") // в этом случае не сбрасываем agent.PollCount
			}
		}
	}
	if successRequestCounter == 0 {
		return fmt.Errorf("couldn't update metrics")
	}
	return nil
}

func SendUpdateRequest(client *http.Client, url string) error {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w. URL: %s", err, url)
	}

	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("cannot send request: %w. URL: %s", err, url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %v for URL: %s", response.StatusCode, url)
	}
	return nil
}

func SendUpdateWithJSONRequest(client *http.Client, url string, body io.Reader) error {
	compressedBody, err := CompressData(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, &compressedBody)
	if err != nil {
		return fmt.Errorf("cannot create request: %w. URL: %s", err, url)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("cannot send request: %w. URL: %s", err, url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %v for URL: %s", response.StatusCode, url)
	}
	return nil
}

func CompressData(data io.Reader) (bytes.Buffer, error) {
	var compressedData bytes.Buffer
	zw := gzip.NewWriter(&compressedData)

	if _, err := io.Copy(zw, data); err != nil {
		return compressedData, fmt.Errorf("cannot compress data: %w", err)
	}

	if err := zw.Close(); err != nil {
		return compressedData, fmt.Errorf("cannot close gzip writer: %w", err)
	}

	return compressedData, nil
}

func SendMetricsBatch(ctx context.Context, client *http.Client, m []*models.Metrics, baseURL string) error {
	body, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("cannot marshal metrics: %v", err)
	}

	compressedBody, err := CompressData(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot compress body: %v", err)
	}
	rawBytes := compressedBody.Bytes()

	attemptIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	var lastErr error

	for i := 0; i <= len(attemptIntervals); i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled before/during retries: %w", err)
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(rawBytes))
		if err != nil {
			return fmt.Errorf("cannot create request: %w. URL: %s", err, baseURL)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Accept-Encoding", "gzip")

		response, err := client.Do(request)
		if err == nil {
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("got status %v for URL: %s", response.StatusCode, baseURL)
			}
			return nil
		}

		lastErr = err

		if i == len(attemptIntervals) || !isNetworkError(err) {
			break
		}

		log.Printf("Failed send metrics batch: network error. Attempt %d. Wait %v. Error: %v", i+1, attemptIntervals[i], err)
		select {
		case <-time.After(attemptIntervals[i]):
		case <-ctx.Done():
			return fmt.Errorf("context timed out during retry delay: %w", ctx.Err())
		}
	}

	return fmt.Errorf("cannot send request after retries: %w. URL: %s", lastErr, baseURL)
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	return false
}
