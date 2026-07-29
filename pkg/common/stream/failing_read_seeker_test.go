package stream

import "errors"

type failingReadSeeker struct {
	err error
}

func (s *failingReadSeeker) Read([]byte) (int, error) {
	return 0, errors.New("unexpected read")
}

func (s *failingReadSeeker) Seek(int64, int) (int64, error) {
	return 0, s.err
}
