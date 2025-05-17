package logger

import (
	"log"
	"net/http"
	"os"
	"time"
)



// statusResponseWriter menangkap status code dari response
type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func init() {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
}

// RequestLogger mencatat log lengkap per request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(srw, r)

		duration := time.Since(start)
		logLine := time.Now().Format(time.RFC3339) + " - " +
			r.Method + " " + r.URL.Path + " - " +
			r.Proto + " - " +
			"Status: " + http.StatusText(srw.status) + " - " +
			"IP: " + r.RemoteAddr + " - " +
			"Duration: " + duration.String()

		// Tulis ke stdout dan ke file log
		log.Println(logLine)
		appendToFile("app.log", logLine)
	})
}

// appendToFile menulis log ke file
func appendToFile(filename string, line string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[ERROR] Failed to write log file: %v", err)
		return
	}
	defer f.Close()

	f.WriteString(line + "\n")
}

// LogInfo mencatat log level info
func LogInfo(message string) {
	logLine := "[INFO] " + time.Now().Format(time.RFC3339) + " - " + message
	log.Println(logLine)
	appendToFile("app.log", logLine)
}

// LogWarning mencatat log level warning
func LogWarning(message string) {
	logLine := "[WARN] " + time.Now().Format(time.RFC3339) + " - " + message
	log.Println(logLine)
	appendToFile("app.log", logLine)
}

// LogError mencatat log level error
func LogError(message string) {
	logLine := "[ERROR] " + time.Now().Format(time.RFC3339) + " - " + message
	log.Println(logLine)
	appendToFile("app.log", logLine)
}
