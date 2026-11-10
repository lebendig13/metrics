package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type ServerConfig struct {
	RunAddress      string
	LogLevel        string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	HashKey         string
}

type AgentConfigIntervals struct {
	ReportInterval int
	PollInterval   int
}

type AgentConfig struct {
	ServerAddress string
	Intervals     AgentConfigIntervals
	HashKey       string
	RateLimit     int
}

func GetServerConfig() ServerConfig {
	var result ServerConfig
	flag.StringVar(&result.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&result.LogLevel, "l", "info", "log level")
	flag.IntVar(&result.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&result.FileStoragePath, "f", "/home/user/metrics/metrics_values.txt", "file storage path")
	flag.BoolVar(&result.Restore, "r", false, "restore saved metrics when server starts")
	flag.StringVar(&result.DatabaseDSN, "d", fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", `localhost`, `metricsuser`, `metricsuser`, `metrics`), "database connection string")
	flag.StringVar(&result.HashKey, "k", "", "hash key")
	flag.Parse()

	runAddrEnv, runAddrEnvExists := os.LookupEnv("ADDRESS")
	if runAddrEnvExists && runAddrEnv != "" {
		result.RunAddress = runAddrEnv
	}

	logLvlEnv, logLvlEnvExists := os.LookupEnv("LOG_LEVEL")
	if logLvlEnvExists && logLvlEnv != "" {
		result.LogLevel = logLvlEnv
	}

	siEnv, siEnvExists := os.LookupEnv("STORE_INTERVAL")
	if siEnvExists && siEnv != "" {
		siEnvValue, err := strconv.Atoi(siEnv)
		if err != nil {
			log.Fatalf("Invalid STORE_INTERVAL value '%v': %v", siEnv, err)
		}
		result.StoreInterval = siEnvValue
	}

	fspathEnv, fspathEnvExists := os.LookupEnv("FILE_STORAGE_PATH")
	if fspathEnvExists && fspathEnv != "" {
		result.FileStoragePath = fspathEnv
	}

	rEnv, rEnvExists := os.LookupEnv("RESTORE")
	if rEnvExists && rEnv != "" {
		rEnvValue, err := strconv.ParseBool(rEnv)
		if err != nil {
			log.Fatalf("Invalid RESTORE value '%v': %v", rEnv, err)
		}
		result.Restore = rEnvValue
	}

	dsnEnv, dsnEnvExists := os.LookupEnv("DATABASE_DSN")
	if dsnEnvExists && dsnEnv != "" {
		result.DatabaseDSN = dsnEnv
	}

	hkEnv, hkEnvExists := os.LookupEnv("KEY")
	if hkEnvExists && hkEnv != "" {
		result.HashKey = hkEnv
	}

	return result
}

func GetAgentConfig() AgentConfig {
	var result AgentConfig

	flag.StringVar(&result.ServerAddress, "a", "localhost:8080", "address and port of the server")
	flag.IntVar(&result.Intervals.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&result.Intervals.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&result.HashKey, "k", "", "hash key")
	flag.IntVar(&result.RateLimit, "l", 1, "rate limit")
	flag.Parse()

	serverAddrEnv, serverAddrEnvExists := os.LookupEnv("ADDRESS")
	if serverAddrEnvExists && serverAddrEnv != "" {
		result.ServerAddress = serverAddrEnv
	}

	riEnv, riEnvExists := os.LookupEnv("REPORT_INTERVAL")
	if riEnvExists && riEnv != "" {
		riEnvValue, err := strconv.Atoi(riEnv)
		if err != nil {
			log.Fatalf("Invalid REPORT_INTERVAL value '%v': %v", riEnv, err)
		}
		result.Intervals.ReportInterval = riEnvValue
	}

	piEnv, piEnvExists := os.LookupEnv("POLL_INTERVAL")
	if piEnvExists && piEnv != "" {
		piEnvValue, err := strconv.Atoi(piEnv)
		if err != nil {
			log.Fatalf("Invalid POLL_INTERVAL value '%v': %v", piEnv, err)
		}
		result.Intervals.PollInterval = piEnvValue
	}

	hkEnv, hkEnvExists := os.LookupEnv("KEY")
	if hkEnvExists && hkEnv != "" {
		result.HashKey = hkEnv
	}

	rlEnv, rlEnvExists := os.LookupEnv("RATE_LIMIT")
	if rlEnvExists && rlEnv != "" {
		rlEnvValue, err := strconv.Atoi(rlEnv)
		if err != nil {
			log.Fatalf("Invalid RATE_LIMIT value '%v': %v", rlEnv, err)
		}
		result.RateLimit = rlEnvValue
	}
	if result.RateLimit < 1 {
		log.Printf("Rate limit must be positive, got %d; using 1", result.RateLimit)
		result.RateLimit = 1
	}

	return result
}
