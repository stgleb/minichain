package main

import (
	"fmt"
	"github.com/willf/bloom"
	"math"
)

func main() {
	filter := bloom.New(2000, uint(math.Ceil(0.69*20)))
	fmt.Println(filter.EstimateFalsePositiveRate(100))
}
