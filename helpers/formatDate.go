package helper

import (
	"fmt"
	"time"
)

func FormatLocalDate(format, date string) string {
	monthName := []string{
		"Januari", "Februari", "Maret", "April",
		"Mei", "Juni", "Juli", "Agustus",
		"September", "Oktober", "November", "Desember",
	}

	current := time.Now().UTC()

	d := current.Day()
	y, m, _ := current.Date()
	strFormat := "%s %02d"
	tm2 := fmt.Sprintf(strFormat, monthName[m-2], y)

	if format == "2006-01-02" {
		strFormat = "%d %s %02d"
		tm2 = fmt.Sprintf(strFormat, d, monthName[m-2], y)
	}

	if date != "" {
		now, _ := time.Parse(format, date)
		y = now.Year()
		m = now.Month()
		d = now.Day()
		tm2 = fmt.Sprintf(strFormat, monthName[m-1], y)
		if format == "2006-01-02" {
			strFormat = "%d %s %02d"
			tm2 = fmt.Sprintf(strFormat, d, monthName[m-1], y)
		}
	}
	return tm2
}
