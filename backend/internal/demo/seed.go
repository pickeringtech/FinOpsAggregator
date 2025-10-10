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

	// Create edges in database and collect edge IDs for strategies
	edgeIDs := make(map[string]uuid.UUID) // key: "parent->child"
	for _, edge := range edges {
		if err := s.store.Edges.Create(ctx, &edge); err != nil {
			return fmt.Errorf("failed to create edge %s->%s: %w",
				getNodeName(nodeMap, edge.ParentID),
				getNodeName(nodeMap, edge.ChildID), err)
		}

		// Store edge ID for strategy creation
		key := fmt.Sprintf("%s->%s", getNodeName(nodeMap, edge.ParentID), getNodeName(nodeMap, edge.ChildID))
		edgeIDs[key] = edge.ID

		log.Debug().
			Str("parent", getNodeName(nodeMap, edge.ParentID)).
			Str("child", getNodeName(nodeMap, edge.ChildID)).
			Str("strategy", edge.DefaultStrategy).
			Msg("Created edge")
	}

	// Create dimension-specific edge strategies
	if err := s.seedEdgeStrategies(ctx, edgeIDs); err != nil {
		return fmt.Errorf("failed to seed edge strategies: %w", err)
	}

	log.Info().
		Int("nodes", len(nodes)).
		Int("edges", len(edges)).
		Msg("Basic DAG structure seeded successfully")

	return nil
}

// SeedCostData creates sample cost data for a large dataset (100,000+ records)
func (s *Seeder) SeedCostData(ctx context.Context) error {
	log.Info().Msg("Seeding large-scale cost data")

	// Get all nodes
	nodes, err := s.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found - run seed basic DAG first")
	}

	// Generate costs for the last 6 months for comprehensive analysis (targeting ~100k records)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -6, 0) // 6 months of data

	var costs []models.NodeCostByDimension

	// Expanded dimensions for more realistic FinOps scenarios
	dimensions := []string{
		"instance_hours", "storage_gb_month", "egress_gb", "iops", "backups_gb_month",
		"cpu_hours", "memory_gb_hours", "network_gb", "requests_count", "lambda_invocations",
		"database_connections", "cache_hits", "cdn_requests", "api_calls", "data_transfer",
		"disk_io_operations", "snapshot_storage", "reserved_instances", "spot_instances",
		"load_balancer_hours", "nat_gateway_hours", "vpn_hours", "cloudwatch_metrics",
		"logs_ingestion_gb", "monitoring_checks",
	}

	// Generate costs with multiple granularities and variations
	log.Info().Msg("Generating comprehensive cost dataset...")

	totalRecords := 0
	batchSize := 10000 // Process in batches to avoid memory issues

	for _, node := range nodes {
		log.Info().Str("node", node.Name).Msg("Processing node")

		// Generate multiple service instances per node for more realistic data
		serviceCount := s.getServiceCountForNode(node.Name)

		for serviceIdx := 0; serviceIdx < serviceCount; serviceIdx++ {
			for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
				for _, dim := range dimensions {
					// Generate multiple records per day for high-frequency dimensions
					recordsPerDay := s.getRecordsPerDay(dim)

					for recordIdx := 0; recordIdx < recordsPerDay; recordIdx++ {
						amount := s.generateCostAmount(node.Name, dim, serviceIdx, recordIdx, date)
						if amount.IsZero() {
							continue // Skip zero costs
						}

						// Add time variation for sub-daily records
						recordTime := date.Add(time.Duration(recordIdx) * time.Hour * 24 / time.Duration(recordsPerDay))

						// Make dimension unique by including service and record identifiers
						uniqueDimension := dim
						if serviceCount > 1 || recordsPerDay > 1 {
							uniqueDimension = fmt.Sprintf("%s_s%d_r%d", dim, serviceIdx, recordIdx)
						}

						costs = append(costs, models.NodeCostByDimension{
							NodeID:    node.ID,
							CostDate:  recordTime,
							Dimension: uniqueDimension,
							Amount:    amount,
							Currency:  "USD",
							Metadata: map[string]interface{}{
								"generated":      true,
								"base_dimension": dim,
								"service_index":  serviceIdx,
								"record_index":   recordIdx,
								"granularity":    s.getGranularity(dim),
							},
						})

						totalRecords++

						// Batch insert to avoid memory issues
						if len(costs) >= batchSize {
							if err := s.store.Costs.BulkUpsert(ctx, costs); err != nil {
								return fmt.Errorf("failed to bulk insert costs batch: %w", err)
							}
							log.Info().Int("records_inserted", len(costs)).Int("total_so_far", totalRecords).Msg("Batch inserted")
							costs = costs[:0] // Reset slice
						}
					}
				}
			}
		}
	}

	// Insert remaining costs
	if len(costs) > 0 {
		if err := s.store.Costs.BulkUpsert(ctx, costs); err != nil {
			return fmt.Errorf("failed to bulk insert final costs: %w", err)
		}
	}

	log.Info().Int("cost_records", totalRecords).Msg("Large-scale cost data seeded successfully")
	return nil
}

