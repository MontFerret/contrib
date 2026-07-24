package core

import stdhttp "net/http"

type tokenRoundTripFunc func(*stdhttp.Request) (*stdhttp.Response, error)

func (f tokenRoundTripFunc) RoundTrip(
	request *stdhttp.Request,
) (*stdhttp.Response, error) {
	return f(request)
}
