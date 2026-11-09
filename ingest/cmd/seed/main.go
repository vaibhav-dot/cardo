// Command seed generates 12 months of synthetic UK bank transactions.
//
// It writes two files:
//
//	data/transactions.csv   messy, bank-export-style rows (what Cardo ingests)
//	data/answers.json       the facts we planted, used to test analytics + agent
//
// Because we planted the answers, we can check Cardo's output exactly.
//
// Run: go run ./cmd/seed -out ../data
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Txn struct {
	Date        time.Time
	Description string  // raw, messy bank text
	Amount      float64 // negative = money out
	Merchant    string  // ground truth (not exported to the CSV)
	Category    string  // ground truth (not exported to the CSV)
}

type Subscription struct {
	Merchant   string  `json:"merchant"`
	DayOfMonth int     `json:"day_of_month"`
	Amount     float64 `json:"amount"`
	PriceRise  *Rise   `json:"price_rise,omitempty"`
	Cancelled  string  `json:"cancelled_after,omitempty"`
}

type Rise struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Month string  `json:"month"`
}

type Answers struct {
	Start         string             `json:"start"`
	End           string             `json:"end"`
	Subscriptions []Subscription     `json:"subscriptions"`
	Habits        []string           `json:"habits"`
	SpendByCat    map[string]float64 `json:"total_spend_by_category"`
}

var rng = rand.New(rand.NewSource(42)) // fixed seed → same data every run

func pick[T any](xs []T) T { return xs[rng.Intn(len(xs))] }

func money(min, max float64) float64 {
	v := min + rng.Float64()*(max-min)
	return float64(int(v*100)) / 100
}

