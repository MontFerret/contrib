package stream

type writerFunc func([]byte) (int, error)

func (fn writerFunc) Write(buffer []byte) (int, error) {
	return fn(buffer)
}