// SeedUsageData creates sample usage data for allocation calculations (large scale)
func (s *Seeder) SeedUsageData(ctx context.Context) error {
	log.Info().Msg("Seeding large-scale usage data")

	// Get all nodes
	nodes, err := s.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found - run seed basic DAG first")
	}

	// Generate usage for the last 6 months to match cost data
	endDate := time.Now()
	startDate := endDate.AddDate(0, -6, 0)

	var usage []models.NodeUsageByDimension
	metrics := []string{
		"db_queries", "requests", "cpu_hours", "memory_gb_hours", "storage_operations",
		"network_packets", "cache_hits", "api_requests", "lambda_executions", "disk_reads",
		"disk_writes", "connections", "transactions", "backup_operations", "log_entries",
	}

	totalRecords := 0
	batchSize := 5000

	for _, node := range nodes {
		log.Info().Str("node", node.Name).Msg("Processing usage for node")

		// Generate multiple service instances for usage data too
		serviceCount := s.getServiceCountForNode(node.Name)

		for serviceIdx := 0; serviceIdx < serviceCount; serviceIdx++ {
			for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
				for _, metric := range metrics {
					// Generate hourly usage for high-frequency metrics
					recordsPerDay := s.getUsageRecordsPerDay(metric)

					for recordIdx := 0; recordIdx < recordsPerDay; recordIdx++ {
						value := s.generateUsageValue(node.Name, metric, serviceIdx, recordIdx, date)
						if value.IsZero() {
							continue // Skip zero usage
						}

						recordTime := date.Add(time.Duration(recordIdx) * time.Hour * 24 / time.Duration(recordsPerDay))

						// Make metric unique by including service and record identifiers
						uniqueMetric := metric
						if serviceCount > 1 || recordsPerDay > 1 {
							uniqueMetric = fmt.Sprintf("%s_s%d_r%d", metric, serviceIdx, recordIdx)
						}

						usage = append(usage, models.NodeUsageByDimension{
							NodeID:    node.ID,
							UsageDate: recordTime,
							Metric:    uniqueMetric,
							Value:     value,
							Unit:      s.getUsageUnit(metric),
						})

						totalRecords++

						// Batch insert to avoid memory issues
						if len(usage) >= batchSize {
							if err := s.store.Usage.BulkUpsert(ctx, usage); err != nil {
								return fmt.Errorf("failed to bulk insert usage batch: %w", err)
							}
							log.Info().Int("records_inserted", len(usage)).Int("total_so_far", totalRecords).Msg("Usage batch inserted")
							usage = usage[:0] // Reset slice
						}
					}
				}
			}
		}
	}

	// Insert remaining usage
	if len(usage) > 0 {
		if err := s.store.Usage.BulkUpsert(ctx, usage); err != nil {
			return fmt.Errorf("failed to bulk insert final usage: %w", err)
		}
	}

	log.Info().Int("usage_records", totalRecords).Msg("Large-scale usage data seeded successfully")
	return nil
}

