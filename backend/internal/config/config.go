// Date: 2026-05-25
// Author: XinYang Li

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config contains the runtime settings for the backend service.
type Config struct {
	Port           string
	DatabaseURL    string
	FrontendOrigin string
	AgentRootDir   string
	PythonBin      string
	RedisHost      string
	RedisPort      string
}

/**
 * Load builds backend configuration from environment variables and local .env files.
 * Params:
 * - none: the function reads directly from process environment variables and fallback env files.
 */
func Load() Config {
	envValues := loadEnvFiles("../.env", ".env")

	port := envOrFallback("AGENTHUB_BACKEND_PORT", envValues, "8080")
	frontendOrigin := envOrFallback("AGENTHUB_FRONTEND_ORIGIN", envValues, "http://192.168.139.155:3000")
	agentRootDir := envOrFallback("AGENTHUB_AGENT_ROOT_DIR", envValues, "../agent")
	pythonBin := envOrFallback("AGENTHUB_PYTHON_BIN", envValues, "python3")
	redisHost := envOrFallback("Redis_HOST", envValues, "")
	redisPort := envOrFallback("Redis_PORT", envValues, "")

	databaseURL := os.Getenv("AGENTHUB_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = envValues["AGENTHUB_DATABASE_URL"]
	}
	if databaseURL == "" {
		pgHost := envOrFallback("PG_HOST", envValues, "localhost")
		pgPort := envOrFallback("PG_PORT", envValues, "5432")
		pgUser := envOrFallback("PG_USER", envValues, "postgres")
		pgPassword := envOrFallback("PG_PASSWORD", envValues, "postgres")
		pgDB := envOrFallback("PG_DB", envValues, "agenthub")
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPassword, pgHost, pgPort, pgDB)
	}

	return Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		FrontendOrigin: frontendOrigin,
		AgentRootDir:   agentRootDir,
		PythonBin:      pythonBin,
		RedisHost:      redisHost,
		RedisPort:      redisPort,
	}
}

func envOrFallback(key string, values map[string]string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := values[key]; value != "" {
		return value
	}
	return fallback
}

func loadEnvFiles(paths ...string) map[string]string {
	values := map[string]string{}
	for _, path := range paths {
		absolutePath := path
		if !filepath.IsAbs(path) {
			absolutePath = filepath.Clean(path)
		}

		file, err := os.Open(absolutePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			values[key] = value
		}

		_ = file.Close()
	}

	return values
}
