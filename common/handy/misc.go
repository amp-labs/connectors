package handy

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Returner[T any](self T) func() *T {
	return func() *T {
		return &self
	}
}