// generateCostAmount generates realistic cost amounts based on node, dimension, and variations
func (s *Seeder) generateCostAmount(nodeName, dimension string, serviceIdx, recordIdx int, date time.Time) decimal.Decimal {
	// Add variation based on service index, time, and other factors
	baseAmount := s.getBaseCostAmount(nodeName, dimension)
	if baseAmount.IsZero() {
		return decimal.Zero
	}

	// Add realistic variations
	variation := s.calculateCostVariation(nodeName, dimension, serviceIdx, recordIdx, date)
	return baseAmount.Mul(variation)
}

// getBaseCostAmount returns base cost amounts for different node/dimension combinations
func (s *Seeder) getBaseCostAmount(nodeName, dimension string) decimal.Decimal {
	switch nodeName {
	case "rds_shared":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(5.02) // $5.02/hour for RDS instance
		case "storage_gb_month":
			return decimal.NewFromFloat(0.115) // $0.115/GB/month for RDS storage
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for RDS egress
		case "iops":
			return decimal.NewFromFloat(0.065) // $0.065/IOPS for RDS
		case "backups_gb_month":
			return decimal.NewFromFloat(0.095) // $0.095/GB/month for RDS backups
		case "cpu_hours":
			return decimal.NewFromFloat(0.25) // $0.25/CPU hour
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.05) // $0.05/GB/hour memory
		case "database_connections":
			return decimal.NewFromFloat(0.001) // $0.001/connection
		}
	case "ec2_p":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.096) // $0.096/hour for EC2 m5.large
		case "storage_gb_month":
			return decimal.NewFromFloat(0.10) // $0.10/GB/month for EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(0.048) // $0.048/CPU hour
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.012) // $0.012/GB/hour memory
		case "disk_io_operations":
			return decimal.NewFromFloat(0.0001) // $0.0001/IO operation
		case "snapshot_storage":
			return decimal.NewFromFloat(0.05) // $0.05/GB/month for snapshots
		case "load_balancer_hours":
			return decimal.NewFromFloat(0.0225) // $0.0225/hour for ALB
		}
	case "s3_p":
		switch dimension {
		case "storage_gb_month":
			return decimal.NewFromFloat(0.023) // $0.023/GB/month for S3 Standard
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for S3 egress
		case "requests_count":
			return decimal.NewFromFloat(0.0004) // $0.0004/1000 requests
		case "data_transfer":
			return decimal.NewFromFloat(0.02) // $0.02/GB internal transfer
		}
	case "platform_pool":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.192) // $0.192/hour for larger instances
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for platform egress
		case "load_balancer_hours":
			return decimal.NewFromFloat(0.0225) // $0.0225/hour for ALB
		case "nat_gateway_hours":
			return decimal.NewFromFloat(0.045) // $0.045/hour for NAT Gateway
		case "vpn_hours":
			return decimal.NewFromFloat(0.05) // $0.05/hour for VPN
		case "cloudwatch_metrics":
			return decimal.NewFromFloat(0.30) // $0.30/metric/month
		case "logs_ingestion_gb":
			return decimal.NewFromFloat(0.50) // $0.50/GB for log ingestion
		}
	case "product_p", "product_q":
		// Product nodes get allocated costs, but also have some direct costs
		switch dimension {
		case "lambda_invocations":
			return decimal.NewFromFloat(0.0000002) // $0.0000002/invocation
		case "api_calls":
			return decimal.NewFromFloat(0.000003) // $0.000003/API call
		case "cdn_requests":
			return decimal.NewFromFloat(0.0001) // $0.0001/10000 requests
		case "monitoring_checks":
			return decimal.NewFromFloat(0.001) // $0.001/check
		}
	}
	return decimal.Zero
}

