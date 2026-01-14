package api

import "os"

type Config struct {
	GenerateLimitPerMonth int
	ExportLimitPerMonth   int
	HuggingFaceAPIKey     string
	HuggingFaceModel      string
}

func LoadConfig() Config {
	return Config{
		GenerateLimitPerMonth: envInt("GENERATE_LIMIT_PER_MONTH", 50),
		ExportLimitPerMonth:   envInt("EXPORT_LIMIT_PER_MONTH", 200),
		HuggingFaceAPIKey:     envString("HUGGINGFACE_API_KEY", ""),
		HuggingFaceModel:      envString("HUGGINGFACE_MODEL", "mistralai/Mixtral-8x7B-Instruct-v0.1"),
	}
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func envString(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
