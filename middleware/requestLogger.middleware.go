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
	logTypeEnum "github.com/Neeraj-Neurofin/requests-logger/store/enum"
	"github.com/Neeraj-Neurofin/requests-logger/store/types"
	"github.com/google/uuid"
)

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

		// Generate a traceID for the entire request-response cycle
		traceID, ok := c.Get("traceID").(string)
		if !ok {
			traceID := uuid.New().String() // Generate a new UUID for the traceID
			c.Set("traceID", traceID)
		}

		var requestBody []byte
		if req.Body != nil {
			requestBody, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		logRequest(req, requestBody, start, traceID)

		resBody := new(bytes.Buffer)
		crw := &CustomResponseWriter{ResponseWriter: res.Writer, body: resBody}
		res.Writer = crw

		err := next(c)

		end := time.Now()

		responseBody := crw.body.Bytes()
		logResponse(res, responseBody, crw.Header(), start, end, traceID)

		return err
	}
}

func logRequest(req *http.Request, requestBody []byte, start time.Time, traceID string) {
	logData := map[string]interface{}{
		"method":         req.Method,
		"url":            req.URL.String(),
		"requestHeaders": req.Header,
		"requestBody":    string(requestBody),
		"startTime":      start,
	}

	logInput := loggerTypes.PostLogInput{
		Type:      logTypeEnum.API,
		Data:      logData,
		TraceId:   traceID,
		Timestamp: time.Now(),
	}

	postLog(logInput)
}

func logResponse(res *echo.Response, responseBody []byte, responseHeaders http.Header, start, end time.Time, traceID string) {
	logData := map[string]interface{}{
		"responseStatus":  res.Status,
		"responseHeaders": responseHeaders,
		"responseBody":    string(responseBody),
		"startTime":       start,
		"endTime":         end,
		"duration":        end.Sub(start).String(),
	}

	logInput := loggerTypes.PostLogInput{
		Type:      logTypeEnum.API,
		Data:      logData,
		TraceId:   traceID,
		Timestamp: time.Now(),
	}

	postLog(logInput)
}

func postLog(logInput loggerTypes.PostLogInput) {
	logInputJSON, err := json.Marshal(logInput)
	if err != nil {
		log.Printf("Error marshaling log data: %v", err)
		return
	}

	logServiceURL := os.Getenv("LOG_SERVICE_URL")
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
