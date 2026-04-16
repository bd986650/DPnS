package helper

func PickSeat(pref string) string {
	switch pref {
	case "window":
		return "12A"
	case "aisle":
		return "12C"
	case "middle":
		return "12B"
	default:
		return "10C"
	}
}
