package domain

type Page[T any] struct {
	Data       []T
	Page       int
	PageSize   int
	TotalCount int
}
