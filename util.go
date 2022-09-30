package boorumux

import (
	"fmt"
)

func humanSize(b int) string {
	var f float64
	var r string
	if b > 1024*1024 {
		f = float64(b) / (1024 * 1024)
		r = "MB"
	} else if b > 1024 {
		f = float64(b) / 1024
		r = "KB"
	} else {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.2f %s", f, r)
}
