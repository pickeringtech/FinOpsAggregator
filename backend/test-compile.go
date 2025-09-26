// +build ignore

package main

import (
	"fmt"
	"os"

	_ "github.com/pickeringtech/FinOpsAggregator/internal/charts"
)

func main() {
	fmt.Println("✅ Compilation test passed!")
	os.Exit(0)
}
