-- Drop triggers
DROP TRIGGER IF EXISTS update_contribution_results_updated_at ON contribution_results_by_dimension;
DROP TRIGGER IF EXISTS update_allocation_results_updated_at ON allocation_results_by_dimension;
DROP TRIGGER IF EXISTS update_computation_runs_updated_at ON computation_runs;
DROP TRIGGER IF EXISTS update_node_usage_updated_at ON node_usage_by_dimension;
DROP TRIGGER IF EXISTS update_node_costs_updated_at ON node_costs_by_dimension;
DROP TRIGGER IF EXISTS update_edge_strategies_updated_at ON edge_strategies;
DROP TRIGGER IF EXISTS update_dependency_edges_updated_at ON dependency_edges;
DROP TRIGGER IF EXISTS update_cost_nodes_updated_at ON cost_nodes;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS contribution_results_by_dimension;
DROP TABLE IF EXISTS allocation_results_by_dimension;
DROP TABLE IF EXISTS computation_runs;
DROP TABLE IF EXISTS node_usage_by_dimension;
DROP TABLE IF EXISTS node_costs_by_dimension;
DROP TABLE IF EXISTS edge_strategies;
DROP TABLE IF EXISTS dependency_edges;
DROP TABLE IF EXISTS cost_nodes;

-- Drop extension (only if no other tables use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
