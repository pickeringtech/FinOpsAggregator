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

	// Create nodes - Payment Cards Industry Products
	nodes := []models.CostNode{
		// Products (no direct costs - only allocated costs from dependencies)
		{
			ID:         uuid.New(),
			Name:       "card_issuing",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "card_issuing", "team": "payments", "business_unit": "issuing"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Card Issuing Platform - handles card creation, activation, and lifecycle management"},
		},
		{
			ID:         uuid.New(),
			Name:       "payment_processing",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "payment_processing", "team": "payments", "business_unit": "processing"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Payment Processing Engine - handles transaction authorization and settlement"},
		},
		{
			ID:         uuid.New(),
			Name:       "fraud_detection",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "fraud_detection", "team": "risk", "business_unit": "security"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Real-time Fraud Detection and Risk Management System"},
		},
		{
			ID:         uuid.New(),
			Name:       "merchant_onboarding",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "merchant_onboarding", "team": "merchant_services", "business_unit": "acquiring"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Merchant Onboarding and KYC Platform"},
		},

		// Platform Services (shared across products)
		{
			ID:         uuid.New(),
			Name:       "api_gateway_platform",
			Type:       string(models.NodeTypePlatform),
			CostLabels: map[string]interface{}{"environment": "prod", "region": "us-east-1", "service": "api_gateway"},
			IsPlatform: true,
			Metadata:   map[string]interface{}{"description": "Shared API Gateway and Load Balancing Platform"},
		},
		{
			ID:         uuid.New(),
			Name:       "kubernetes_platform",
			Type:       string(models.NodeTypePlatform),
			CostLabels: map[string]interface{}{"environment": "prod", "region": "us-east-1", "service": "eks"},
			IsPlatform: true,
			Metadata:   map[string]interface{}{"description": "Shared Kubernetes Platform (EKS)"},
		},

		// Dedicated Resources per Product
		{
			ID:         uuid.New(),
			Name:       "card_issuing_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "card_issuing", "workload": "compute"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Dedicated EC2 instances for Card Issuing workloads"},
		},
		{
			ID:         uuid.New(),
			Name:       "payment_processing_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "payment_processing", "workload": "compute"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "High-performance EC2 instances for Payment Processing"},
		},
		{
			ID:         uuid.New(),
			Name:       "fraud_ml_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "fraud_detection", "workload": "ml"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "GPU-enabled EC2 instances for ML-based Fraud Detection"},
		},
		{
			ID:         uuid.New(),
			Name:       "card_data_storage",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "s3", "product": "card_issuing", "workload": "storage"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Encrypted S3 storage for card data and documents"},
		},

		// Shared Infrastructure
		{
			ID:         uuid.New(),
			Name:       "payments_database_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "rds", "shared": true, "workload": "database"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared RDS PostgreSQL cluster for payment data"},
		},
		{
			ID:         uuid.New(),
			Name:       "redis_cache_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "elasticache", "shared": true, "workload": "cache"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared Redis cluster for session and transaction caching"},
		},
		{
			ID:         uuid.New(),
			Name:       "compliance_logging",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "cloudwatch", "shared": true, "workload": "logging"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Centralized compliance and audit logging infrastructure"},
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

	// Create edges - Payment Cards Industry Dependencies
	activeFrom := time.Now().AddDate(0, 0, -30) // 30 days ago
	edges := []models.DependencyEdge{
		// Card Issuing Product Dependencies
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["card_issuing_compute"],
			DefaultStrategy: string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["card_data_storage"],
			DefaultStrategy: string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["payments_database_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "db_queries",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["api_gateway_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "api_requests",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["kubernetes_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "pod_hours",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["card_issuing"],
			ChildID:         nodeMap["compliance_logging"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "log_volume",
			},
			ActiveFrom: activeFrom,
		},

		// Payment Processing Product Dependencies
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["payment_processing"],
			ChildID:         nodeMap["payment_processing_compute"],
			DefaultStrategy: string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["payment_processing"],
			ChildID:         nodeMap["payments_database_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "transaction_volume",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["payment_processing"],
			ChildID:         nodeMap["redis_cache_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "cache_operations",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["payment_processing"],
			ChildID:         nodeMap["api_gateway_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "api_requests",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["payment_processing"],
			ChildID:         nodeMap["kubernetes_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "pod_hours",
			},
			ActiveFrom: activeFrom,
		},

		// Fraud Detection Product Dependencies
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["fraud_detection"],
			ChildID:         nodeMap["fraud_ml_compute"],
			DefaultStrategy: string(models.StrategyEqual),
			DefaultParameters: map[string]interface{}{},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["fraud_detection"],
			ChildID:         nodeMap["payments_database_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "fraud_checks",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["fraud_detection"],
			ChildID:         nodeMap["redis_cache_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "ml_model_cache",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["fraud_detection"],
			ChildID:         nodeMap["kubernetes_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "ml_pod_hours",
			},
			ActiveFrom: activeFrom,
		},

		// Merchant Onboarding Product Dependencies
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["merchant_onboarding"],
			ChildID:         nodeMap["payments_database_cluster"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "kyc_checks",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["merchant_onboarding"],
			ChildID:         nodeMap["api_gateway_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "onboarding_api_calls",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["merchant_onboarding"],
			ChildID:         nodeMap["kubernetes_platform"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "pod_hours",
			},
			ActiveFrom: activeFrom,
		},
		{
			ID:              uuid.New(),
			ParentID:        nodeMap["merchant_onboarding"],
			ChildID:         nodeMap["compliance_logging"],
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "compliance_events",
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
	// Products should have NO direct costs - only allocated costs from dependencies
	switch nodeName {
	case "card_issuing", "payment_processing", "fraud_detection", "merchant_onboarding":
		return decimal.Zero // Products have no direct infrastructure costs

	// Shared Infrastructure
	case "payments_database_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(8.50) // $8.50/hour for high-performance RDS cluster
		case "storage_gb_month":
			return decimal.NewFromFloat(0.125) // $0.125/GB/month for encrypted RDS storage
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for RDS egress
		case "iops":
			return decimal.NewFromFloat(0.075) // $0.075/IOPS for provisioned IOPS
		case "backups_gb_month":
			return decimal.NewFromFloat(0.095) // $0.095/GB/month for automated backups
		case "cpu_hours":
			return decimal.NewFromFloat(0.35) // $0.35/CPU hour for high-performance
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.08) // $0.08/GB/hour memory
		case "database_connections":
			return decimal.NewFromFloat(0.002) // $0.002/connection for payment workloads
		}

	// Dedicated Compute Resources
	case "card_issuing_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.192) // $0.192/hour for EC2 m5.xlarge
		case "storage_gb_month":
			return decimal.NewFromFloat(0.10) // $0.10/GB/month for EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(0.096) // $0.096/CPU hour
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.024) // $0.024/GB/hour memory
		case "disk_io_operations":
			return decimal.NewFromFloat(0.0001) // $0.0001/IO operation
		case "snapshot_storage":
			return decimal.NewFromFloat(0.05) // $0.05/GB/month for snapshots
		}
	case "payment_processing_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.384) // $0.384/hour for EC2 c5.2xlarge (high-performance)
		case "storage_gb_month":
			return decimal.NewFromFloat(0.125) // $0.125/GB/month for high-IOPS EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(0.192) // $0.192/CPU hour for high-performance
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.048) // $0.048/GB/hour memory
		case "disk_io_operations":
			return decimal.NewFromFloat(0.0002) // $0.0002/IO operation for high-IOPS
		}
	case "fraud_ml_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(3.06) // $3.06/hour for GPU instances (p3.2xlarge)
		case "storage_gb_month":
			return decimal.NewFromFloat(0.15) // $0.15/GB/month for NVMe SSD
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(0.765) // $0.765/CPU hour for GPU instances
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.153) // $0.153/GB/hour memory for GPU instances
		}
	case "card_data_storage":
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

	// Platform Services
	case "api_gateway_platform":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.0036) // $0.0036/hour per million requests for API Gateway
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for API Gateway egress
		case "load_balancer_hours":
			return decimal.NewFromFloat(0.0225) // $0.0225/hour for ALB
		case "requests_count":
			return decimal.NewFromFloat(0.0000035) // $3.50 per million API requests
		case "data_transfer":
			return decimal.NewFromFloat(0.02) // $0.02/GB internal transfer
		}
	case "kubernetes_platform":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.10) // $0.10/hour for EKS cluster
		case "storage_gb_month":
			return decimal.NewFromFloat(0.10) // $0.10/GB/month for EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for EKS egress
		case "cpu_hours":
			return decimal.NewFromFloat(0.05) // $0.05/CPU hour for worker nodes
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.0125) // $0.0125/GB/hour memory
		case "load_balancer_hours":
			return decimal.NewFromFloat(0.0225) // $0.0225/hour for ALB
		}

	// Additional Shared Infrastructure
	case "redis_cache_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(0.017) // $0.017/hour for cache.t3.micro
		case "egress_gb":
			return decimal.NewFromFloat(0.09) // $0.09/GB for ElastiCache egress
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.0085) // $0.0085/GB/hour for cache memory
		}
	case "compliance_logging":
		switch dimension {
		case "logs_ingestion_gb":
			return decimal.NewFromFloat(0.50) // $0.50/GB for log ingestion
		case "storage_gb_month":
			return decimal.NewFromFloat(0.03) // $0.03/GB/month for log storage
		case "cloudwatch_metrics":
			return decimal.NewFromFloat(0.30) // $0.30/metric/month
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