// getServiceCountForNode returns the number of service instances per node
func (s *Seeder) getServiceCountForNode(nodeName string) int {
	switch nodeName {
	case "rds_shared":
		return 2 // 2 RDS instances
	case "ec2_p":
		return 4 // 4 EC2 instances
	case "s3_p":
		return 3 // 3 S3 buckets/services
	case "platform_pool":
		return 5 // 5 platform services
	case "product_p":
		return 3 // 3 product services
	case "product_q":
		return 2 // 2 product services
	default:
		return 1
	}
}

// getRecordsPerDay returns how many records to generate per day for a dimension
func (s *Seeder) getRecordsPerDay(dimension string) int {
	switch dimension {
	case "instance_hours", "cpu_hours", "memory_gb_hours":
		return 24 // Hourly records
	case "requests_count", "lambda_invocations", "api_calls":
		return 24 // Hourly aggregation
	case "disk_io_operations", "database_connections":
		return 12 // Every 2 hours
	case "load_balancer_hours", "nat_gateway_hours", "vpn_hours":
		return 24 // Hourly
	case "cloudwatch_metrics", "monitoring_checks":
		return 4 // Every 6 hours
	default:
		return 1 // Daily records for storage, egress, etc.
	}
}

// getGranularity returns the granularity level for a dimension
func (s *Seeder) getGranularity(dimension string) string {
	switch dimension {
	case "instance_hours", "cpu_hours", "memory_gb_hours", "load_balancer_hours", "nat_gateway_hours", "vpn_hours":
		return "hourly"
	case "requests_count", "lambda_invocations", "api_calls", "disk_io_operations", "database_connections":
		return "hourly_aggregated"
	case "cloudwatch_metrics", "monitoring_checks":
		return "6_hourly"
	default:
		return "daily"
	}
}

// calculateCostVariation adds realistic variations to base costs
func (s *Seeder) calculateCostVariation(nodeName, dimension string, serviceIdx, recordIdx int, date time.Time) decimal.Decimal {
	variation := decimal.NewFromFloat(1.0) // Start with base multiplier

	// Add service-specific variation (different instance sizes, etc.)
	serviceVariation := 1.0 + (float64(serviceIdx%5)-2)*0.15 // ±30% variation across services
	variation = variation.Mul(decimal.NewFromFloat(serviceVariation))

	// Add time-based variations
	hour := date.Hour()
	dayOfWeek := date.Weekday()

	// Business hours effect (higher usage during business hours)
	if hour >= 9 && hour <= 17 && dayOfWeek >= time.Monday && dayOfWeek <= time.Friday {
		variation = variation.Mul(decimal.NewFromFloat(1.3)) // 30% higher during business hours
	} else if hour >= 22 || hour <= 6 {
		variation = variation.Mul(decimal.NewFromFloat(0.7)) // 30% lower during night hours
	}

	// Weekend effect
	if dayOfWeek == time.Saturday || dayOfWeek == time.Sunday {
		variation = variation.Mul(decimal.NewFromFloat(0.6)) // 40% lower on weekends
	}

	// Monthly growth trend (simulate business growth)
	monthsSinceStart := float64(date.Sub(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)).Hours()) / (24 * 30)
	growthFactor := 1.0 + (monthsSinceStart * 0.02) // 2% growth per month
	variation = variation.Mul(decimal.NewFromFloat(growthFactor))

	// Add some random variation (±10%)
	randomFactor := 0.9 + (float64((serviceIdx+recordIdx+date.Day())%20) / 100.0) // 0.9 to 1.1
	variation = variation.Mul(decimal.NewFromFloat(randomFactor))

	// Dimension-specific variations
	switch dimension {
	case "egress_gb", "data_transfer":
		// Higher egress during business hours and month-end
		if date.Day() > 25 {
			variation = variation.Mul(decimal.NewFromFloat(1.4)) // 40% higher at month-end
		}
	case "requests_count", "api_calls", "lambda_invocations":
		// Much higher variation for request-based metrics
		requestVariation := 0.5 + (float64((serviceIdx*recordIdx+date.Hour())%100) / 100.0) // 0.5 to 1.5
		variation = variation.Mul(decimal.NewFromFloat(requestVariation))
	case "storage_gb_month":
		// Storage grows more steadily
		variation = variation.Mul(decimal.NewFromFloat(0.95 + monthsSinceStart*0.01)) // Steady growth
	}

	// Ensure variation is reasonable (between 0.1 and 3.0)
	if variation.LessThan(decimal.NewFromFloat(0.1)) {
		variation = decimal.NewFromFloat(0.1)
	}
	if variation.GreaterThan(decimal.NewFromFloat(3.0)) {
		variation = decimal.NewFromFloat(3.0)
	}

	return variation
}

