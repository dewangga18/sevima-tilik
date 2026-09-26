package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"github.com/sevima/tilik-api/internal/service"
)

func run() error {
	dir := flag.String("dir", "", "Directory containing questions_*.json (server-side only)")
	reviewFile := flag.String("reviews", "", "Hash-bound review manifest; no manifest means drafts only")
	apply := flag.Bool("apply", false, "Commit validated candidates and approved questions; default dry-run")
	flag.Parse()
	if *dir == "" {
		return fmt.Errorf("--dir is required")
	}
	reviews := map[string]domain.CandidateReview{}
	if *reviewFile != "" {
		body, err := os.ReadFile(*reviewFile)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(body, &reviews); err != nil {
			return err
		}
	}
	files, err := filepath.Glob(filepath.Join(*dir, "questions_*.json"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no questions_*.json files found")
	}
	var entries []domain.BankEntry
	ids := map[string]bool{}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		parsed, err := service.ParseQuestionCandidates(body, filepath.Base(file), reviews)
		if err != nil {
			return err
		}
		for _, e := range parsed {
			if ids[e.Candidate.ID] {
				return fmt.Errorf("duplicate ID: %s", e.Candidate.ID)
			}
			ids[e.Candidate.ID] = true
		}
		entries = append(entries, parsed...)
	}
	for id := range reviews {
		if !ids[id] {
			return fmt.Errorf("review refers to missing question %s", id)
		}
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("database open failed")
	}
	defer sqlDB.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database unavailable; check backend DATABASE_URL")
	}
	db := &repository.DB{DB: sqlDB}
	// Migrations/seeding belong to app startup, never to dry-run import.
	report, err := db.RegisterQuestionBank(ctx, entries, *apply)
	if err != nil {
		return err
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(report)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Question import failed:", err)
		os.Exit(1)
	}
}
