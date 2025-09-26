-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Cost nodes table
CREATE TABLE cost_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    cost_labels JSONB NOT NULL DEFAULT '{}',
    is_platform BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ,
    
    CONSTRAINT cost_nodes_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT cost_nodes_type_not_empty CHECK (length(trim(type)) > 0)
);

-- Create indexes for cost_nodes
CREATE INDEX idx_cost_nodes_name ON cost_nodes(name);
CREATE INDEX idx_cost_nodes_type ON cost_nodes(type);
CREATE INDEX idx_cost_nodes_is_platform ON cost_nodes(is_platform);
CREATE INDEX idx_cost_nodes_archived_at ON cost_nodes(archived_at) WHERE archived_at IS NOT NULL;

-- Dependency edges table
CREATE TABLE dependency_edges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    child_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    default_strategy TEXT NOT NULL,
    default_parameters JSONB NOT NULL DEFAULT '{}',
    active_from DATE NOT NULL,
    active_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT dependency_edges_parent_child_different CHECK (parent_id != child_id),
    CONSTRAINT dependency_edges_active_dates CHECK (active_to IS NULL OR active_to > active_from),
    CONSTRAINT dependency_edges_strategy_not_empty CHECK (length(trim(default_strategy)) > 0),
    UNIQUE(parent_id, child_id, active_from)
);

-- Create indexes for dependency_edges
CREATE INDEX idx_dependency_edges_parent_id ON dependency_edges(parent_id);
CREATE INDEX idx_dependency_edges_child_id ON dependency_edges(child_id);
CREATE INDEX idx_dependency_edges_active_from ON dependency_edges(active_from);
CREATE INDEX idx_dependency_edges_active_to ON dependency_edges(active_to) WHERE active_to IS NOT NULL;

-- Edge strategies table (dimension-specific overrides)
CREATE TABLE edge_strategies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    edge_id UUID NOT NULL REFERENCES dependency_edges(id) ON DELETE CASCADE,
    dimension TEXT,
    strategy TEXT NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT edge_strategies_strategy_not_empty CHECK (length(trim(strategy)) > 0),
    UNIQUE(edge_id, dimension)
);

-- Create indexes for edge_strategies
CREATE INDEX idx_edge_strategies_edge_id ON edge_strategies(edge_id);
CREATE INDEX idx_edge_strategies_dimension ON edge_strategies(dimension);

-- Node costs by dimension table
CREATE TABLE node_costs_by_dimension (
    node_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    cost_date DATE NOT NULL,
    dimension TEXT NOT NULL,
    amount NUMERIC(38, 9) NOT NULL,
    currency TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT node_costs_dimension_not_empty CHECK (length(trim(dimension)) > 0),
    CONSTRAINT node_costs_currency_not_empty CHECK (length(trim(currency)) > 0),
    CONSTRAINT node_costs_amount_non_negative CHECK (amount >= 0),
    PRIMARY KEY (node_id, cost_date, dimension)
);

-- Create indexes for node_costs_by_dimension
CREATE INDEX idx_node_costs_cost_date ON node_costs_by_dimension(cost_date);
CREATE INDEX idx_node_costs_dimension ON node_costs_by_dimension(dimension);
CREATE INDEX idx_node_costs_currency ON node_costs_by_dimension(currency);

-- Node usage by dimension table
CREATE TABLE node_usage_by_dimension (
    node_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    metric TEXT NOT NULL,
    value NUMERIC(38, 9) NOT NULL,
    unit TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT node_usage_metric_not_empty CHECK (length(trim(metric)) > 0),
    CONSTRAINT node_usage_unit_not_empty CHECK (length(trim(unit)) > 0),
    CONSTRAINT node_usage_value_non_negative CHECK (value >= 0),
    PRIMARY KEY (node_id, usage_date, metric)
);

-- Create indexes for node_usage_by_dimension
CREATE INDEX idx_node_usage_usage_date ON node_usage_by_dimension(usage_date);
CREATE INDEX idx_node_usage_metric ON node_usage_by_dimension(metric);
CREATE INDEX idx_node_usage_unit ON node_usage_by_dimension(unit);

