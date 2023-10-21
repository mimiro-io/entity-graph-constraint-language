package main

import (
	"flag"
	"fmt"
)

func main() {
	// Define command-line flags
	var (
		intFlag   int
		strFlag   string
		boolFlag  bool
		floatFlag float64
	)

	flag.IntVar(&intFlag, "int", 0, "An integer flag")
	flag.StringVar(&strFlag, "str", "", "A string flag")
	flag.BoolVar(&boolFlag, "bool", false, "A boolean flag")
	flag.Float64Var(&floatFlag, "float", 0.0, "A float64 flag")

	// Parse the command-line arguments
	flag.Parse()

	// Access the values of the parsed flags
	fmt.Printf("intFlag: %d\n", intFlag)
	fmt.Printf("strFlag: %s\n", strFlag)
	fmt.Printf("boolFlag: %v\n", boolFlag)
	fmt.Printf("floatFlag: %f\n", floatFlag)
}
