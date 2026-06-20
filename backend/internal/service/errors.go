package service

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ServiceError struct {
	Code    string
	Message string
	Details []FieldError
}

func (e *ServiceError) Error() string {
	return e.Message
}
