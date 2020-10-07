package paramutils

type MockPreqFlagSet struct {
	StringMap map[string]interface{}
}

func (fs *MockPreqFlagSet) GetStringOrDefault(flag, d string) string {
	if val, ok := fs.StringMap[flag]; ok {
		return val.(string)
	}

	return d
}

func (fs *MockPreqFlagSet) GetBoolOrDefault(flag string, d bool) bool {
	if val, ok := fs.StringMap[flag]; ok {
		return val.(bool)
	}

	return d
}
