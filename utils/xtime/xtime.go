package xtime

import (
	"time"
)

// IsLeapYear 判断给定年份是否为闰年
// 参数: year: 需要判断的年份
// 返回: 如果是闰年返回true，否则返回false
func IsLeapYear(year int) bool {
	// 如果年份能被4整除且不能被100整除，或者能被400整除，则为闰年
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// GetDayStartTime 获取一天的开始时间
// 参数: time: 需要获取开始时间的时间
// 返回: 一天的开始时间
func GetDayStartTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// GetDayEndTime 获取一天的结束时间
// 参数: time: 需要获取结束时间的时间
// 返回: 一天的结束时间
func GetDayEndTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// GetHourStartTime 获取小时的开始时间
// 参数: time: 需要获取开始时间的时间
// 返回: 小时的开始时间
func GetHourStartTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), 0, 0, 0, t.Location())
}

// GetHourEndTime 获取小时的结束时间
// 参数: time: 需要获取结束时间的时间
// 返回: 小时的结束时间
func GetHourEndTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), 59, 59, 999999999, t.Location())
}

// FormatTimeSecond 秒级时间戳转time.Time格式
// 参数: ts: 秒级时间戳
// 返回: time.Time格式的时间
func FormatTimeSecond(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// FormatTimeMillisecond 毫秒级时间戳转time.Time格式
// 参数: ts: 毫秒级时间戳
// 返回: time.Time格式的时间
func FormatTimeMillisecond(ts int64) time.Time {
	return time.Unix(ts/1e3, 0)
}

// FormatDuration 将时间持续时间转换持续运行时间
// 参数: d: 时间持续时间
// 返回: 持续运行时间 小时 分钟 秒
func FormatDuration(d time.Duration) (h, m, s int) {
	hh := d.Hours()
	if hh >= 1 {
		h = int(hh - 0.5)
	}
	mm := d.Minutes()
	if mm >= 1 {
		m = int(mm) % 60
	}
	ss := d.Seconds()
	if ss >= 1 {
		s = int(ss) % 60
	}
	return h, m, s
}
