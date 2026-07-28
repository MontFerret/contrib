package stream

type readerFunc func([]byte) (int, error)

func (fn readerFunc) Read(buffer []byte) (int, error) {
	return fn(buffer)
}
