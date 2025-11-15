package flags

var AllowedDatabases = []string{"postgres", "dynamodb"}

func IsValidDatabase(db string) bool {
	for _, allowed := range AllowedDatabases {
		if db == allowed {
			return true
		}
	}
	return false
}

