package http

import (
	"context"

	"github.com/jackc/pgx/v5"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
)

// IsDemoOwner returns true if the owner identified by ownerUUIDStr has is_demo set.
// Used to reject mutations (update/delete) on demo data. On lookup error other than
// no rows, returns (false, err).
func IsDemoOwner(ctx context.Context, queries *sqlc.Queries, ownerUUIDStr string) (bool, error) {
	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return false, err
	}
	owner, err := queries.GetOwnerByID(ctx, ownerUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return owner.IsDemo, nil
}
