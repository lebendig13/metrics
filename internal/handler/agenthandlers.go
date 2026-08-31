package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	models "github.com/lebendig13/metrics/internal/model"
)

func SendMetrics(client *http.Client, m []*models.Metrics, baseURL string) {
	for _, v := range m {
		metricValue := ""
		if v.MType == models.Counter {
			metricValue = strconv.FormatInt(*v.Delta, 10)
		}
		if v.MType == models.Gauge {
			metricValue = strconv.FormatFloat(*v.Value, 'f', 10, 64)
		}
		fmt.Printf("metricValue %s = %s\r\n", v.ID, metricValue)
		if metricValue == "" {
			fmt.Println("Cannot get metricValue")
			return
		}

		url := baseURL + v.MType + "/" + v.ID + "/" + metricValue
		SendUpdateRequest(client, url)
	}
}

func SendUpdateRequest(client *http.Client, url string) error {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("Cannot create request with url: ", url)
		return errors.New(fmt.Sprintf("Cannot create request: %s", err.Error()))
	}

	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Cannot send request with url: ", url)
		return errors.New(fmt.Sprintf("Cannot send request: %s", err.Error()))
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Got status %v for url %s\r\n", response.StatusCode, url)
		return errors.New(fmt.Sprintf("Got status %v", response.StatusCode))
	}
	return nil
}