// getUsageRecordsPerDay returns how many usage records to generate per day for a metric
func (s *Seeder) getUsageRecordsPerDay(metric string) int {
	switch metric {
	case "cpu_hours", "memory_gb_hours", "connections", "api_requests":
		return 24 // Hourly records
	case "requests", "lambda_executions", "cache_hits", "network_packets":
		return 12 // Every 2 hours
	case "disk_reads", "disk_writes", "storage_operations", "log_entries":
		return 6 // Every 4 hours
	default:
		return 1 // Daily records
	}
}

// generateUsageValue generates realistic usage values with variations
func (s *Seeder) generateUsageValue(nodeName, metric string, serviceIdx, recordIdx int, date time.Time) decimal.Decimal {
	baseValue := s.getBaseUsageValue(nodeName, metric)
	if baseValue.IsZero() {
		return decimal.Zero
	}

	// Add realistic variations similar to cost data
	variation := s.calculateUsageVariation(nodeName, metric, serviceIdx, recordIdx, date)
	return baseValue.Mul(variation)
}

// getBaseUsageValue returns base usage values for different node/metric combinations
func (s *Seeder) getBaseUsageValue(nodeName, metric string) decimal.Decimal {
	switch nodeName {
	case "product_p":
		switch metric {
		case "db_queries":
			return decimal.NewFromInt(625) // 625 queries/hour (15k/day)
		case "requests":
			return decimal.NewFromInt(2083) // 2083 requests/hour (50k/day)
		case "api_requests":
			return decimal.NewFromInt(1250) // 1250 API requests/hour
		case "lambda_executions":
			return decimal.NewFromInt(500) // 500 lambda executions/hour
		case "cache_hits":
			return decimal.NewFromInt(8333) // 8333 cache hits/hour
		}
	case "product_q":
		switch metric {
		case "db_queries":
			return decimal.NewFromInt(208) // 208 queries/hour (5k/day)
		case "requests":
			return decimal.NewFromInt(833) // 833 requests/hour (20k/day)
		case "api_requests":
			return decimal.NewFromInt(417) // 417 API requests/hour
		case "lambda_executions":
			return decimal.NewFromInt(167) // 167 lambda executions/hour
		case "cache_hits":
			return decimal.NewFromInt(2778) // 2778 cache hits/hour
		}
	case "rds_shared":
		switch metric {
		case "db_queries":
			return decimal.NewFromInt(1042) // 1042 queries/hour
		case "connections":
			return decimal.NewFromInt(45) // 45 connections/hour
		case "transactions":
			return decimal.NewFromInt(2500) // 2500 transactions/hour
		case "disk_reads":
			return decimal.NewFromInt(15000) // 15k disk reads/hour
		case "disk_writes":
			return decimal.NewFromInt(8000) // 8k disk writes/hour
		case "backup_operations":
			return decimal.NewFromInt(2) // 2 backup operations/hour
		}
	case "ec2_p":
		switch metric {
		case "cpu_hours":
			return decimal.NewFromFloat(0.75) // 75% CPU utilization
		case "memory_gb_hours":
			return decimal.NewFromFloat(6.2) // 6.2 GB memory usage/hour
		case "network_packets":
			return decimal.NewFromInt(125000) // 125k packets/hour
		case "disk_reads":
			return decimal.NewFromInt(5000) // 5k disk reads/hour
		case "disk_writes":
			return decimal.NewFromInt(2000) // 2k disk writes/hour
		case "storage_operations":
			return decimal.NewFromInt(1500) // 1500 storage ops/hour
		}
	case "s3_p":
		switch metric {
		case "requests":
			return decimal.NewFromInt(8333) // 8333 S3 requests/hour
		case "storage_operations":
			return decimal.NewFromInt(2083) // 2083 storage ops/hour
		case "network_packets":
			return decimal.NewFromInt(41667) // 41667 packets/hour
		}
	case "platform_pool":
		switch metric {
		case "cpu_hours":
			return decimal.NewFromFloat(1.2) // 120% CPU utilization (multiple cores)
		case "memory_gb_hours":
			return decimal.NewFromFloat(12.8) // 12.8 GB memory usage/hour
		case "connections":
			return decimal.NewFromInt(150) // 150 connections/hour
		case "log_entries":
			return decimal.NewFromInt(50000) // 50k log entries/hour
		case "network_packets":
			return decimal.NewFromInt(200000) // 200k packets/hour
		}
	}
	return decimal.Zero
}

