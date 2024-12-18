package entity

type ResponseOK struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
