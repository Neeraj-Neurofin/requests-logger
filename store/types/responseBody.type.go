package loggerTypes

type ResponseBody struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}