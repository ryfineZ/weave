package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	fmt.Fprintf(os.Stdout, "Weave daemon v%s\n", version)
}
