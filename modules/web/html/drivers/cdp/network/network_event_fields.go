package network

import cdpnetwork "github.com/mafredri/cdp/protocol/network"

func optionalEventString(value string) *string {
	if value == "" {
		return nil
	}

	return eventString(value)
}

func eventString(value string) *string {
	return &value
}

func optionalEventFloat(value float64) *float64 {
	if value == 0 {
		return nil
	}

	return eventFloat(value)
}

func eventFloat(value float64) *float64 {
	return &value
}

func eventBool(value bool) *bool {
	return &value
}

func eventInt(value int) *int {
	return &value
}

func eventRequestID(value *cdpnetwork.RequestID) *string {
	if value == nil {
		return nil
	}

	return optionalEventString(string(*value))
}
