package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
)

func main() {
	limit := flag.Int("limit", 0, "Maximum number of venues to geocode (0 = no limit)")
	dryRun := flag.Bool("dry-run", false, "Do not write updates to DB")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL environment variable is required")
	}

	s, err := store.NewStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	rows, err := s.Queries.ListVenuesNeedingGeocode(ctx)
	if err != nil {
		log.Fatalf("Failed to list venues needing geocode: %v", err)
	}

	count := 0
	updated := 0
	for _, row := range rows {
		if *limit > 0 && updated >= *limit {
			break
		}
		count++
		geo, _ := service.MaybeGeocodeAddress(ctx, s.Queries, row.Address)
		if geo == "" {
			continue
		}
		if *dryRun {
			fmt.Printf("[dry-run] %s => %s\n", row.VenueUuid.String(), geo)
			updated++
			continue
		}
		if err := s.Queries.SetVenueGeolocation(ctx, sqlc.SetVenueGeolocationParams{
			VenueUuid:   row.VenueUuid,
			Geolocation: geo,
		}); err != nil {
			log.Printf("Failed to update venue %s: %v", row.VenueUuid.String(), err)
			continue
		}
		fmt.Printf("Updated %s => %s\n", row.VenueUuid.String(), geo)
		updated++
	}

	fmt.Printf("Scanned %d venues; updated %d.\n", count, updated)
}

