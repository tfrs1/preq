package mocks

type PreqFlagSetMock struct {
	StringMap map[string]interface{}
}

func (fs *PreqFlagSetMock) GetStringOrDefault(flag, d string) string {
	if val, ok := fs.StringMap[flag]; ok {
		return val.(string)
	}

	return d
}

func (fs *PreqFlagSetMock) GetBoolOrDefault(flag string, d bool) bool {
	if val, ok := fs.StringMap[flag]; ok {
		return val.(bool)
	}

	return d
}
