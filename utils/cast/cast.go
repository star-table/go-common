package cast

import "github.com/spf13/cast"

func SliceStringToInt64(ss []string) []int64 {
	is := make([]int64, 0, len(ss))
	for _, s := range ss {
		is = append(is, cast.ToInt64(s))
	}

	return is
}
