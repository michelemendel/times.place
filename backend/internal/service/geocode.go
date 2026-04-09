package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sqlc "github.com/michelemendel/times.place/db/sqlc"
)

type nominatimSearchResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func normalizeAddress(address string) string {
	s := strings.TrimSpace(strings.ToLower(address))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func formatGeolocation(lat, lng float64) string {
	// Keep storage consistent with existing frontend behavior (6 decimals).
	return fmt.Sprintf("%.6f,%.6f", lat, lng)
}

// MaybeGeocodeAddress returns a "lat,lng" geolocation string when:
// - address is non-empty
// - we have a cached result, or Nominatim returns one
//
// If geocoding fails for any reason, it returns empty string and nil error
// (callers should treat geocoding as best-effort).
func MaybeGeocodeAddress(ctx context.Context, q *sqlc.Queries, address string) (string, error) {
	addrNorm := normalizeAddress(address)
	if addrNorm == "" {
		return "", nil
	}

	// 1) Cache lookup
	if cached, err := q.GetGeocodeCache(ctx, addrNorm); err == nil {
		return formatGeolocation(cached.Lat, cached.Lng), nil
	}

	// 2) External lookup (best-effort)
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	endpoint := "https://nominatim.openstreetmap.org/search"
	u, _ := url.Parse(endpoint)
	params := u.Query()
	params.Set("format", "jsonv2")
	params.Set("q", address)
	params.Set("limit", "1")
	params.Set("addressdetails", "0")
	params.Set("dedupe", "1")
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("User-Agent", "times.place/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil
	}

	var results []nominatimSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", nil
	}
	if len(results) == 0 {
		return "", nil
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return "", nil
	}
	lng, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return "", nil
	}

	_ = q.UpsertGeocodeCache(ctx, sqlc.UpsertGeocodeCacheParams{
		NormalizedAddress: addrNorm,
		Lat:               lat,
		Lng:               lng,
		DisplayName:       results[0].DisplayName,
	})

	return formatGeolocation(lat, lng), nil
}

