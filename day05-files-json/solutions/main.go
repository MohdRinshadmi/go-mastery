// Day 05 Phase-1 CAPSTONE — statsgen (reference solution)
// Run: go run main.go data.json report.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type Sale struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Qty      int     `json:"qty"`
}

type CategoryStat struct {
	Category     string  `json:"category"`
	TotalRevenue float64 `json:"total_revenue"`
	UnitsSold    int     `json:"units_sold"`
	OrderCount   int     `json:"order_count"`
}

func loadSales(path string) ([]Sale, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadSales reading %s: %w", path, err)
	}
	var sales []Sale
	if err := json.Unmarshal(data, &sales); err != nil {
		return nil, fmt.Errorf("loadSales decoding %s: %w", path, err)
	}
	return sales, nil
}

func aggregate(sales []Sale) []CategoryStat {
	// pointer values so we accumulate in place
	byCat := make(map[string]*CategoryStat)
	for _, s := range sales {
		st, ok := byCat[s.Category]
		if !ok {
			st = &CategoryStat{Category: s.Category}
			byCat[s.Category] = st
		}
		st.TotalRevenue += s.Amount * float64(s.Qty)
		st.UnitsSold += s.Qty
		st.OrderCount++
	}

	out := make([]CategoryStat, 0, len(byCat))
	for _, st := range byCat {
		out = append(out, *st)
	}
	// deterministic order — map iteration is randomized
	sort.Slice(out, func(i, j int) bool { return out[i].Category < out[j].Category })
	return out
}

func writeReport(path string, stats []CategoryStat) error {
	b, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("writeReport marshal: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
		return fmt.Errorf("writeReport writing %s: %w", path, err)
	}
	return nil
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
