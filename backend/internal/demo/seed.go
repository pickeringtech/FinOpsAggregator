package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// Seeder creates demo data for testing and examples
type Seeder struct {
	store *store.Store
}

// NewSeeder creates a new demo data seeder
func NewSeeder(store *store.Store) *Seeder {
	return &Seeder{
		store: store,
	}
}

// SeedBasicDAG creates a basic DAG structure for demonstration
func (s *Seeder) SeedBasicDAG(ctx context.Context) error {
	log.Info().Msg("Seeding basic DAG structure")

	// Create nodes
	nodes := []models.CostNode{
		{
			ID:         uuid.New(),
			Name:       "product_p",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "p", "team": "alpha"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Product P - main customer-facing application"},
		},
		{
			ID:         uuid.New(),
			Name:       "product_q",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "q", "team": "beta"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Product Q - secondary application"},
		},
		{
			ID:         uuid.New(),
			Name:       "rds_shared",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "rds", "shared": true},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared RDS database cluster"},
		},
		{
			ID:         uuid.New(),
			Name:       "ec2_p",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "p"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "EC2 instances for Product P"},
		},
		{
			ID:         uuid.New(),
			Name:       "s3_p",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "s3", "product": "p"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "S3 storage for Product P"},
		},
		{
			ID:         uuid.New(),
			Name:       "platform_pool",
			Type:       string(models.NodeTypePlatform),
			CostLabels: map[string]interface{}{"platform": true},
			IsPlatform: true,
			Metadata:   map[string]interface{}{"description": "Shared platform services"},
		},
	}

	// Create nodes in database
	nodeMap := make(map[string]uuid.UUID)
	for _, node := range nodes {
		if err := s.store.Nodes.Create(ctx, &node); err != nil {
			return fmt.Errorf("failed to create node %s: %w", node.Name, err)
		}
		nodeMap[node.Name] = node.ID
		log.Debug().Str("name", node.Name).Str("id", node.ID.String()).Msg("Created node")
	}

	// Create edges
	activeFrom := time.Now().AddDate(0, 0, -30) // 30 days ago
	edges := []models.DependencyEdge{
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["product_p"],
			ChildID:         nodeMap["rds_shared"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "db_queries",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["product_q"],
			ChildID:         nodeMap["rds_shared"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "db_queries",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:                uuid.New(),
			ParentID:          nodeMap["product_p"],
			ChildID:           nodeMap["ec2_p"],
			DefaultStrategy:   string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom:        activeFrom,
		},
		{
			ID:                uuid.New(),
			ParentID:          nodeMap["product_p"],
			ChildID:           nodeMap["s3_p"],
			DefaultStrategy:   string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom:        activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["product_p"],
			ChildID:         nodeMap["platform_pool"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "requests",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["product_q"],
			ChildID:         nodeMap["platform_pool"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "requests",
			},
			ActiveFrom: activeFrom,
		},
	}

	// Create edges in database
	for _, edge := range edges {
		if err := s.store.Edges.Create(ctx, &edge); err != nil {
			return fmt.Errorf("failed to create edge %s->%s: %w",
				getNodeName(nodeMap, edge.ParentID),
				getNodeName(nodeMap, edge.ChildID), err)
		}
		log.Debug().
			Str("parent", getNodeName(nodeMap, edge.ParentID)).
			Str("child", getNodeName(nodeMap, edge.ChildID)).
			Str("strategy", edge.DefaultStrategy).
			Msg("Created edge")
	}

	log.Info().
		Int("nodes", len(nodes)).
		Int("edges", len(edges)).
		Msg("Basic DAG structure seeded successfully")

	return nil
}

// SeedCostData creates sample cost data for the last 30 days
func (s *Seeder) SeedCostData(ctx context.Context) error {
	log.Info().Msg("Seeding cost data")

	// Get all nodes
	nodes, err := s.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found - run seed basic DAG first")
	}

	// Generate costs for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	var costs []models.NodeCostByDimension
	dimensions := []string{"instance_hours", "storage_gb_month", "egress_gb", "iops", "backups_gb_month"}

	for _, node := range nodes {
		for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
			for _, dim := range dimensions {
				amount := s.generateCostAmount(node.Name, dim)
				if amount.IsZero() {
					continue // Skip zero costs
				}

				costs = append(costs, models.NodeCostByDimension{
					NodeID:    node.ID,
					CostDate:  date,
					Dimension: dim,
					Amount:    amount,
					Currency:  "USD",
					Metadata:  map[string]interface{}{"generated": true},
				})
			}
		}
	}

	// Bulk insert costs
	if err := s.store.Costs.BulkUpsert(ctx, costs); err != nil {
		return fmt.Errorf("failed to bulk insert costs: %w", err)
	}

	log.Info().Int("cost_records", len(costs)).Msg("Cost data seeded successfully")
	return nil
}

