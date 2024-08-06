package loggerTypes

import (
	"errors"
	"time"

	logTypeEnum "github.com/Neeraj-Neurofin/requests-logger/store/enum"
)

type PostLogInput struct {
	Type logTypeEnum.LogType 	`json:"type"`
	Data map[string]interface{} `json:"data"`
	TraceId string 				`json:"traceId"`
	Timestamp time.Time 		`json:"timestamp"`
}

func (i *PostLogInput) Validate() error {
	if err := i.Type.Validate(); err != nil {
		return err
	}

	if i.TraceId == "" {
		return errors.New("trace id not found")
	}

	return nil
}