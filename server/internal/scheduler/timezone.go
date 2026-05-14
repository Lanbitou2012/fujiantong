package scheduler

import "time"

// loadTZ 加载东八区时区，失败则使用本地时区
func loadTZ() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Local
	}
	return loc
}
