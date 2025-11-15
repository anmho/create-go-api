package flags

var AllowedFrameworks = []string{"chi", "connectrpc"}

func IsValidFramework(fw string) bool {
	for _, allowed := range AllowedFrameworks {
		if fw == allowed {
			return true
		}
	}
	return false
}

