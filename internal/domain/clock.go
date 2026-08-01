package domain

import "time"

// Clock allows mocking the current time in unit tests.
type Clock interface {
	Now() time.Time
}

// realClock implements Clock using system time.
type realClock struct{}

func NewRealClock() Clock {
	return realClock{}
}

func (realClock) Now() time.Time {
	return time.Now()
}
