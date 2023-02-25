package clock

import "time"

// Provider returns the current time.
type Provider interface {
	Now() time.Time
}

type localProvider struct{}

// Local provides the local system time.
func Local() Provider {
	return localProvider{}
}

// Now returns the current local time.
func (p localProvider) Now() time.Time {
	return time.Now()
}
