package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
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

var (
	maxHistoryEntries = getEnvInt("MAX_HISTORY_ENTRIES", 10)
	historyFile = getEnv("STATUS_HISTORY_FILE", "history.json")
)

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

func loadHistory() (history map[string][]HistoryEntry) {
	file, err := os.Open(historyFile)
	if err != nil {
		return map[string][]HistoryEntry{}
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_ = json.NewDecoder(file).Decode(&history)
	if history == nil {
		history = make(map[string][]HistoryEntry)
	}

	return
}

func saveHistory(history map[string][]HistoryEntry) {
	file, err := os.Create(historyFile)
	if err != nil {
		log.Println("Failed to save history:", err)
		return
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_ = json.NewEncoder(file).Encode(history)
}

func updateHistory(results []map[string]interface{}) {
	history := loadHistory()
	currentTime := time.Now().Format(time.RFC3339)
	for _, result := range results {
		name := result["name"].(string)
		if _, exists := history[name]; !exists {
			history[name] = []HistoryEntry{}
		}

		history[name] = append(history[name], HistoryEntry{currentTime, result["status"].(bool)})
		sort.Slice(history[name], func(i, j int) bool {
			timeI, _ := time.Parse(time.RFC3339, history[name][i].Timestamp)
			timeJ, _ := time.Parse(time.RFC3339, history[name][j].Timestamp)
			return timeI.After(timeJ)
		})

		if len(history[name]) > maxHistoryEntries {
			history[name] = history[name][:maxHistoryEntries]
		}
	}
	saveHistory(history)
}
