package main

import (
	"encoding/json"
	"fmt"
)

// Config is loaded from JSON.
//
// FIX: every field that JSON should (de)serialize must be EXPORTED (capitalized).
// The json tag still controls the on-the-wire key name, so the JSON stays
// lowercase while the Go fields are uppercase.
type Config struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	Timeout int    `json:"timeout"`
}

func main() {
	input := []byte(`{"port": 8080, "host": "db.internal", "timeout": 30}`)

	var cfg Config
	if err := json.Unmarshal(input, &cfg); err != nil {
		fmt.Println("decode error:", err)
		return
	}

	fmt.Printf("Port=%d Host=%q Timeout=%d\n", cfg.Port, cfg.Host, cfg.Timeout)
	// Port=8080 Host="db.internal" Timeout=30

	out, _ := json.Marshal(cfg)
	fmt.Println("re-encoded:", string(out))
	// {"port":8080,"host":"db.internal","timeout":30}
}
