package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	fmt.Printf("%s\n", time.Now().Format(time.RFC3339))
	s := "Replace this %1 and this %2\n"
	r := "STRING 1 REPLACED"
	r1 := "STRING 2 REPLACED"

	s = strings.Replace(strings.Replace(s, "%2", r1, 1), "%1", r, 1)

	fmt.Println(s)
}
