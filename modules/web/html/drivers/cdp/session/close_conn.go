package session

type closeConn struct {
	close func() error
}

func (c *closeConn) Read(_ []byte) (int, error) {
	panic("closeConn.Read should never be called")
}

func (c *closeConn) Write(_ []byte) (int, error) {
	panic("closeConn.Write should never be called")
}

func (c *closeConn) Close() error {
	if c.close == nil {
		return nil
	}

	return c.close()
}
