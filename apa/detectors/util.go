package detectors

func padding(s string, max int) string {
	count := max - len(s)
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
}
