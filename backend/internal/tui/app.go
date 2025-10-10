package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pickeringtech/FinOpsAggregator/internal/analysis"
	"github.com/pickeringtech/FinOpsAggregator/internal/reports"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rivo/tview"
	"github.com/shopspring/decimal"
)

// App represents the TUI application
type App struct {
	app       *tview.Application
	store     *store.Store
	analyzer  *analysis.FinOpsAnalyzer
	generator *reports.ReportGenerator
	
	// UI components
	pages     *tview.Pages
	sidebar   *tview.List
	content   *tview.Flex
	
	// Current state
	currentView string
	dateRange   struct {
		start time.Time
		end   time.Time
	}
}

// NewApp creates a new TUI application
func NewApp(store *store.Store) *App {
	app := &App{
		app:       tview.NewApplication(),
		store:     store,
		analyzer:  analysis.NewFinOpsAnalyzer(store),
		generator: reports.NewReportGenerator(store),
	}
	
	// Set default date range (last 30 days)
	app.dateRange.end = time.Now()
	app.dateRange.start = app.dateRange.end.AddDate(0, 0, -30)
	
	app.setupUI()
	return app
}

// Run starts the TUI application
func (a *App) Run() error {
	return a.app.Run()
}

// setupUI initializes the user interface
func (a *App) setupUI() {
	// Create main layout
	a.pages = tview.NewPages()
	
	// Create sidebar
	a.sidebar = tview.NewList().
		AddItem("📊 Cost Overview", "View cost summary and trends", '1', a.showCostOverview).
		AddItem("🔍 Cost Analysis", "Detailed cost breakdown by nodes", '2', a.showCostAnalysis).
		AddItem("💡 Optimization", "Cost optimization insights", '3', a.showOptimization).
		AddItem("📈 Allocation", "Cost allocation efficiency", '4', a.showAllocation).
		AddItem("📋 Reports", "Generate comprehensive reports", '5', a.showReports).
		AddItem("⚙️  Settings", "Configure date ranges and options", '6', a.showSettings).
		AddItem("❌ Exit", "Exit the application", 'q', func() { a.app.Stop() })
	
	a.sidebar.SetBorder(true).
		SetTitle(" FinOps Dashboard ").
		SetTitleAlign(tview.AlignCenter)
	
	// Create content area
	a.content = tview.NewFlex().SetDirection(tview.FlexRow)
	a.content.SetBorder(true).SetTitle(" Content ")
	
	// Create main layout
	main := tview.NewFlex().
		AddItem(a.sidebar, 25, 0, true).
		AddItem(a.content, 0, 1, false)
	
	a.pages.AddPage("main", main, true, true)
	a.app.SetRoot(a.pages, true)
	
	// Set up key bindings
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		case tcell.KeyTab:
			// Switch focus between sidebar and content
			if a.app.GetFocus() == a.sidebar {
				a.app.SetFocus(a.content)
			} else {
				a.app.SetFocus(a.sidebar)
			}
			return nil
		}
		return event
	})
	
	// Show initial view
	a.showCostOverview()
}

// showCostOverview displays the cost overview
func (a *App) showCostOverview() {
	a.currentView = "overview"
	a.content.Clear()
	
	// Create loading message
	loading := tview.NewTextView().
		SetText("Loading cost overview...").
		SetTextAlign(tview.AlignCenter)
	a.content.AddItem(loading, 0, 1, false)
	a.app.Draw()
	
	// Load data in background
	go func() {
		ctx := context.Background()
		summary, err := a.analyzer.AnalyzeCosts(ctx, a.dateRange.start, a.dateRange.end)
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.showError(fmt.Sprintf("Failed to load cost data: %v", err))
			})
			return
		}
		
		a.app.QueueUpdateDraw(func() {
			a.displayCostOverview(summary)
		})
	}()
}

// displayCostOverview shows the cost overview data
func (a *App) displayCostOverview(summary *analysis.CostSummary) {
	a.content.Clear()
	
	// Create overview text
	overview := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	
	text := fmt.Sprintf(`[yellow]Cost Overview - %s[white]

[green]Total Cost:[white] $%s
[green]Period:[white] %s
[green]Nodes:[white] %d
[green]Dimensions:[white] %d

[yellow]Top Cost Nodes:[white]
`, summary.Period, summary.TotalCost.String(), summary.Period, len(summary.ByNode), len(summary.ByDimension))
	
	for i, node := range summary.TopCosts {
		if i >= 5 { // Show top 5
			break
		}
		text += fmt.Sprintf("  %d. %s: $%s (%.1f%%)\n", i+1, node.NodeName, node.Cost.String(), node.Percentage)
	}
	
	text += "\n[yellow]Cost by Dimension:[white]\n"
	for dim, cost := range summary.ByDimension {
		percentage := cost.Div(summary.TotalCost).InexactFloat64() * 100
		text += fmt.Sprintf("  • %s: $%s (%.1f%%)\n", dim, cost.String(), percentage)
	}
	
	text += "\n[blue]Press Tab to navigate, 'q' to quit[white]"
	
	overview.SetText(text)
	a.content.AddItem(overview, 0, 1, false)
}

