package types

// This represents a IBL xavagebb API Error
type ApiError struct {
	Context map[string]string `json:"context,omitempty" description:"Context of the error. Usually used for validation error contexts"`
	Message string            `json:"message" description:"Message of the error"`
}

// Paged result common
type PagedResult[T any] struct {
	Count   uint64 `json:"count"`
	PerPage uint64 `json:"per_page"`
	Results T      `json:"results"`
}
