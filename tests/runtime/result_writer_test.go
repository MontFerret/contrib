package main

import "bytes"

type resultWriter struct {
	data     bytes.Buffer
	err      error
	calls    int
	failCall int
}

func (w *resultWriter) Write(data []byte) (int, error) {
	w.calls++

	if w.calls == w.failCall {
		return 0, w.err
	}

	return w.data.Write(data)
}
