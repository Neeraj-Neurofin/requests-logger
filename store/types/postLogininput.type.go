package loggerTypes

import (
	"time"

	logTypeEnum "github.com/Neeraj-Neurofin/requests-logger/store/enum"
)

type PostLogInput struct {
	Type logTypeEnum.LogType 	`json:"type"`
	Data map[string]interface{} `json:"data"`
	TraceId string 				`json:"traceId"`
	Timestamp time.Time 		`json:"timestamp"`
}
