package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	requestLogFile  *os.File
	responseLogFile *os.File
	promptLogFile   *os.File
	costLogFile     *os.File
}

func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	requestLogFile, err := os.OpenFile(filepath.Join(logDir, "requests.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	responseLogFile, err := os.OpenFile(filepath.Join(logDir, "responses.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	promptLogFile, err := os.OpenFile(filepath.Join(logDir, "prompts.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	costLogFile, err := os.OpenFile(filepath.Join(logDir, "costs.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		requestLogFile:  requestLogFile,
		responseLogFile: responseLogFile,
		promptLogFile:   promptLogFile,
		costLogFile:     costLogFile,
	}, nil
}

func (l *Logger) LogRequest(method, path string, body map[string]interface{}) {
	log := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    method,
		"path":      path,
		"body":      body,
	}
	json.NewEncoder(l.requestLogFile).Encode(log)
}

func (l *Logger) LogResponse(statusCode int, body map[string]interface{}) {
	log := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"statusCode": statusCode,
		"body":       body,
	}
	json.NewEncoder(l.responseLogFile).Encode(log)
}

func (l *Logger) LogPrompt(prompt string) {
	log := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"prompt":    prompt,
	}
	json.NewEncoder(l.promptLogFile).Encode(log)
}

func (l *Logger) LogCost(model string, tokens int, cost float64) {
	log := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"model":     model,
		"tokens":    tokens,
		"cost":      cost,
	}
	json.NewEncoder(l.costLogFile).Encode(log)
}

func (l *Logger) Close() {
	l.requestLogFile.Close()
	l.responseLogFile.Close()
	l.promptLogFile.Close()
	l.costLogFile.Close()
}
