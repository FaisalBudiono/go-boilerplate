package domain

import (
	"slices"
	"strconv"
	"strings"
)

type IMEI string

func (i IMEI) String() string {
	return string(i)
}

func (i IMEI) IsValid() bool {
	numbers := strings.Split(string(i), "")
	if len(numbers) != 15 {
		return false
	}

	slices.Reverse(numbers)

	total := 0
	for i, num := range numbers {
		ori, err := strconv.Atoi(num)
		if err != nil {
			return false
		}

		if i%2 == 0 {
			total += ori
			continue
		}

		doubled := ori * 2
		if doubled > 9 {
			doubled -= 9
		}
		total += doubled
	}

	return total%10 == 0
}
