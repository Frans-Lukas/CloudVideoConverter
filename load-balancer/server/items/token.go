package items

import "time"

type Token struct {
	CreationTime      *time.Time
	ConversionStarted *bool
	ConversionDone    *bool
	ConversionFailed  *bool
	OutputType        *string
}
