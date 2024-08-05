package loggerTypes

import (
	logTypeEnum "github.com/Neeraj-Neurofin/requests-logger/store/enum"
	"time"
)

type PostLogInput struct {
	Type logTypeEnum.LogType 	`json:"type"`
	Data map[string]interface{} `json:"data"`
	TraceId string 				`json:"traceId"`
	Timestamp time.Time 		`json:"timestamp"`
}
