package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

const customLayout = "2006-01-02 15:04"

type CustomTime struct {
	time.Time
}

// UnmarshalJSON mengubah input JSON (string) ke CustomTime
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Hapus tanda kutip dari string waktu
	s = s[1 : len(s)-1]

	t, err := time.Parse(customLayout, s)
	if err != nil {
		return fmt.Errorf("invalid time format (use 'YYYY-MM-DD HH:mm'), got: %s", s)
	}

	ct.Time = t
	return nil
}

// MarshalJSON mengubah CustomTime ke format string JSON
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Format(customLayout))
}

// String memberi representasi string dari waktu
func (ct CustomTime) String() string {
	return ct.Format(customLayout)
}
