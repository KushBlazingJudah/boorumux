package boorumux

func mutual[T comparable](i, j []T) bool {
	for _, v := range i {
		for _, w := range j {
			if v == w {
				return true
			}
		}
	}
	return false
}
