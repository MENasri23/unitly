package domain

// Page represents a paginated result containing a slice of items and pagination metadata.
type Page[T any] struct {
	// Data contains the items for the current page.
	Data []T
	// Page is the current page number (1-indexed).
	Page int
	// PageSize is the number of items per page.
	PageSize int
	// TotalCount is the total number of items across all pages.
	TotalCount int
}
