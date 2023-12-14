package wtime

import (
	"strconv"
	"time"
)

var (
	// Location time location
	Location = time.Local
	// StrDayFormat string time format
	StrDayFormat = "20060102"
	// StrMonthFormat month time format
	StrMonthFormat = "200601"
	// WeekdayStart week first day
	WeekdayStart = time.Monday // 默认周一是一周的第一天
)

func DayStamp() int64 {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).Unix()
}

func WeekStamp() int64 {
	now := time.Now().In(Location)
	// 0	Sunday
	// 1	Monday
	// 2	Tuesday
	// 3	Wednesday
	// 4	Thursday
	// 5	Friday
	// 6	Saturday
	offset := int(WeekdayStart - now.Weekday())
	if offset > 0 {
		offset -= 6
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).AddDate(0, 0, offset).Unix()
}

func MonthStamp() int64 {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, Location).Unix()
}

func YearStamp() int64 {
	now := time.Now().In(Location)
	return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, Location).Unix()
}

func CurDayString() string {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).Format(StrDayFormat)
}

func NextDayString() string {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).AddDate(0, 0, 1).Format(StrDayFormat)
}

func CurWeekString() string {
	now := time.Now().In(Location)
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset -= 6
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).AddDate(0, 0, offset).Format(StrDayFormat)
}

func NextWeekString() string {
	now := time.Now().In(Location)
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset -= 6
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location).AddDate(0, 0, offset+7).Format(StrDayFormat)
}

func CurMonthString() string {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, Location).Format(StrMonthFormat)
}

func NextMonthString() string {
	now := time.Now().In(Location)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, Location).AddDate(0, 1, 0).Format(StrMonthFormat)
}

func CurYearString() string {
	now := time.Now().In(Location)
	return strconv.Itoa(now.Year())
}

func NextYearString() string {
	t := time.Now().In(Location)
	return strconv.Itoa(t.Year() + 1)
}
