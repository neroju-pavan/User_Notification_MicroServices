package logger

import (
	"encoding/json"
	"os"
	"time"
)

// LogEntry defines a standardized log format
type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Method    string                 `json:"method"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Global service name (set from main.go for each microservice)
var ServiceName = "user-service"

func SetServiceName(name string) {
	ServiceName = name
}

// core logger function
func log(level, method, msg string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   ServiceName,
		Method:    method,
		Message:   msg,
		Fields:    fields,
	}

	data, _ := json.Marshal(entry)
	os.Stdout.Write(append(data, '\n'))
}

// =====================
// Public Logger Methods
// =====================

func Info(method, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields...)
	log("INFO", method, msg, f)
}

func Error(method, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields...)
	log("ERROR", method, msg, f)
}

func Warn(method, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields...)
	log("WARN", method, msg, f)
}

func Debug(method, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields...)
	log("DEBUG", method, msg, f)
}

// merge optional fields
func mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := map[string]interface{}{}
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}
