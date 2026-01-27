package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/michelemendel/times.place/internal/testdata"
)

func main() {
	clear := flag.Bool("clear", false, "Clear all test data before seeding")
	flag.Parse()

	// Get database URL from environment (required)
	// This should point to the development database, not the test database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL environment variable is required. This command seeds the development database.")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	// Clear existing data if requested
	if *clear {
		fmt.Println("Clearing existing test data...")
		if err := testdata.ClearTestData(ctx, db); err != nil {
			log.Fatalf("Failed to clear test data: %v", err)
		}
		fmt.Println("Test data cleared.")
	}

	// Seed test data
	fmt.Println("Seeding test data...")
	testData, err := testdata.SeedTestData(ctx, db)
	if err != nil {
		log.Fatalf("Failed to seed test data: %v", err)
	}

	fmt.Println("Test data seeded successfully!")
	fmt.Println()
	fmt.Println("Seeded data:")
	fmt.Printf("  Owner 1 (Abe): %s\n", testData.Owner1UUID)
	fmt.Printf("  Owner 2 (Ben): %s\n", testData.Owner2UUID)
	fmt.Printf("  Venue 1 (Beth El Synagogue): %s\n", testData.Venue1UUID)
	fmt.Printf("  Venue 2 (Community Center): %s\n", testData.Venue2UUID)
	fmt.Printf("  Venue 3 (Beit Midrash): %s\n", testData.Venue3UUID)
	fmt.Printf("  Venue 4 (Chagat House): %s\n", testData.Venue4UUID)
	fmt.Println()
	fmt.Println("Test credentials:")
	fmt.Println("  Owner 1: abe@demo.org / demo")
	fmt.Println("  Owner 2: ben@demo.org / demo")
}
