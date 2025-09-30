package util

func FormatSizeMBorGB(sizeMB int64) (float64, string) {
	if sizeMB >= 1024 {
		return float64(sizeMB) / 1024.0, "GB"
	}
	return float64(sizeMB), "MB"
}
func FormatSize(allocatedMB, usedMB int64) (allocatedVal, usedVal, wastedVal float64, unit string) {
	wastedMB := allocatedMB - usedMB

	if allocatedMB >= 1024 {
		unit = "GB"
		allocatedVal = float64(allocatedMB) / 1024
		usedVal = float64(usedMB) / 1024
		wastedVal = float64(wastedMB) / 1024
	} else {
		unit = "MB"
		allocatedVal = float64(allocatedMB)
		usedVal = float64(usedMB)
		wastedVal = float64(wastedMB)
	}

	return
}
