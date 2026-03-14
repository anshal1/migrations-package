package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type MigrationsConfig struct {
	InputDirectory string   `json:"inputDirectory"`
	Exclude        []string `json:"exclude"`
}

func getConfig() MigrationsConfig {
	var config MigrationsConfig
	path := filepath.Join(".", "m-config.json")
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return config
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		return config
	}
	return config
}

func main() {
	fmt.Println(getConfig())
}