// SeedUsageData creates sample usage data for allocation calculations
func (s *Seeder) SeedUsageData(ctx context.Context) error {
	log.Info().Msg("Seeding usage data")

	// Get all nodes
	nodes, err := s.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found - run seed basic DAG first")
	}

	// Generate usage for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	var usage []models.NodeUsageByDimension
	metrics := []string{"db_queries", "requests", "cpu_hours", "memory_gb_hours"}

	for _, node := range nodes {
		for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
			for _, metric := range metrics {
				value := s.generateUsageValue(node.Name, metric)
				if value.IsZero() {
					continue // Skip zero usage
				}

				usage = append(usage, models.NodeUsageByDimension{
					NodeID:    node.ID,
					UsageDate: date,
					Metric:    metric,
					Value:     value,
					Unit:      s.getUsageUnit(metric),
				})
			}
		}
	}

	// Bulk insert usage
	if err := s.store.Usage.BulkUpsert(ctx, usage); err != nil {
		return fmt.Errorf("failed to bulk insert usage: %w", err)
	}

	log.Info().Int("usage_records", len(usage)).Msg("Usage data seeded successfully")
	return nil
}

// generateCostAmount generates realistic cost amounts based on node and dimension
func (s *Seeder) generateCostAmount(nodeName, dimension string) decimal.Decimal {
	switch nodeName {
	case "rds_shared":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(120.50) // $120.50/day for RDS instance
		case "storage_gb_month":
			return decimal.NewFromFloat(45.20) // $45.20/day for storage
		case "iops":
			return decimal.NewFromFloat(15.75) // $15.75/day for IOPS
		case "backups_gb_month":
			return decimal.NewFromFloat(8.30) // $8.30/day for backups
		}
	case "ec2_p":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(85.40) // $85.40/day for EC2
		case "egress_gb":
			return decimal.NewFromFloat(12.60) // $12.60/day for egress
		}
	case "s3_p":
		switch dimension {
		case "storage_gb_month":
			return decimal.NewFromFloat(25.80) // $25.80/day for S3 storage
		case "egress_gb":
			return decimal.NewFromFloat(18.90) // $18.90/day for S3 egress
		}
	case "platform_pool":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(200.00) // $200/day for platform
		case "egress_gb":
			return decimal.NewFromFloat(35.50) // $35.50/day for platform egress
		}
	}
	return decimal.Zero
}

// generateUsageValue generates realistic usage values
func (s *Seeder) generateUsageValue(nodeName, metric string) decimal.Decimal {
	switch nodeName {
	case "product_p":
		switch metric {
		case "db_queries":
			return decimal.NewFromInt(15000) // 15k queries/day
		case "requests":
			return decimal.NewFromInt(50000) // 50k requests/day
		}
	case "product_q":
		switch metric {
		case "db_queries":
			return decimal.NewFromInt(5000) // 5k queries/day
		case "requests":
			return decimal.NewFromInt(20000) // 20k requests/day
		}
	}
	return decimal.Zero
}

// getUsageUnit returns the appropriate unit for a metric
func (s *Seeder) getUsageUnit(metric string) string {
	switch metric {
	case "db_queries":
		return "queries"
	case "requests":
		return "requests"
	case "cpu_hours":
		return "hours"
	case "memory_gb_hours":
		return "gb_hours"
	default:
		return "units"
	}
}

// getNodeName is a helper to get node name from ID (for logging)
func getNodeName(nodeMap map[string]uuid.UUID, id uuid.UUID) string {
	for name, nodeID := range nodeMap {
		if nodeID == id {
			return name
		}
	}
	return id.String()
}
