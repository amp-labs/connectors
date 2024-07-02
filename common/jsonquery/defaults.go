package jsonquery

func (q *Query) IntegerWithDefault(key string, defaultValue int64) (int64, error) {
	result, err := q.Integer(key, true)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}

func (q *Query) StrWithDefault(key string, defaultValue string) (string, error) {
	result, err := q.Str(key, true)
	if err != nil {
		return "", err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}

func (q *Query) BoolWithDefault(key string, defaultValue bool) (bool, error) {
	result, err := q.Bool(key, true)
	if err != nil {
		return false, err
	}

	if result == nil {
		return defaultValue, nil
	}

	return *result, nil
}
