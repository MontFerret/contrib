package core

func cloneTokenExtra(extra map[string]any) map[string]any {
	if extra == nil {
		return nil
	}

	cloned := make(map[string]any, len(extra))

	for key, value := range extra {
		cloned[key] = cloneTokenExtraValue(value)
	}

	return cloned
}

func cloneTokenExtraValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return cloneTokenExtra(value)
	case []any:
		cloned := make([]any, len(value))

		for index, item := range value {
			cloned[index] = cloneTokenExtraValue(item)
		}

		return cloned
	case []string:
		return append([]string(nil), value...)
	case map[string]string:
		cloned := make(map[string]string, len(value))

		for key, item := range value {
			cloned[key] = item
		}

		return cloned
	default:
		return value
	}
}
