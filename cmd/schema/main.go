package main

import (
	"flag"
	"fmt"
)

func main() {
	var closedWorldFlag bool
	flag.BoolVar(&closedWorldFlag, "closedWorld", false, "Closed world assumption. Only allow what is defined in the model.")
	flag.Parse()
	fmt.Printf("closedWorldFlag: %v", closedWorldFlag)
}
