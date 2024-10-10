package deep

type EmptyCloser struct{}

func (e EmptyCloser) Close() error {
	return nil
}
