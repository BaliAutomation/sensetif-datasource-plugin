package handler

func getParams(params map[string]string, names ...string) (values, missing []string) {
	for _, paramName := range names {
		if value, ok := params[paramName]; ok {
			values = append(values, value)
		} else {
			values = append(values, "")
			missing = append(missing, paramName)
		}
	}

	return values, missing
}
