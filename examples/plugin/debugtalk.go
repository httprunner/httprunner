package main

import (
	"fmt"
	"log"
)

func init() {
	log.Println("plugin init function called")
}

func Concatenate(a int, b string, c float64) string {
	return fmt.Sprintf("%v_%v_%v", a, b, c)
}
