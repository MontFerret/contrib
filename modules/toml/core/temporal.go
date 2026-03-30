package core

import (
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	localDateTimeLocation = "datetime-local"
	localDateLocation     = "date-local"
	localTimeLocation     = "time-local"

	localDateTimeLayout = "2006-01-02T15:04:05.999999999"
	localDateLayout     = "2006-01-02"
	localTimeLayout     = "15:04:05.999999999"
)

func decodeTemporalValue(value time.Time, opts DecodeOptions) (runtime.Value, error) {
	switch opts.DateTime {
	case DecodeDateTimeString:
		return runtime.NewString(canonicalTOMLDateTime(value)), nil
	case DecodeDateTimeNative:
		return runtime.NewDateTime(value), nil
	default:
		return nil, newErrorf(`unsupported decode datetime mode %q`, opts.DateTime)
	}
}

func encodeTemporalValue(value runtime.DateTime, opts EncodeOptions) (string, error) {
	switch opts.DateTime {
	case EncodeDateTimeRFC3339:
		return value.Time.Format(time.RFC3339Nano), nil
	case EncodeDateTimePreserve:
		return canonicalTOMLDateTime(value.Time), nil
	default:
		return "", newErrorf(`unsupported encode datetime mode %q`, opts.DateTime)
	}
}

func canonicalTOMLDateTime(value time.Time) string {
	switch temporalLocation(value) {
	case localDateTimeLocation:
		return value.Format(localDateTimeLayout)
	case localDateLocation:
		return value.Format(localDateLayout)
	case localTimeLocation:
		return value.Format(localTimeLayout)
	default:
		return value.Format(time.RFC3339Nano)
	}
}

func temporalLocation(value time.Time) string {
	if value.Location() == nil {
		return ""
	}

	return value.Location().String()
}
