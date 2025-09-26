package charts

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// GraphRenderer renders graph visualizations
type GraphRenderer struct {
	store *store.Store
}

// NewGraphRenderer creates a new graph renderer
func NewGraphRenderer(store *store.Store) *GraphRenderer {
	return &GraphRenderer{
		store: store,
	}
}

// RenderGraphStructure renders the DAG structure as a visual graph
func (gr *GraphRenderer) RenderGraphStructure(ctx context.Context, date time.Time, output io.Writer, format string) error {
	// Build graph for the date
	builder := graph.NewGraphBuilder(gr.store)
	g, err := builder.BuildForDate(ctx, date)
	if err != nil {
		// If we can't build the graph (e.g., no database), render a "no data" chart
		return gr.RenderNoDataChart(ctx, fmt.Sprintf("Failed to build graph: %v", err), output, format)
	}

	// Get nodes and create simple layout
	nodes := g.Nodes()
	if len(nodes) == 0 {
		// If no nodes, render a "no data" chart
		return gr.RenderNoDataChart(ctx, "No nodes found in graph", output, format)
	}

	// Create a simple scatter plot showing nodes
	var xValues, yValues []float64
	var nodeNames []string

	// Simple circular layout
	centerX, centerY := 600.0, 400.0
	radius := 250.0
	i := 0

	for _, node := range nodes {
		angle := 2 * math.Pi * float64(i) / float64(len(nodes))
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)

		xValues = append(xValues, x)
		yValues = append(yValues, y)
		nodeNames = append(nodeNames, node.Name)
		i++
	}

	// Create chart
	chartGraph := chart.Chart{
		Title: fmt.Sprintf("FinOps Graph Structure (%s)", date.Format("2006-01-02")),
		TitleStyle: chart.Style{
			FontSize: 16,
		},
		Width:  1200,
		Height: 800,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name: "Nodes",
				Style: chart.Style{
					StrokeColor: drawing.ColorBlue,
					FillColor:   drawing.ColorBlue.WithAlpha(100),
					DotColor:    drawing.ColorBlue,
					DotWidth:    10,
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	// Render based on format
	switch format {
	case "png":
		return chartGraph.Render(chart.PNG, output)
	case "svg":
		return chartGraph.Render(chart.SVG, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}



// RenderCostTrend renders a cost trend chart for a specific node
func (gr *GraphRenderer) RenderCostTrend(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimension string, output io.Writer, format string) error {
	// Get node info first
	node, err := gr.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return gr.RenderNoDataChart(ctx, fmt.Sprintf("Failed to get node: %v", err), output, format)
	}

	// Get cost data - using the actual method signature from the store
	costs, err := gr.store.Costs.GetByNodeAndDateRange(ctx, nodeID, startDate, endDate, []string{dimension})
	if err != nil {
		return gr.RenderNoDataChart(ctx, fmt.Sprintf("Failed to get cost data: %v", err), output, format)
	}

	if len(costs) == 0 {
		// Create a placeholder chart with no data message
		return gr.RenderNoDataChart(ctx, fmt.Sprintf("No cost data found for %s (%s)", node.Name, dimension), output, format)
	}

	// Prepare data for chart
	var xValues []float64
	var yValues []float64

	for i, cost := range costs {
		xValues = append(xValues, float64(i)) // Simple index-based x-axis for now
		amount, _ := cost.Amount.Float64()
		yValues = append(yValues, amount)
	}

	// Create chart
	chartGraph := chart.Chart{
		Title: fmt.Sprintf("Cost Trend: %s (%s)", node.Name, dimension),
		TitleStyle: chart.Style{
			FontSize: 16,
		},
		Width:  1200,
		Height: 600,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		XAxis: chart.XAxis{
			Name: "Time Period",
		},
		YAxis: chart.YAxis{
			Name: "Cost Amount",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name: fmt.Sprintf("%s - %s", node.Name, dimension),
				Style: chart.Style{
					StrokeColor: drawing.ColorBlue,
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	// Render based on format
	switch format {
	case "png":
		return chartGraph.Render(chart.PNG, output)
	case "svg":
		return chartGraph.Render(chart.SVG, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}



// RenderAllocationWaterfall renders a waterfall chart showing cost allocation breakdown
func (gr *GraphRenderer) RenderAllocationWaterfall(ctx context.Context, nodeID uuid.UUID, date time.Time, runID uuid.UUID, output io.Writer, format string) error {
	// Get allocation results
	allocations, err := gr.store.Runs.GetAllocationResults(ctx, runID, store.AllocationResultFilters{
		NodeID:    nodeID,
		StartDate: date,
		EndDate:   date,
	})
	if err != nil {
		return fmt.Errorf("failed to get allocation results: %w", err)
	}
	
	if len(allocations) == 0 {
		return fmt.Errorf("no allocation results found")
	}
	
	// Get node info
	node, err := gr.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	
	// Prepare waterfall data
	var categories []string
	var values []float64
	var colors []drawing.Color
	
	totalDirect := 0.0
	totalIndirect := 0.0
	
	for _, allocation := range allocations {
		direct, _ := allocation.DirectAmount.Float64()
		indirect, _ := allocation.IndirectAmount.Float64()
		
		if direct > 0 {
			categories = append(categories, fmt.Sprintf("Direct\n%s", allocation.Dimension))
			values = append(values, direct)
			colors = append(colors, drawing.ColorBlue)
			totalDirect += direct
		}
		
		if indirect > 0 {
			categories = append(categories, fmt.Sprintf("Indirect\n%s", allocation.Dimension))
			values = append(values, indirect)
			colors = append(colors, drawing.ColorRed)
			totalIndirect += indirect
		}
	}
	
	// Add total
	categories = append(categories, "Total")
	values = append(values, totalDirect + totalIndirect)
	colors = append(colors, drawing.ColorGreen)
	
	// Create bar chart (simplified waterfall)
	bars := make([]chart.Value, len(categories))
	for i, category := range categories {
		bars[i] = chart.Value{
			Label: category,
			Value: values[i],
			Style: chart.Style{
				FillColor: colors[i],
			},
		}
	}
	
	graph := chart.BarChart{
		Title: fmt.Sprintf("Cost Allocation Breakdown: %s", node.Name),
		TitleStyle: chart.Style{
			FontSize: 16,
		},
		Width:  1200,
		Height: 600,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   20,
				Right:  20,
				Bottom: 60,
			},
		},
		YAxis: chart.YAxis{
			Name: "Cost Amount",
		},
		Bars: bars,
	}
	
	// Render based on format
	switch format {
	case "png":
		return graph.Render(chart.PNG, output)
	case "svg":
		return graph.Render(chart.SVG, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// RenderNoDataChart renders a simple chart indicating no data is available
func (gr *GraphRenderer) RenderNoDataChart(ctx context.Context, message string, output io.Writer, format string) error {
	// Create a simple chart with just text
	graph := chart.Chart{
		Title: "FinOps DAG Cost Attribution",
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   40,
				Right:  40,
				Bottom: 40,
			},
		},
		Width:  800,
		Height: 600,
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "No Data",
				XValues: []float64{0, 1},
				YValues: []float64{0, 0},
				Style: chart.Style{
					StrokeColor: chart.ColorTransparent,
					FillColor:   chart.ColorTransparent,
				},
			},
		},
	}

	// Add title with the message
	graph.Title = message
	graph.TitleStyle = chart.Style{
		FontSize:  16,
		FontColor: drawing.ColorRed,
	}

	// Render based on format
	if format == "svg" {
		return graph.Render(chart.SVG, output)
	}
	return graph.Render(chart.PNG, output)
}