-- Computation runs table
CREATE TABLE computation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    window_start DATE NOT NULL,
    window_end DATE NOT NULL,
    graph_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    notes TEXT,
    
    CONSTRAINT computation_runs_window_valid CHECK (window_end >= window_start),
    CONSTRAINT computation_runs_status_valid CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    CONSTRAINT computation_runs_graph_hash_not_empty CHECK (length(trim(graph_hash)) > 0)
);

-- Create indexes for computation_runs
CREATE INDEX idx_computation_runs_window_start ON computation_runs(window_start);
CREATE INDEX idx_computation_runs_window_end ON computation_runs(window_end);
CREATE INDEX idx_computation_runs_status ON computation_runs(status);
CREATE INDEX idx_computation_runs_graph_hash ON computation_runs(graph_hash);

-- Allocation results by dimension table
CREATE TABLE allocation_results_by_dimension (
    run_id UUID NOT NULL REFERENCES computation_runs(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    allocation_date DATE NOT NULL,
    dimension TEXT NOT NULL,
    direct_amount NUMERIC(38, 9) NOT NULL,
    indirect_amount NUMERIC(38, 9) NOT NULL,
    total_amount NUMERIC(38, 9) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT allocation_results_dimension_not_empty CHECK (length(trim(dimension)) > 0),
    CONSTRAINT allocation_results_amounts_non_negative CHECK (
        direct_amount >= 0 AND indirect_amount >= 0 AND total_amount >= 0
    ),
    CONSTRAINT allocation_results_total_equals_sum CHECK (
        total_amount = direct_amount + indirect_amount
    ),
    PRIMARY KEY (run_id, node_id, allocation_date, dimension)
);

-- Create indexes for allocation_results_by_dimension
CREATE INDEX idx_allocation_results_run_id ON allocation_results_by_dimension(run_id);
CREATE INDEX idx_allocation_results_node_id ON allocation_results_by_dimension(node_id);
CREATE INDEX idx_allocation_results_allocation_date ON allocation_results_by_dimension(allocation_date);
CREATE INDEX idx_allocation_results_dimension ON allocation_results_by_dimension(dimension);

-- Contribution results by dimension table
CREATE TABLE contribution_results_by_dimension (
    run_id UUID NOT NULL REFERENCES computation_runs(id) ON DELETE CASCADE,
    parent_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    child_id UUID NOT NULL REFERENCES cost_nodes(id) ON DELETE CASCADE,
    contribution_date DATE NOT NULL,
    dimension TEXT NOT NULL,
    contributed_amount NUMERIC(38, 9) NOT NULL,
    path JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT contribution_results_dimension_not_empty CHECK (length(trim(dimension)) > 0),
    CONSTRAINT contribution_results_amount_non_negative CHECK (contributed_amount >= 0),
    CONSTRAINT contribution_results_parent_child_different CHECK (parent_id != child_id),
    PRIMARY KEY (run_id, parent_id, child_id, contribution_date, dimension)
);

-- Create indexes for contribution_results_by_dimension
CREATE INDEX idx_contribution_results_run_id ON contribution_results_by_dimension(run_id);
CREATE INDEX idx_contribution_results_parent_id ON contribution_results_by_dimension(parent_id);
CREATE INDEX idx_contribution_results_child_id ON contribution_results_by_dimension(child_id);
CREATE INDEX idx_contribution_results_contribution_date ON contribution_results_by_dimension(contribution_date);
CREATE INDEX idx_contribution_results_dimension ON contribution_results_by_dimension(dimension);

-- Update triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables
CREATE TRIGGER update_cost_nodes_updated_at BEFORE UPDATE ON cost_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dependency_edges_updated_at BEFORE UPDATE ON dependency_edges FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_edge_strategies_updated_at BEFORE UPDATE ON edge_strategies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_node_costs_updated_at BEFORE UPDATE ON node_costs_by_dimension FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_node_usage_updated_at BEFORE UPDATE ON node_usage_by_dimension FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_computation_runs_updated_at BEFORE UPDATE ON computation_runs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_allocation_results_updated_at BEFORE UPDATE ON allocation_results_by_dimension FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contribution_results_updated_at BEFORE UPDATE ON contribution_results_by_dimension FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
