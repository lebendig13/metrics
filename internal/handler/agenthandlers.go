package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

		url := baseURL + v.MType + "/" + v.ID + "/" + metricValue
		err := SendUpdateRequest(client, url)
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