// showCostAnalysis displays detailed cost analysis
func (a *App) showCostAnalysis() {
	a.currentView = "analysis"
	a.content.Clear()
	
	loading := tview.NewTextView().
		SetText("Loading cost analysis...").
		SetTextAlign(tview.AlignCenter)
	a.content.AddItem(loading, 0, 1, false)
	a.app.Draw()
	
	go func() {
		ctx := context.Background()
		summary, err := a.analyzer.AnalyzeCosts(ctx, a.dateRange.start, a.dateRange.end)
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.showError(fmt.Sprintf("Failed to load analysis: %v", err))
			})
			return
		}
		
		a.app.QueueUpdateDraw(func() {
			a.displayCostAnalysis(summary)
		})
	}()
}

// displayCostAnalysis shows detailed cost analysis
func (a *App) displayCostAnalysis(summary *analysis.CostSummary) {
	a.content.Clear()
	
	// Create table for detailed analysis
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)
	
	// Headers
	headers := []string{"Node Name", "Type", "Cost", "Percentage", "Trend"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}
	
	// Data rows
	for row, node := range summary.TopCosts {
		table.SetCell(row+1, 0, tview.NewTableCell(node.NodeName))
		table.SetCell(row+1, 1, tview.NewTableCell(node.NodeType))
		table.SetCell(row+1, 2, tview.NewTableCell(fmt.Sprintf("$%s", node.Cost.String())))
		table.SetCell(row+1, 3, tview.NewTableCell(fmt.Sprintf("%.1f%%", node.Percentage)))
		table.SetCell(row+1, 4, tview.NewTableCell("📈")) // Placeholder for trend
	}
	
	a.content.AddItem(table, 0, 1, false)
}

// showOptimization displays optimization insights
func (a *App) showOptimization() {
	a.currentView = "optimization"
	a.content.Clear()
	
	loading := tview.NewTextView().
		SetText("Generating optimization insights...").
		SetTextAlign(tview.AlignCenter)
	a.content.AddItem(loading, 0, 1, false)
	a.app.Draw()
	
	go func() {
		ctx := context.Background()
		insights, err := a.analyzer.GenerateOptimizationInsights(ctx, a.dateRange.start, a.dateRange.end)
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.showError(fmt.Sprintf("Failed to generate insights: %v", err))
			})
			return
		}
		
		a.app.QueueUpdateDraw(func() {
			a.displayOptimization(insights)
		})
	}()
}

// displayOptimization shows optimization insights
func (a *App) displayOptimization(insights []analysis.CostOptimizationInsight) {
	a.content.Clear()
	
	if len(insights) == 0 {
		noInsights := tview.NewTextView().
			SetText("No optimization insights found for the current period.").
			SetTextAlign(tview.AlignCenter)
		a.content.AddItem(noInsights, 0, 1, false)
		return
	}
	
	// Create list of insights
	list := tview.NewList()
	
	totalSavings := decimal.Zero
	for _, insight := range insights {
		totalSavings = totalSavings.Add(insight.PotentialSavings)
		
		severityIcon := "💡"
		switch insight.Severity {
		case "high":
			severityIcon = "🔴"
		case "medium":
			severityIcon = "🟡"
		case "low":
			severityIcon = "🟢"
		}
		
		title := fmt.Sprintf("%s %s - $%s savings", severityIcon, insight.Title, insight.PotentialSavings.String())
		desc := fmt.Sprintf("Node: %s | %s", insight.NodeName, insight.Description)
		
		list.AddItem(title, desc, 0, nil)
	}
	
	// Add header with total savings
	header := tview.NewTextView().
		SetText(fmt.Sprintf("[yellow]Optimization Opportunities - Total Potential Savings: $%s[white]", totalSavings.String())).
		SetDynamicColors(true)
	
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(list, 0, 1, false)
	
	a.content.AddItem(flex, 0, 1, false)
}

// showAllocation displays allocation efficiency
func (a *App) showAllocation() {
	a.currentView = "allocation"
	a.content.Clear()
	
	text := tview.NewTextView().
		SetText("Allocation efficiency analysis coming soon...").
		SetTextAlign(tview.AlignCenter)
	a.content.AddItem(text, 0, 1, false)
}

// showReports displays report generation options
func (a *App) showReports() {
	a.currentView = "reports"
	a.content.Clear()
	
	text := tview.NewTextView().
		SetText("Report generation options coming soon...").
		SetTextAlign(tview.AlignCenter)
	a.content.AddItem(text, 0, 1, false)
}

// showSettings displays settings
func (a *App) showSettings() {
	a.currentView = "settings"
	a.content.Clear()
	
	form := tview.NewForm().
		AddInputField("Start Date", a.dateRange.start.Format("2006-01-02"), 20, nil, nil).
		AddInputField("End Date", a.dateRange.end.Format("2006-01-02"), 20, nil, nil).
		AddButton("Apply", func() {
			// TODO: Parse and apply date changes
			a.showCostOverview()
		}).
		AddButton("Cancel", func() {
			a.showCostOverview()
		})
	
	form.SetBorder(true).SetTitle(" Settings ")
	a.content.AddItem(form, 0, 1, false)
}

// showError displays an error message
func (a *App) showError(message string) {
	a.content.Clear()
	
	errorView := tview.NewTextView().
		SetText(fmt.Sprintf("Error: %s\n\nPress any key to continue...", message)).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorRed)
	
	errorView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		a.showCostOverview()
		return nil
	})
	
	a.content.AddItem(errorView, 0, 1, false)
	a.app.SetFocus(errorView)
}