func main() {
	out := flag.String("out", "../data", "output directory")
	flag.Parse()

	start := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	var txns []Txn
	add := func(d time.Time, desc string, amt float64, merchant, cat string) {
		txns = append(txns, Txn{d, desc, amt, merchant, cat})
	}

	// --- Planted subscriptions (the answers your analytics must find) ---
	subs := []Subscription{
		{Merchant: "Netflix", DayOfMonth: 3, Amount: 12.99, PriceRise: &Rise{From: 10.99, To: 12.99, Month: "2026-03"}},
		{Merchant: "Spotify", DayOfMonth: 14, Amount: 11.99},
		{Merchant: "PureGym", DayOfMonth: 1, Amount: 34.99},
		{Merchant: "Disney+", DayOfMonth: 20, Amount: 8.99, Cancelled: "2026-05"},
		{Merchant: "iCloud", DayOfMonth: 9, Amount: 2.99},
	}
	subDesc := map[string]string{
		"Netflix": "NETFLIX.COM 866-579-7172",
		"Spotify": "SPOTIFY P1F2A9C3 STOCKHOLM",
		"PureGym": "DD PUREGYM LTD",
		"Disney+": "DISNEY PLUS LONDON GB",
		"iCloud":  "APPLE.COM/BILL ITUNES.COM",
	}

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		month := d.Format("2006-01")
		wd := d.Weekday()

		// Income and fixed bills
		if d.Day() == 25 {
			add(d, "BGC ACME ANALYTICS LTD SALARY", 2850.00, "Acme Analytics", "Income")
		}
		if d.Day() == 1 {
			add(d, "SO RENT - J SMITH LANDLORD", -1250.00, "Rent", "Housing")
		}
		if d.Day() == 5 {
			add(d, "DD OCTOPUS ENERGY", -money(65, 110), "Octopus Energy", "Bills")
		}
		if d.Day() == 12 {
			add(d, "DD THAMES WATER", -32.40, "Thames Water", "Bills")
		}

		// Subscriptions
		for _, s := range subs {
			if d.Day() != s.DayOfMonth {
				continue
			}
			if s.Cancelled != "" && month > s.Cancelled {
				continue
			}
			amt := s.Amount
			if s.PriceRise != nil && month < s.PriceRise.Month {
				amt = s.PriceRise.From
			}
			add(d, subDesc[s.Merchant], -amt, s.Merchant, "Subscriptions")
		}

		// Habit: big weekly shop on Saturdays
		if wd == time.Saturday {
			store := pick([]string{"TESCO STORES 3297 LONDON GB", "TESCO STORES 2041 LONDON GB"})
			add(d, store, -money(55, 95), "Tesco", "Groceries")
		}
		// Habit: weekday coffee
		if wd >= time.Monday && wd <= time.Friday && rng.Float64() < 0.7 {
			add(d, "CARD PAYMENT TO PRET A MANGER", -money(3.2, 6.5), "Pret A Manger", "Eating Out")
		}
		// Habit: commuting
		if wd >= time.Monday && wd <= time.Friday && rng.Float64() < 0.85 {
			add(d, "TFL TRAVEL CH TFL.GOV.UK/CP", -money(5.6, 8.9), "TfL", "Transport")
		}
		// Top-up shops
		if rng.Float64() < 0.2 {
			add(d, pick([]string{"SAINSBURYS S/MKT", "CO-OP GROUP 070123", "LIDL GB LONDON"}), -money(6, 25), "Top-up shop", "Groceries")
		}
		// Friday/Saturday night out
		if (wd == time.Friday || wd == time.Saturday) && rng.Float64() < 0.5 {
			add(d, pick([]string{"DELIVEROO.CO.UK", "UBER *EATS", "THE CROWN & ANCHOR"}), -money(14, 48), "Takeaway/Pub", "Eating Out")
		}
		// Random shopping
		if rng.Float64() < 0.12 {
			add(d, fmt.Sprintf("AMZNMKTPLACE*%s", pick([]string{"2K4L91", "7HX0P2", "Q81ZZ3"})), -money(8, 70), "Amazon", "Shopping")
		}
	}

	// Planted anomaly: a one-off spike
	add(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), "CURRYS ONLINE", -649.00, "Currys", "Shopping")

	sort.Slice(txns, func(i, j int) bool { return txns[i].Date.Before(txns[j].Date) })

	if err := os.MkdirAll(*out, 0o755); err != nil {
		panic(err)
	}

	// 1. Bank-export style CSV (no merchant/category: Cardo has to work those out)
	writeCSV(filepath.Join(*out, "transactions.csv"), []string{"date", "description", "amount"}, txns,
		func(t Txn) []string {
			return []string{t.Date.Format("2006-01-02"), t.Description, fmt.Sprintf("%.2f", t.Amount)}
		})

	// 2. Labelled descriptions → golden dataset for the categoriser eval
	seen := map[string]bool{}
	var labelled []Txn
	for _, t := range txns {
		if !seen[t.Description] {
			seen[t.Description] = true
			labelled = append(labelled, t)
		}
	}
	writeCSV(filepath.Join(*out, "labelled.csv"), []string{"description", "amount", "__expected"}, labelled,
		func(t Txn) []string {
			return []string{t.Description, fmt.Sprintf("%.2f", t.Amount),
				fmt.Sprintf("javascript: output.trim() === '%s'", t.Category)}
		})

	// 3. Planted answers
	totals := map[string]float64{}
	for _, t := range txns {
		if t.Amount < 0 {
			totals[t.Category] += -t.Amount
		}
	}
	for k, v := range totals {
		totals[k] = float64(int(v*100+0.5)) / 100
	}
	ans := Answers{
		Start:         start.Format("2006-01-02"),
		End:           end.AddDate(0, 0, -1).Format("2006-01-02"),
		Subscriptions: subs,
		Habits: []string{
			"Weekly Tesco shop every Saturday (£55–95)",
			"Pret coffee on most weekdays",
			"Salary of £2,850 on the 25th",
			"One-off £649 Currys purchase on 2026-09-13 (anomaly)",
		},
		SpendByCat: totals,
	}
	f, err := os.Create(filepath.Join(*out, "answers.json"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ans); err != nil {
		panic(err)
	}

	fmt.Printf("wrote %d transactions, %d labelled descriptions to %s\n", len(txns), len(labelled), *out)
}

func writeCSV(path string, header []string, txns []Txn, row func(Txn) []string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write(header)
	for _, t := range txns {
		_ = w.Write(row(t))
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
}
