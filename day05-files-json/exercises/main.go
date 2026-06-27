// Day 05 Phase-1 CAPSTONE — statsgen
//
// Reads a JSON file of sale records, aggregates per category, writes a report.
// Exercises slices, maps, structs, pointers, error wrapping, files, JSON.
//
// Run:  go run main.go data.json report.json
// Then inspect report.json.
//
// Fill the TODOs. Mentor reviews like a PR. Reference: ../solutions/main.go
package main

import (
	"fmt"
	"os"
)

type Sale struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Qty      int     `json:"qty"`
}

// CategoryStat is one line of the report.
type CategoryStat struct {
	Category    string  `json:"category"`
	TotalRevenue float64 `json:"total_revenue"` // sum of amount*qty
	UnitsSold   int     `json:"units_sold"`     // sum of qty
	OrderCount  int     `json:"order_count"`    // number of sales
}

// loadSales reads and decodes the JSON file.
// TODO: use os.ReadFile, json.Unmarshal into []Sale, wrap errors with %w.
func loadSales(path string) ([]Sale, error) {
	return nil, fmt.Errorf("TODO: implement loadSales")
}

// aggregate groups sales by category.
// TODO: build a map[string]*CategoryStat, accumulate, then return a slice
// SORTED by Category (import "sort"). Sorting matters: map order is random.
func aggregate(sales []Sale) []CategoryStat {
	return nil
}

// writeReport marshals stats (indented) and writes to path.
// TODO: json.MarshalIndent + os.WriteFile(perm 0644), wrap errors.
func writeReport(path string, stats []CategoryStat) error {
	return fmt.Errorf("TODO: implement writeReport")
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: statsgen <input.json> <output.json>")
		os.Exit(1)
	}
	in, out := os.Args[1], os.Args[2]

	sales, err := loadSales(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	stats := aggregate(sales)
	if err := writeReport(out, stats); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d category stats to %s\n", len(stats), out)
}
