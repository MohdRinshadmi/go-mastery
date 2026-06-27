package main

import (
	"encoding/json"
	"fmt"
)

// Config is loaded from JSON.
//
// BUG: `host` and `timeout` are lowercase (unexported). encoding/json can only
// see EXPORTED (capitalized) fields, so these two are silently ignored on both
// decode and encode — no error, just empty values.
type Config struct {
	Port    int    `json:"port"`
	host    string `json:"host"`    //nolint -- intentionally buggy
	timeout int    `json:"timeout"` //nolint -- intentionally buggy
}

func main() {
	input := []byte(`{"port": 8080, "host": "db.internal", "timeout": 30}`)

	var cfg Config
	if err := json.Unmarshal(input, &cfg); err != nil {
		fmt.Println("decode error:", err)
		return
	}

	// host and timeout come back empty because json never populated them.
	fmt.Printf("Port=%d host=%q timeout=%d\n", cfg.Port, cfg.host, cfg.timeout)
	// Want: Port=8080 host="db.internal" timeout=30
	// Got:  Port=8080 host="" timeout=0

	// Re-encoding also drops them.
	out, _ := json.Marshal(cfg)
	fmt.Println("re-encoded:", string(out)) // {"port":8080} — host & timeout gone
}
