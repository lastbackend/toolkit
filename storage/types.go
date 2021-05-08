package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"
)

type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns NullString) UnmarshalJSON(v interface{}) error {
	if !ns.Valid {
		return nil
	}
	return json.Unmarshal([]byte(ns.String), v)
}

type NullInt64 struct {
	sql.NullInt64
}

type NullBool struct {
	sql.NullBool
}

type NullFloat64 struct {
	sql.NullFloat64
}

type NullTime struct {
	Time  time.Time
	Valid bool // Verified is true if Date is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

type NullStringArray struct {
	Slice []string
	Valid bool // Verified is true if array string is not NULL
}

// Scan implements the Scanner interface.
func (nsa *NullStringArray) Scan(value interface{}) error {
	if value == nil {
		nsa.Slice, nsa.Valid = make([]string, 0), false
		return nil
	}
	nsa.Valid = true
	nsa.Slice = strToStringSlice(string(value.([]byte)))
	return nil
}

// Value implements the driver Valuer interface.
func (nsa NullStringArray) Value() (driver.Value, error) {
	if !nsa.Valid {
		return nil, nil
	}
	return nsa.Slice, nil
}

func strToStringSlice(s string) []string {
	r := strings.Trim(s, "{}")
	a := make([]string, 0)
	for _, s := range strings.Split(r, ",") {
		a = append(a, s)
	}
	return a
}