// calculateUsageVariation adds realistic variations to base usage values
func (s *Seeder) calculateUsageVariation(nodeName, metric string, serviceIdx, recordIdx int, date time.Time) decimal.Decimal {
	// Similar to cost variation but with different patterns for usage
	variation := decimal.NewFromFloat(1.0)

	// Service-specific variation
	serviceVariation := 1.0 + (float64(serviceIdx%3)-1)*0.2 // ±20% variation across services
	variation = variation.Mul(decimal.NewFromFloat(serviceVariation))

	// Time-based variations (usage patterns differ from cost patterns)
	hour := date.Hour()
	dayOfWeek := date.Weekday()

	// Business hours have higher usage
	if hour >= 8 && hour <= 18 && dayOfWeek >= time.Monday && dayOfWeek <= time.Friday {
		variation = variation.Mul(decimal.NewFromFloat(1.5)) // 50% higher during business hours
	} else if hour >= 23 || hour <= 5 {
		variation = variation.Mul(decimal.NewFromFloat(0.3)) // 70% lower during night hours
	}

	// Weekend effect (less usage)
	if dayOfWeek == time.Saturday || dayOfWeek == time.Sunday {
		variation = variation.Mul(decimal.NewFromFloat(0.4)) // 60% lower on weekends
	}

	// Growth trend
	monthsSinceStart := float64(date.Sub(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)).Hours()) / (24 * 30)
	growthFactor := 1.0 + (monthsSinceStart * 0.03) // 3% growth per month for usage
	variation = variation.Mul(decimal.NewFromFloat(growthFactor))

	// Random variation
	randomFactor := 0.8 + (float64((serviceIdx+recordIdx+date.Hour())%40) / 100.0) // 0.8 to 1.2
	variation = variation.Mul(decimal.NewFromFloat(randomFactor))

	// Metric-specific variations
	switch metric {
	case "requests", "api_requests", "lambda_executions":
		// High variability for request-based metrics
		requestVariation := 0.3 + (float64((serviceIdx*recordIdx+date.Minute())%140) / 100.0) // 0.3 to 1.7
		variation = variation.Mul(decimal.NewFromFloat(requestVariation))
	case "cpu_hours", "memory_gb_hours":
		// More stable resource usage
		variation = variation.Mul(decimal.NewFromFloat(0.9 + (float64(date.Hour()%10) / 50.0))) // 0.9 to 1.1
	}

	// Ensure reasonable bounds
	if variation.LessThan(decimal.NewFromFloat(0.05)) {
		variation = decimal.NewFromFloat(0.05)
	}
	if variation.GreaterThan(decimal.NewFromFloat(5.0)) {
		variation = decimal.NewFromFloat(5.0)
	}

	return variation
}

