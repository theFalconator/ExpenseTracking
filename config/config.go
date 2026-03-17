package config

import (
	"os"
	"bufio"
	"strings"
)

const DATABASE_URL = "DATABASE_URL"

func ReadConfig() (map[string]string, error) {
	file, err := os.Open("develop.env")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !strings.Contains(line, "=") {
			continue
		}

		pieces := strings.SplitN(line, "=", 2)
		key := pieces[0]
		value := pieces[1]

		config[key] = value
	}

	return config, nil
}
