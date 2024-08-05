package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

type PostLogInput struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	TraceId   string                 `json:"traceId"`
	Timestamp time.Time              `json:"timestamp"`
}

type CustomResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()
		res := c.Response()

		var requestBody []byte
		if req.Body != nil {
			requestBody, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		logRequest(req, requestBody, start)

		resBody := new(bytes.Buffer)
		crw := &CustomResponseWriter{ResponseWriter: res.Writer, body: resBody}
		res.Writer = crw

		err := next(c)

		end := time.Now()

		responseBody := crw.body.Bytes()
		logResponse(res, responseBody, start, end)

		return err
	}
}

func logRequest(req *http.Request, requestBody []byte, start time.Time) {
	logData := map[string]interface{}{
		"method":         req.Method,
		"url":            req.URL.String(),
		"requestHeaders": req.Header,
		"requestBody":    string(requestBody),
		"startTime":      start,
	}

	logInput := PostLogInput{
		Type:      "API",
		Data:      logData,
		TraceId:   "12345",
		Timestamp: time.Now(),
	}

	postLog(logInput)
}

func logResponse(res *echo.Response, responseBody []byte, start, end time.Time) {
	logData := map[string]interface{}{
		"responseStatus": res.Status,
		"responseBody":   string(responseBody),
		"startTime":      start,
		"endTime":        end,
		"duration":       end.Sub(start).String(),
	}

	logInput := PostLogInput{
		Type:      "API",
		Data:      logData,
		TraceId:   "12345",
		Timestamp: time.Now(),
	}

	postLog(logInput)
}

func postLog(logInput PostLogInput) {
	logInputJSON, err := json.Marshal(logInput)
	if err != nil {
		log.Printf("Error marshaling log data: %v", err)
		return
	}

	logServiceURL := "http://localhost:" + os.Getenv("LOGGER_PORT") + "/log"
	go func() {
		resp, err := http.Post(logServiceURL, "application/json", bytes.NewBuffer(logInputJSON))
		if err != nil {
			log.Printf("Error posting log data: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Unexpected response from log service: %s", body)
		}
	}()
}
