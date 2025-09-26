package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	fmt.Println("Testing minimal chart generation...")

	// Create a simple line chart
	graph := chart.Chart{
		Title: "Test Chart",
		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 20,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Test Series",
				XValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			},
		},
	}

	// Render to buffer
	var buf bytes.Buffer
	err := graph.Render(chart.PNG, &buf)
	if err != nil {
		fmt.Printf("Error rendering chart: %v\n", err)
		return
	}

	fmt.Printf("Success: Generated %d bytes\n", buf.Len())

	// Write to file
	if err := os.WriteFile("test-minimal.png", buf.Bytes(), 0644); err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
	} else {
		fmt.Println("Wrote test-minimal.png")
	}

	fmt.Println("Minimal chart test complete!")
}
