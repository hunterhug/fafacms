// Copyright 2021 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convert

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xorm.io/xorm/internal/utils"
)

// String2Time converts a string to time with original location
// be aware for time strings (HH:mm:ss) returns zero year (LMT) for converted location
func String2Time(s string, originalLocation *time.Location, convertedLocation *time.Location) (*time.Time, error) {
	if len(s) == 19 {
		if s == utils.ZeroTime0 || s == utils.ZeroTime1 {
			return &time.Time{}, nil
		}
		dt, err := time.ParseInLocation("2006-01-02 15:04:05", s, originalLocation)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) == 20 && s[10] == 'T' && s[19] == 'Z' {
		if strings.HasPrefix(s, "0000-00-00T00:00:00") || strings.HasPrefix(s, "0001-01-01T00:00:00") {
			return &time.Time{}, nil
		}
		dt, err := time.ParseInLocation("2006-01-02T15:04:05", s[:19], originalLocation)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) == 25 && s[10] == 'T' && s[19] == '+' && s[22] == ':' {
		if strings.HasPrefix(s, "0000-00-00T00:00:00") || strings.HasPrefix(s, "0001-01-01T00:00:00") {
			return &time.Time{}, nil
		}
		dt, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) >= 21 && s[10] == 'T' && s[19] == '.' {
		if strings.HasPrefix(s, "0000-00-00T00:00:00."+strings.Repeat("0", len(s)-20)) ||
			strings.HasPrefix(s, "0001-01-01T00:00:00."+strings.Repeat("0", len(s)-20)) {
			return &time.Time{}, nil
		}
		dt, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) >= 21 && s[19] == '.' {
		if strings.HasPrefix(s, "0000-00-00T00:00:00."+strings.Repeat("0", len(s)-20)) ||
			strings.HasPrefix(s, "0001-01-01T00:00:00."+strings.Repeat("0", len(s)-20)) {
			return &time.Time{}, nil
		}
		layout := "2006-01-02 15:04:05." + strings.Repeat("0", len(s)-20)
		dt, err := time.ParseInLocation(layout, s, originalLocation)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) == 10 && s[4] == '-' {
		if s == "0000-00-00" || s == "0001-01-01" {
			return &time.Time{}, nil
		}
		dt, err := time.ParseInLocation("2006-01-02", s, originalLocation)
		if err != nil {
			return nil, err
		}
		dt = dt.In(convertedLocation)
		return &dt, nil
	} else if len(s) == 8 && s[2] == ':' && s[5] == ':' {
		dt, err := time.ParseInLocation("15:04:05", s, originalLocation)
		if err != nil {
			return nil, err
		}
		dt = dt.AddDate(2006, 01, 02).In(convertedLocation)
		// back to zero year
		dt = dt.AddDate(-2006, -01, -02)
		return &dt, nil
	} else {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			if i == 0 {
				return &time.Time{}, nil
			}
			tm := time.Unix(i, 0).In(convertedLocation)
			return &tm, nil
		}
	}
	return nil, fmt.Errorf("unsupported conversion from %s to time", s)
}

// AsTime converts interface as time
func AsTime(src interface{}, dbLoc *time.Location, uiLoc *time.Location) (*time.Time, error) {
	switch t := src.(type) {
	case string:
		return String2Time(t, dbLoc, uiLoc)
	case *sql.NullString:
		if !t.Valid {
			return nil, nil
		}
		return String2Time(t.String, dbLoc, uiLoc)
	case []uint8:
		if t == nil {
			return nil, nil
		}
		return String2Time(string(t), dbLoc, uiLoc)
	case *sql.NullTime:
		if !t.Valid {
			return nil, nil
		}
		if utils.IsTimeZero(t.Time) {
			return &time.Time{}, nil
		}
		z, _ := t.Time.Zone()
		if len(z) == 0 || t.Time.Year() == 0 || t.Time.Location().String() != dbLoc.String() {
			tm := time.Date(t.Time.Year(), t.Time.Month(), t.Time.Day(), t.Time.Hour(),
				t.Time.Minute(), t.Time.Second(), t.Time.Nanosecond(), dbLoc).In(uiLoc)
			return &tm, nil
		}
		tm := t.Time.In(uiLoc)
		return &tm, nil
	case *time.Time:
		if utils.IsTimeZero(*t) {
			return &time.Time{}, nil
		}
		z, _ := t.Zone()
		if len(z) == 0 || t.Year() == 0 || t.Location().String() != dbLoc.String() {
			tm := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(),
				t.Minute(), t.Second(), t.Nanosecond(), dbLoc).In(uiLoc)
			return &tm, nil
		}
		tm := t.In(uiLoc)
		return &tm, nil
	case time.Time:
		if utils.IsTimeZero(t) {
			return &time.Time{}, nil
		}
		z, _ := t.Zone()
		if len(z) == 0 || t.Year() == 0 || t.Location().String() != dbLoc.String() {
			tm := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(),
				t.Minute(), t.Second(), t.Nanosecond(), dbLoc).In(uiLoc)
			return &tm, nil
		}
		tm := t.In(uiLoc)
		return &tm, nil
	case int:
		if t == 0 {
			return &time.Time{}, nil
		}
		tm := time.Unix(int64(t), 0).In(uiLoc)
		return &tm, nil
	case int64:
		if t == 0 {
			return &time.Time{}, nil
		}
		tm := time.Unix(t, 0).In(uiLoc)
		return &tm, nil
	case *sql.NullInt64:
		if t.Int64 == 0 {
			return &time.Time{}, nil
		}
		tm := time.Unix(t.Int64, 0).In(uiLoc)
		return &tm, nil
	}
	return nil, fmt.Errorf("unsupported value %#v as time", src)
}
