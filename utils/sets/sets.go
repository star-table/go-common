package sets

// 差集
func Difference(a []int64, b []int64) []int64 {
	var c []int64
	temp := map[int64]struct{}{}
	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}
	for _, val := range a {
		if _, ok := temp[val]; !ok {
			c = append(c, val)
		}
	}
	return c
}
