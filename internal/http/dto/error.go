package dto

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// NewErrorResponse для хендлеров
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	}
}