// getUsageUnit returns the appropriate unit for a metric
func (s *Seeder) getUsageUnit(metric string) string {
	switch metric {
	case "db_queries":
		return "queries"
	case "requests", "api_requests":
		return "requests"
	case "cpu_hours":
		return "hours"
	case "memory_gb_hours":
		return "gb_hours"
	case "lambda_executions":
		return "executions"
	case "connections":
		return "connections"
	case "transactions":
		return "transactions"
	case "disk_reads", "disk_writes":
		return "operations"
	case "storage_operations":
		return "operations"
	case "network_packets":
		return "packets"
	case "cache_hits":
		return "hits"
	case "backup_operations":
		return "operations"
	case "log_entries":
		return "entries"
	default:
		return "units"
	}
}

// seedEdgeStrategies creates dimension-specific allocation strategies for edges
func (s *Seeder) seedEdgeStrategies(ctx context.Context, edgeIDs map[string]uuid.UUID) error {
	log.Info().Msg("Seeding edge strategies")

	// Define dimension-specific strategies for different edge types
	strategies := []struct {
		edgeKey   string
		dimension string
		strategy  string
		params    map[string]interface{}
	}{
		// RDS shared strategies - database costs should be proportional to queries
		{"product_p->rds_shared", "instance_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "database_connections"}},
		{"product_q->rds_shared", "instance_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "database_connections"}},
		{"product_p->rds_shared", "storage_gb_month", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "database_connections"}},
		{"product_q->rds_shared", "storage_gb_month", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "database_connections"}},

		// Platform pool strategies - compute costs should be proportional to requests
		{"product_p->platform_pool", "instance_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "requests_count"}},
		{"product_q->platform_pool", "instance_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "requests_count"}},
		{"product_p->platform_pool", "cpu_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "requests_count"}},
		{"product_q->platform_pool", "cpu_hours", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "requests_count"}},

		// EC2 strategies - infrastructure costs should be equal split for shared resources
		{"platform_pool->ec2_p", "instance_hours", string(models.StrategyEqual), map[string]interface{}{}},
		{"rds_shared->ec2_p", "instance_hours", string(models.StrategyEqual), map[string]interface{}{}},
		{"s3_p->ec2_p", "egress_gb", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "data_transfer"}},

		// S3 strategies - storage costs proportional to usage
		{"platform_pool->s3_p", "storage_gb_month", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "requests_count"}},
		{"rds_shared->s3_p", "storage_gb_month", string(models.StrategyProportionalOn), map[string]interface{}{"metric": "backups_gb_month"}},
	}

	for _, strat := range strategies {
		edgeID, exists := edgeIDs[strat.edgeKey]
		if !exists {
			log.Warn().Str("edge_key", strat.edgeKey).Msg("Edge not found for strategy")
			continue
		}

		strategy := models.EdgeStrategy{
			ID:         uuid.New(),
			EdgeID:     edgeID,
			Dimension:  &strat.dimension,
			Strategy:   strat.strategy,
			Parameters: strat.params,
		}

		if err := s.store.Edges.CreateStrategy(ctx, &strategy); err != nil {
			return fmt.Errorf("failed to create edge strategy for %s dimension %s: %w", strat.edgeKey, strat.dimension, err)
		}

		log.Debug().
			Str("edge", strat.edgeKey).
			Str("dimension", strat.dimension).
			Str("strategy", strat.strategy).
			Msg("Created edge strategy")
	}

	log.Info().Int("strategies", len(strategies)).Msg("Edge strategies seeded successfully")
	return nil
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
