package logger

import (
    "bytes"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/Neeraj-Neurofin/requests-logger/store/enum"
    "github.com/Neeraj-Neurofin/requests-logger/store/types"
)

func LogAuthServiceRequest(req *http.Request, traceID string) {
    var requestBody []byte
    if req.Body != nil {
        var err error
        requestBody, err = io.ReadAll(req.Body)
        if err != nil {
            log.Printf("Error reading request body: %v", err)
            return
        }
        req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
    }

    logData := map[string]interface{}{
        "method":         req.Method,
        "url":            req.URL.String(),
        "requestHeaders": req.Header,
        "requestBody":    string(requestBody),
        "traceID":        traceID,
        "timestamp":      time.Now(),
    }

    logInput := loggerTypes.PostLogInput{
        Type:      logTypeEnum.API,
        Data:      logData,
        TraceId:   traceID,
        Timestamp: time.Now(),
    }

    postLog(logInput)
}

func LogAuthServiceResponse(res *http.Response, traceID string) {
    var responseBodyInBytes []byte
    if res.Body != nil {
        var err error
        responseBodyInBytes, err = io.ReadAll(res.Body)
        if err != nil {
            log.Printf("Error reading response body: %v", err)
            return
        }
        res.Body = io.NopCloser(bytes.NewBuffer(responseBodyInBytes))
    }

    logData := map[string]interface{}{
        "responseStatus":  res.StatusCode,
        "responseHeaders": res.Header,
        "responseBody":    string(responseBodyInBytes),
        "traceID":         traceID,
        "timestamp":       time.Now(),
    }

    logInput := loggerTypes.PostLogInput{
        Type:      logTypeEnum.API,
        Data:      logData,
        TraceId:   traceID,
        Timestamp: time.Now(),
    }

    postLog(logInput)
}

func LogAuthServiceError(err error, traceID string) {
    logData := map[string]interface{}{
        "error":    err.Error(),
        "traceID":  traceID,
        "timestamp": time.Now(),
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
