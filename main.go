package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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

const (
	indexfile   = "index.html"
	historyfile = "history.html"
)

var (
	checkInterval     = getEnvInt("CHECK_INTERVAL", 60)
	maxHistoryEntries = getEnvInt("MAX_HISTORY_ENTRIES", 10)
	checksFile        = getEnv("CHECKS_FILE", "checks.yaml")
	incidentsFile     = getEnv("INCIDENTS_FILE", "incidents.html")
	historyFile       = getEnv("STATUS_HISTORY_FILE", "history.json")
	port              = getEnv("PORT", "")
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

func generateHistoryPage() {
	history := loadHistory()
	tmpl, err := template.New("history").Funcs(template.FuncMap{
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
	}).Parse(historyTemplateFile)
	if err != nil {
		log.Fatal("Failed to parse history template:", err)
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"history":      history,
		"last_updated": time.Now().Format("2006-01-02 15:04:05"),
	}
	if err = tmpl.Execute(&buf, data); err != nil {
		log.Fatal("Failed to execute history template:", err)
	}

	if err = os.WriteFile(historyfile, buf.Bytes(), 0644); err != nil {
		log.Fatal("Failed to write history page:", err)
	}
}

func renderTemplate(data map[string]interface{}) string {
	funcMap := template.FuncMap{
		"check": func(status bool) string {
			if status {
				return "Up"
			}
			return "Down"
		},
	}
	tmpl, err := template.New("status").Funcs(funcMap).Parse(templateFile)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	return buf.String()
}

func serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed!", http.StatusMethodNotAllowed)
		return
	}

	_, _ = fmt.Fprintf(w, templateStatus, runtime.Version(), runtime.NumGoroutine())
}

func monitorServices() {
	for {
		checksData, err := os.ReadFile(checksFile)
		if err != nil {
			log.Fatal("Failed to load checks file:", err)
		}

		var checks []Check
		if err = yaml.Unmarshal(checksData, &checks); err != nil {
			log.Fatal("Failed to parse checks file:", err)
		}

		results := runChecks(checks)
		updateHistory(results)
		incidentMarkdown, err := os.ReadFile(incidentsFile)
		if err != nil {
			log.Println("Failed to load incidents:", err)
			incidentMarkdown = []byte("<h2>All Fine!</h2>")
		}

		data := map[string]interface{}{
			"checks":       results,
			"incidents":    template.HTML(incidentMarkdown),
			"last_updated": time.Now().Format("2006-01-02 15:04:05"),
		}
		html := renderTemplate(data)
		if err = os.WriteFile(indexfile, []byte(html), 0644); err != nil {
			log.Fatal("Failed to write index.html:", err)
		}

		generateHistoryPage()
		log.Println("Status pages updated!")
		time.Sleep(time.Duration(checkInterval) * time.Second)
	}
}

func main() {
	log.Println("Monitoring services ...")
	if port != "" {
		go monitorServices()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				serveFile(w, r, "./"+indexfile)
			} else {
				http.NotFound(w, r)
			}
		})

		http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/status") {
				handleHome(w, r)
			} else {
				http.NotFound(w, r)
			}
		})

		http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/history") {
				serveFile(w, r, "./"+historyfile)
			} else {
				http.NotFound(w, r)
			}
		})

		log.Println("Server started!")

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Println(err.Error())
		}
	} else {
		monitorServices()
	}
	log.Println("Bye!")
}
