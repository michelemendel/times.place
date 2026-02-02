package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetTimestamp() time.Time {
	return time.Now()
}

func GetTimestampAsString() string {
	return time.Now().Format(time.RFC3339)
}

func GenerateToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

func PP(s any) {
	res, err := PrettyStruct(s)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(res)
}

func PrettyStruct(data any) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// UUIDToString converts pgtype.UUID to string
func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Convert [16]byte to uuid.UUID, then to string
	uuidVal, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		// Should not happen if bytes are correct length
		return ""
	}
	return uuidVal.String()
}

// StringToUUID converts string UUID to pgtype.UUID
func StringToUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var result pgtype.UUID
	// Convert uuid.UUID to [16]byte
	var bytes [16]byte
	copy(bytes[:], parsed[:])
	result.Bytes = bytes
	result.Valid = true
	return result, nil
}
