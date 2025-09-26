package charts

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
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
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// Create a simple node layout
	nodes := g.Nodes()
	nodePositions := gr.calculateNodePositions(g)
	
	// Create chart
	graph := chart.Chart{
		Width:  1200,
		Height: 800,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    20,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		Series: []chart.Series{
			gr.createNodeSeries(nodes, nodePositions),
			gr.createEdgeSeries(g, nodePositions),
		},
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

// calculateNodePositions calculates positions for nodes in a hierarchical layout
func (gr *GraphRenderer) calculateNodePositions(g *graph.Graph) map[uuid.UUID]Position {
	positions := make(map[uuid.UUID]Position)
	
	// Get topological order to determine levels
	order, err := g.TopologicalSort()
	if err != nil {
		// Fallback to simple layout
		return gr.simpleLayout(g.Nodes())
	}
	
	// Group nodes by level (distance from roots)
	levels := make(map[int][]uuid.UUID)
	nodeLevel := make(map[uuid.UUID]int)
	
	// Calculate levels using BFS from roots
	roots := g.GetRoots()
	queue := make([]uuid.UUID, 0)
	visited := make(map[uuid.UUID]bool)
	
	// Start with roots at level 0
	for _, root := range roots {
		levels[0] = append(levels[0], root)
		nodeLevel[root] = 0
		queue = append(queue, root)
		visited[root] = true
	}
	
	// BFS to assign levels
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentLevel := nodeLevel[current]
		
		// Process children
		for _, edge := range g.Edges(current) {
			child := edge.ChildID
			if !visited[child] {
				level := currentLevel + 1
				levels[level] = append(levels[level], child)
				nodeLevel[child] = level
				queue = append(queue, child)
				visited[child] = true
			}
		}
	}
	
	// Position nodes within levels
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}
	
	for level, nodesInLevel := range levels {
		y := float64(level) / float64(maxLevel) * 600 + 100 // Y position based on level
		
		for i, nodeID := range nodesInLevel {
			x := float64(i) / math.Max(1, float64(len(nodesInLevel)-1)) * 1000 + 100
			positions[nodeID] = Position{X: x, Y: y}
		}
	}
	
	return positions
}

// simpleLayout creates a simple circular layout for nodes
func (gr *GraphRenderer) simpleLayout(nodes map[uuid.UUID]*models.CostNode) map[uuid.UUID]Position {
	positions := make(map[uuid.UUID]Position)
	
	nodeList := make([]uuid.UUID, 0, len(nodes))
	for id := range nodes {
		nodeList = append(nodeList, id)
	}
	
	centerX, centerY := 600.0, 400.0
	radius := 250.0
	
	for i, nodeID := range nodeList {
		angle := 2 * math.Pi * float64(i) / float64(len(nodeList))
		x := centerX + radius * math.Cos(angle)
		y := centerY + radius * math.Sin(angle)
		positions[nodeID] = Position{X: x, Y: y}
	}
	
	return positions
}

// createNodeSeries creates a scatter series for nodes
func (gr *GraphRenderer) createNodeSeries(nodes map[uuid.UUID]*models.CostNode, positions map[uuid.UUID]Position) chart.ContinuousSeries {
	var xValues, yValues []float64
	
	for nodeID, pos := range positions {
		xValues = append(xValues, pos.X)
		yValues = append(yValues, pos.Y)
	}
	
	return chart.ContinuousSeries{
		Style: chart.Style{
			StrokeColor: drawing.ColorBlue,
			FillColor:   drawing.ColorBlue.WithAlpha(100),
			DotColor:    drawing.ColorBlue,
		},
		XValues: xValues,
		YValues: yValues,
	}
}

// createEdgeSeries creates line series for edges
func (gr *GraphRenderer) createEdgeSeries(g *graph.Graph, positions map[uuid.UUID]Position) chart.ContinuousSeries {
	var xValues, yValues []float64
	
	// Draw edges as lines
	for parentID, edges := range g.Edges(parentID) {
		parentPos, parentExists := positions[parentID]
		if !parentExists {
			continue
		}
		
		for _, edge := range edges {
			childPos, childExists := positions[edge.ChildID]
			if !childExists {
				continue
			}
			
			// Add line from parent to child
			xValues = append(xValues, parentPos.X, childPos.X, math.NaN())
			yValues = append(yValues, parentPos.Y, childPos.Y, math.NaN())
		}
	}
	
	return chart.ContinuousSeries{
		Style: chart.Style{
			StrokeColor: drawing.ColorRed,
			StrokeWidth: 2,
		},
		XValues: xValues,
		YValues: yValues,
	}
}

// Position represents a 2D position
type Position struct {
	X, Y float64
}

// RenderCostTrend renders a cost trend chart for a specific node
func (gr *GraphRenderer) RenderCostTrend(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimension string, output io.Writer, format string) error {
	// Get cost data
	costs, err := gr.store.Costs.GetByNodeAndDateRange(ctx, nodeID, startDate, endDate, []string{dimension})
	if err != nil {
		return fmt.Errorf("failed to get cost data: %w", err)
	}
	
	if len(costs) == 0 {
		return fmt.Errorf("no cost data found for node %s", nodeID)
	}
	
	// Get node info
	node, err := gr.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	
	// Prepare data for chart
	var dates []time.Time
	var amounts []float64
	
	for _, cost := range costs {
		dates = append(dates, cost.CostDate)
		amount, _ := cost.Amount.Float64()
		amounts = append(amounts, amount)
	}
	
	// Create time series
	timeSeries := chart.TimeSeries{
		Name: fmt.Sprintf("%s - %s", node.Name, dimension),
		Style: chart.Style{
			StrokeColor: drawing.ColorBlue,
			StrokeWidth: 2,
		},
	}
	
	for i, date := range dates {
		timeSeries.XValues = append(timeSeries.XValues, date)
		timeSeries.YValues = append(timeSeries.YValues, amounts[i])
	}
	
	// Create chart
	graph := chart.Chart{
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
			Name: "Date",
			Style: chart.Style{
				TextRotationDegrees: 45,
			},
		},
		YAxis: chart.YAxis{
			Name: fmt.Sprintf("Cost (%s)", costs[0].Currency),
		},
		Series: []chart.Series{
			timeSeries,
		},
	}
	
	// Add legend
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
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
	var colors []color.Color
	
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
