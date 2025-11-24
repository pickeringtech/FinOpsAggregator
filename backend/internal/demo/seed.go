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

	// Create nodes - 25 Products across multiple business units
	nodes := []models.CostNode{
		// Payment Products
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
		{
			ID:         uuid.New(),
			Name:       "payment_gateway",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "payment_gateway", "team": "payments", "business_unit": "processing"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Payment Gateway API for merchant integrations"},
		},
		{
			ID:         uuid.New(),
			Name:       "recurring_billing",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "recurring_billing", "team": "payments", "business_unit": "subscriptions"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Subscription and recurring billing management"},
		},
		{
			ID:         uuid.New(),
			Name:       "dispute_management",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "dispute_management", "team": "risk", "business_unit": "operations"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Chargeback and dispute resolution platform"},
		},

		// Banking Products
		{
			ID:         uuid.New(),
			Name:       "digital_banking",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "digital_banking", "team": "banking", "business_unit": "retail"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Digital banking platform with account management"},
		},
		{
			ID:         uuid.New(),
			Name:       "loan_origination",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "loan_origination", "team": "lending", "business_unit": "credit"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Automated loan origination and underwriting"},
		},
		{
			ID:         uuid.New(),
			Name:       "wealth_management",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "wealth_management", "team": "investment", "business_unit": "wealth"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Investment portfolio and wealth management platform"},
		},

		// Analytics & Data Products
		{
			ID:         uuid.New(),
			Name:       "business_intelligence",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "business_intelligence", "team": "analytics", "business_unit": "data"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Business intelligence and reporting platform"},
		},
		{
			ID:         uuid.New(),
			Name:       "data_warehouse",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "data_warehouse", "team": "data_engineering", "business_unit": "data"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Enterprise data warehouse and ETL pipelines"},
		},
		{
			ID:         uuid.New(),
			Name:       "ml_platform",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "ml_platform", "team": "ml_engineering", "business_unit": "ai"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Machine learning model training and serving platform"},
		},

		// Customer Experience Products
		{
			ID:         uuid.New(),
			Name:       "customer_portal",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "customer_portal", "team": "customer_experience", "business_unit": "cx"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Self-service customer portal and dashboard"},
		},
		{
			ID:         uuid.New(),
			Name:       "mobile_app",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "mobile_app", "team": "mobile", "business_unit": "cx"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Native mobile applications for iOS and Android"},
		},
		{
			ID:         uuid.New(),
			Name:       "notification_service",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "notification_service", "team": "communications", "business_unit": "cx"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Multi-channel notification and messaging service"},
		},

		// Compliance & Security Products
		{
			ID:         uuid.New(),
			Name:       "kyc_verification",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "kyc_verification", "team": "compliance", "business_unit": "security"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Know Your Customer verification and identity checks"},
		},
		{
			ID:         uuid.New(),
			Name:       "aml_monitoring",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "aml_monitoring", "team": "compliance", "business_unit": "security"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Anti-Money Laundering transaction monitoring"},
		},
		{
			ID:         uuid.New(),
			Name:       "audit_logging",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "audit_logging", "team": "security", "business_unit": "security"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Centralized audit logging and compliance reporting"},
		},

		// Integration & API Products
		{
			ID:         uuid.New(),
			Name:       "api_marketplace",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "api_marketplace", "team": "platform", "business_unit": "developer"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "API marketplace and developer portal"},
		},
		{
			ID:         uuid.New(),
			Name:       "webhook_service",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "webhook_service", "team": "platform", "business_unit": "developer"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Webhook delivery and event streaming service"},
		},
		{
			ID:         uuid.New(),
			Name:       "partner_integrations",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "partner_integrations", "team": "partnerships", "business_unit": "business_dev"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Third-party partner integration platform"},
		},

		// Operations Products
		{
			ID:         uuid.New(),
			Name:       "admin_console",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "admin_console", "team": "operations", "business_unit": "ops"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Internal admin and operations console"},
		},
		{
			ID:         uuid.New(),
			Name:       "support_ticketing",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "support_ticketing", "team": "support", "business_unit": "ops"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Customer support ticketing and case management"},
		},
		{
			ID:         uuid.New(),
			Name:       "reporting_engine",
			Type:       string(models.NodeTypeProduct),
			CostLabels: map[string]interface{}{"product": "reporting_engine", "team": "analytics", "business_unit": "ops"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Automated reporting and dashboard generation"},
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
		{
			ID:         uuid.New(),
			Name:       "cdn_platform",
			Type:       string(models.NodeTypePlatform),
			CostLabels: map[string]interface{}{"environment": "prod", "region": "global", "service": "cloudfront"},
			IsPlatform: true,
			Metadata:   map[string]interface{}{"description": "Global CDN for static assets and API acceleration"},
		},
		{
			ID:         uuid.New(),
			Name:       "monitoring_platform",
			Type:       string(models.NodeTypePlatform),
			CostLabels: map[string]interface{}{"environment": "prod", "region": "us-east-1", "service": "datadog"},
			IsPlatform: true,
			Metadata:   map[string]interface{}{"description": "Centralized monitoring and observability platform"},
		},

		// Dedicated Resources - High-cost compute for major products
		{
			ID:         uuid.New(),
			Name:       "card_issuing_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "card_issuing", "workload": "compute"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Dedicated EC2 instances for Card Issuing workloads", "instance_type": "c5.4xlarge", "count": 20},
		},
		{
			ID:         uuid.New(),
			Name:       "payment_processing_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "payment_processing", "workload": "compute"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "High-performance EC2 instances for Payment Processing", "instance_type": "c5.9xlarge", "count": 50},
		},
		{
			ID:         uuid.New(),
			Name:       "fraud_ml_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "fraud_detection", "workload": "ml"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "GPU-enabled EC2 instances for ML-based Fraud Detection", "instance_type": "p3.8xlarge", "count": 15},
		},
		{
			ID:         uuid.New(),
			Name:       "data_warehouse_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "redshift", "product": "data_warehouse", "workload": "analytics"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Redshift cluster for data warehousing", "node_type": "ra3.4xlarge", "count": 30},
		},
		{
			ID:         uuid.New(),
			Name:       "ml_platform_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "sagemaker", "product": "ml_platform", "workload": "ml_training"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "SageMaker instances for ML training", "instance_type": "ml.p3.16xlarge", "count": 10},
		},
		{
			ID:         uuid.New(),
			Name:       "digital_banking_compute",
			Type:       string(models.NodeTypeResource),
			CostLabels: map[string]interface{}{"service": "ec2", "product": "digital_banking", "workload": "compute"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "EC2 instances for digital banking platform", "instance_type": "m5.2xlarge", "count": 25},
		},

		// Shared Infrastructure - High-cost shared services
		{
			ID:         uuid.New(),
			Name:       "payments_database_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "rds", "shared": true, "workload": "database"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared RDS PostgreSQL cluster for payment data", "instance_type": "db.r5.12xlarge", "count": 8},
		},
		{
			ID:         uuid.New(),
			Name:       "analytics_database_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "rds", "shared": true, "workload": "analytics"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared RDS cluster for analytics workloads", "instance_type": "db.r5.8xlarge", "count": 5},
		},
		{
			ID:         uuid.New(),
			Name:       "redis_cache_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "elasticache", "shared": true, "workload": "cache"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared Redis cluster for session and transaction caching", "node_type": "cache.r5.4xlarge", "count": 12},
		},
		{
			ID:         uuid.New(),
			Name:       "message_queue_cluster",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "kafka", "shared": true, "workload": "messaging"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Kafka cluster for event streaming", "instance_type": "kafka.m5.4xlarge", "count": 9},
		},
		{
			ID:         uuid.New(),
			Name:       "object_storage",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "s3", "shared": true, "workload": "storage"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Shared S3 storage for documents and backups", "storage_tb": 500},
		},
		{
			ID:         uuid.New(),
			Name:       "compliance_logging",
			Type:       string(models.NodeTypeShared),
			CostLabels: map[string]interface{}{"service": "cloudwatch", "shared": true, "workload": "logging"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Centralized compliance and audit logging infrastructure", "log_ingestion_gb_day": 5000},
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

	// Create edges programmatically - All products connect to shared platform/infrastructure
	activeFrom := time.Now().AddDate(0, 0, -60) // 60 days ago to ensure edges are active for all demo data
	edges := []models.DependencyEdge{}

	// Define product-to-resource mappings (dedicated resources)
	productResources := map[string][]string{
		"card_issuing":         {"card_issuing_compute"},
		"payment_processing":   {"payment_processing_compute"},
		"fraud_detection":      {"fraud_ml_compute"},
		"data_warehouse":       {"data_warehouse_compute"},
		"ml_platform":          {"ml_platform_compute"},
		"digital_banking":      {"digital_banking_compute"},
	}

	// All products that use shared infrastructure
	allProducts := []string{
		"card_issuing", "payment_processing", "fraud_detection", "merchant_onboarding",
		"payment_gateway", "recurring_billing", "dispute_management", "digital_banking",
		"loan_origination", "wealth_management", "business_intelligence", "data_warehouse",
		"ml_platform", "customer_portal", "mobile_app", "notification_service",
		"kyc_verification", "aml_monitoring", "audit_logging", "api_marketplace",
		"webhook_service", "partner_integrations", "admin_console", "support_ticketing",
		"reporting_engine",
	}

	// Create edges from products to their dedicated resources
	for product, resources := range productResources {
		for _, resource := range resources {
			if _, ok := nodeMap[product]; !ok {
				continue
			}
			if _, ok := nodeMap[resource]; !ok {
				continue
			}
			edges = append(edges, models.DependencyEdge{
				ID:              uuid.New(),
				ParentID:        nodeMap[product],
				ChildID:         nodeMap[resource],
				DefaultStrategy: string(models.StrategyEqual),
				DefaultParameters: map[string]interface{}{},
				ActiveFrom: activeFrom,
			})
		}
	}

	// Create edges from all products to shared platform services
	sharedPlatform := []string{"api_gateway_platform", "kubernetes_platform", "cdn_platform", "monitoring_platform"}
	for _, product := range allProducts {
		if _, ok := nodeMap[product]; !ok {
			continue
		}
		for _, platform := range sharedPlatform {
			if _, ok := nodeMap[platform]; !ok {
				continue
			}
			edges = append(edges, models.DependencyEdge{
				ID:              uuid.New(),
				ParentID:        nodeMap[product],
				ChildID:         nodeMap[platform],
				DefaultStrategy: string(models.StrategyProportionalOn),
				DefaultParameters: map[string]interface{}{
					"metric": "api_requests",
				},
				ActiveFrom: activeFrom,
			})
		}
	}

	// Create edges from all products to shared infrastructure
	sharedInfra := []string{
		"payments_database_cluster", "analytics_database_cluster", "redis_cache_cluster",
		"message_queue_cluster", "object_storage", "compliance_logging",
	}
	for _, product := range allProducts {
		if _, ok := nodeMap[product]; !ok {
			continue
		}
		for _, infra := range sharedInfra {
			if _, ok := nodeMap[infra]; !ok {
				continue
			}
			edges = append(edges, models.DependencyEdge{
				ID:              uuid.New(),
				ParentID:        nodeMap[product],
				ChildID:         nodeMap[infra],
				DefaultStrategy: string(models.StrategyProportionalOn),
				DefaultParameters: map[string]interface{}{
					"metric": "usage_metric",
				},
				ActiveFrom: activeFrom,
			})
		}
	}

	// Add product-to-product dependency: Payment Processing calls Fraud Detection
	if _, ok := nodeMap["payment_processing"]; ok {
		if _, ok := nodeMap["fraud_detection"]; ok {
			edges = append(edges, models.DependencyEdge{
				ID:              uuid.New(),
				ParentID:        nodeMap["payment_processing"],
				ChildID:         nodeMap["fraud_detection"],
				DefaultStrategy: string(models.StrategyProportionalOn),
				DefaultParameters: map[string]interface{}{
					"metric": "fraud_check_requests",
				},
				ActiveFrom: activeFrom,
			})
		}
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
	// All products return zero
	productNames := []string{
		"card_issuing", "payment_processing", "fraud_detection", "merchant_onboarding",
		"payment_gateway", "recurring_billing", "dispute_management", "digital_banking",
		"loan_origination", "wealth_management", "business_intelligence", "data_warehouse",
		"ml_platform", "customer_portal", "mobile_app", "notification_service",
		"kyc_verification", "aml_monitoring", "audit_logging", "api_marketplace",
		"webhook_service", "partner_integrations", "admin_console", "support_ticketing",
		"reporting_engine",
	}
	for _, product := range productNames {
		if nodeName == product {
			return decimal.Zero // Products have no direct infrastructure costs
		}
	}

	// High-cost Shared Infrastructure
	switch nodeName {
	case "payments_database_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(85.00) // $85/hour for db.r5.12xlarge cluster (8 instances)
		case "storage_gb_month":
			return decimal.NewFromFloat(1.25) // $1.25/GB/month for encrypted RDS storage
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for RDS egress
		case "iops":
			return decimal.NewFromFloat(0.75) // $0.75/IOPS for provisioned IOPS
		case "backups_gb_month":
			return decimal.NewFromFloat(0.95) // $0.95/GB/month for automated backups
		case "cpu_hours":
			return decimal.NewFromFloat(3.50) // $3.50/CPU hour for high-performance
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.80) // $0.80/GB/hour memory
		case "database_connections":
			return decimal.NewFromFloat(0.02) // $0.02/connection for payment workloads
		}
	case "analytics_database_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(55.00) // $55/hour for db.r5.8xlarge cluster (5 instances)
		case "storage_gb_month":
			return decimal.NewFromFloat(1.00) // $1.00/GB/month
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		case "iops":
			return decimal.NewFromFloat(0.65) // $0.65/IOPS
		}
	case "message_queue_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(45.00) // $45/hour for Kafka cluster (9 instances)
		case "storage_gb_month":
			return decimal.NewFromFloat(0.80) // $0.80/GB/month
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		}
	case "object_storage":
		switch dimension {
		case "storage_gb_month":
			return decimal.NewFromFloat(0.23) // $0.23/GB/month for S3 (500TB)
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		case "requests_count":
			return decimal.NewFromFloat(0.004) // $0.004/1000 requests
		}

	// High-cost Dedicated Compute Resources
	case "card_issuing_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(30.72) // $30.72/hour for 20x c5.4xlarge instances
		case "storage_gb_month":
			return decimal.NewFromFloat(2.00) // $2.00/GB/month for EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(1.92) // $1.92/CPU hour
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.48) // $0.48/GB/hour memory
		}
	case "payment_processing_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(153.60) // $153.60/hour for 50x c5.9xlarge instances
		case "storage_gb_month":
			return decimal.NewFromFloat(2.50) // $2.50/GB/month for high-IOPS EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(3.84) // $3.84/CPU hour for high-performance
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.96) // $0.96/GB/hour memory
		}
	case "fraud_ml_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(183.60) // $183.60/hour for 15x p3.8xlarge GPU instances
		case "storage_gb_month":
			return decimal.NewFromFloat(2.25) // $2.25/GB/month for NVMe SSD
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for EC2 egress
		case "cpu_hours":
			return decimal.NewFromFloat(11.48) // $11.48/CPU hour for GPU instances
		case "memory_gb_hours":
			return decimal.NewFromFloat(2.30) // $2.30/GB/hour memory for GPU instances
		}
	case "data_warehouse_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(195.00) // $195/hour for 30x ra3.4xlarge Redshift nodes
		case "storage_gb_month":
			return decimal.NewFromFloat(3.00) // $3.00/GB/month for Redshift storage
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		}
	case "ml_platform_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(244.80) // $244.80/hour for 10x ml.p3.16xlarge SageMaker instances
		case "storage_gb_month":
			return decimal.NewFromFloat(2.50) // $2.50/GB/month
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		}
	case "digital_banking_compute":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(47.50) // $47.50/hour for 25x m5.2xlarge instances
		case "storage_gb_month":
			return decimal.NewFromFloat(2.00) // $2.00/GB/month
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB
		}

	// Platform Services
	case "api_gateway_platform":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(3.60) // $3.60/hour for high-volume API Gateway
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for API Gateway egress
		case "load_balancer_hours":
			return decimal.NewFromFloat(2.25) // $2.25/hour for ALB
		case "requests_count":
			return decimal.NewFromFloat(0.0035) // $3.50 per million API requests
		case "data_transfer":
			return decimal.NewFromFloat(0.20) // $0.20/GB internal transfer
		}
	case "kubernetes_platform":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(25.00) // $25/hour for large EKS cluster
		case "storage_gb_month":
			return decimal.NewFromFloat(2.00) // $2.00/GB/month for EBS
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for EKS egress
		case "cpu_hours":
			return decimal.NewFromFloat(1.25) // $1.25/CPU hour for worker nodes
		case "memory_gb_hours":
			return decimal.NewFromFloat(0.31) // $0.31/GB/hour memory
		case "load_balancer_hours":
			return decimal.NewFromFloat(2.25) // $2.25/hour for ALB
		}
	case "cdn_platform":
		switch dimension {
		case "egress_gb":
			return decimal.NewFromFloat(0.85) // $0.85/GB for CloudFront
		case "requests_count":
			return decimal.NewFromFloat(0.0075) // $7.50 per million requests
		case "data_transfer":
			return decimal.NewFromFloat(0.20) // $0.20/GB
		}
	case "monitoring_platform":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(15.00) // $15/hour for Datadog infrastructure
		case "metrics_count":
			return decimal.NewFromFloat(0.05) // $0.05/custom metric
		case "logs_ingestion_gb":
			return decimal.NewFromFloat(0.10) // $0.10/GB for log ingestion
		}

	// Additional Shared Infrastructure
	case "redis_cache_cluster":
		switch dimension {
		case "instance_hours":
			return decimal.NewFromFloat(20.40) // $20.40/hour for 12x cache.r5.4xlarge
		case "egress_gb":
			return decimal.NewFromFloat(0.90) // $0.90/GB for ElastiCache egress
		case "memory_gb_hours":
			return decimal.NewFromFloat(1.02) // $1.02/GB/hour for cache memory
		}
	case "compliance_logging":
		switch dimension {
		case "logs_ingestion_gb":
			return decimal.NewFromFloat(5.00) // $5.00/GB for log ingestion (5TB/day)
		case "storage_gb_month":
			return decimal.NewFromFloat(0.30) // $0.30/GB/month for log storage
		case "cloudwatch_metrics":
			return decimal.NewFromFloat(3.00) // $3.00/metric/month
		case "monitoring_checks":
			return decimal.NewFromFloat(0.01) // $0.01/check
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
	// All 25 products generate usage metrics for proportional allocation
	allProducts := []string{
		"card_issuing", "payment_processing", "fraud_detection", "merchant_onboarding",
		"payment_gateway", "recurring_billing", "dispute_management", "digital_banking",
		"loan_origination", "wealth_management", "business_intelligence", "data_warehouse",
		"ml_platform", "customer_portal", "mobile_app", "notification_service",
		"kyc_verification", "aml_monitoring", "audit_logging", "api_marketplace",
		"webhook_service", "partner_integrations", "admin_console", "support_ticketing",
		"reporting_engine",
	}

	// Check if this is a product node
	isProduct := false
	for _, p := range allProducts {
		if nodeName == p {
			isProduct = true
			break
		}
	}

	if isProduct {
		// All products generate usage metrics for proportional allocation
		switch metric {
		case "api_requests":
			// Vary by product type - high traffic products get more
			switch nodeName {
			case "payment_processing", "payment_gateway", "fraud_detection":
				return decimal.NewFromInt(5000) // 5000 API requests/hour
			case "card_issuing", "digital_banking", "mobile_app":
				return decimal.NewFromInt(3000) // 3000 API requests/hour
			default:
				return decimal.NewFromInt(1000) // 1000 API requests/hour
			}
		case "usage_metric":
			// Generic usage metric for shared infrastructure
			return decimal.NewFromInt(100)
		}
	}

	// Infrastructure nodes don't generate usage - they only consume costs
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
