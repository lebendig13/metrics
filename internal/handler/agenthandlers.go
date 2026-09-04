package handler

import (
	"fmt"
	"log"
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
		log.Printf("metricValue %s = %s\r\n", v.ID, metricValue)
		if metricValue == "" {
			log.Println("Cannot get metricValue")
			continue
		}

		url := baseURL + v.MType + "/" + v.ID + "/" + metricValue
		SendUpdateRequest(client, url)
	}
}

func SendUpdateRequest(client *http.Client, url string) error {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println("Cannot create request with url: ", url)
		return fmt.Errorf("cannot create request: %s", err.Error())
	}

	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		log.Println("Cannot send request with url: ", url)
		return fmt.Errorf("cannot send request: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("Got status %v for url %s\r\n", response.StatusCode, url)
		return fmt.Errorf("got status %v", response.StatusCode)
	}
	return nil
}
