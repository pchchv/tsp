package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Check struct {
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ExpectedCode int    `yaml:"expected_code"`
}

type HistoryEntry struct {
	Timestamp string `json:"timestamp"`
	Status    bool   `json:"status"`
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		_, _ = fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return fallback
}

func checkHTTP(url string, expectedCode int) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	_ = resp.Body.Close()
	return resp.StatusCode == expectedCode
}

func checkPing(host string) bool {
	cmd := exec.Command("ping", "-c", "1", "-W", "2", host)
	return cmd.Run() == nil
}

func checkPort(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return false
	}

	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	return true
}

func runChecks(checks []Check) (results []map[string]interface{}) {
	resultsCh := make(chan map[string]interface{}, len(checks))
	for _, check := range checks {
		go func(c Check) {
			var status bool
			switch c.Type {
			case "http":
				status = checkHTTP(c.Host, c.ExpectedCode)
			case "ping":
				status = checkPing(c.Host)
			case "port":
				status = checkPort(c.Host, c.Port)
			}
			resultsCh <- map[string]interface{}{"name": c.Name, "status": status}
		}(check)
	}

	for i := 0; i < len(checks); i++ {
		results = append(results, <-resultsCh)
	}

	return
}
