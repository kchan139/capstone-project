package utils

func ParseEnvKey(env string) string {
	for i, c := range env {
		if c == '=' {
			return env[:i]
		}
	}
	return env
}

func ParseEnvValue(env string) string {
	for i, c := range env {
		if c == '=' {
			return env[i+1:]
		}
	}
	return ""
}
