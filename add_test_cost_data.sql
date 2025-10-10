-- Add test cost data to see the dashboard working with actual numbers
-- This script adds random costs to resource nodes for January 2024

-- First, add some direct costs to resource nodes
INSERT INTO node_costs_by_dimension (node_id, cost_date, dimension, amount, currency)
SELECT 
    id,
    date_series.cost_date,
    'compute_hours',
    (random() * 100 + 50)::numeric(38,9), -- Random cost between $50-$150 per day
    'USD'
FROM cost_nodes
CROSS JOIN (
    SELECT generate_series(
        '2024-01-01'::date,
        '2024-01-31'::date,
        '1 day'::interval
    )::date as cost_date
) date_series
WHERE type = 'resource'
ON CONFLICT (node_id, cost_date, dimension) DO NOTHING;

-- Add some costs to platform nodes
INSERT INTO node_costs_by_dimension (node_id, cost_date, dimension, amount, currency)
SELECT 
    id,
    date_series.cost_date,
    'infrastructure',
    (random() * 200 + 100)::numeric(38,9), -- Random cost between $100-$300 per day
    'USD'
FROM cost_nodes
CROSS JOIN (
    SELECT generate_series(
        '2024-01-01'::date,
        '2024-01-31'::date,
        '1 day'::interval
    )::date as cost_date
) date_series
WHERE type = 'platform'
ON CONFLICT (node_id, cost_date, dimension) DO NOTHING;

-- Add some costs to shared services
INSERT INTO node_costs_by_dimension (node_id, cost_date, dimension, amount, currency)
SELECT 
    id,
    date_series.cost_date,
    'shared_infrastructure',
    (random() * 150 + 75)::numeric(38,9), -- Random cost between $75-$225 per day
    'USD'
FROM cost_nodes
CROSS JOIN (
    SELECT generate_series(
        '2024-01-01'::date,
        '2024-01-31'::date,
        '1 day'::interval
    )::date as cost_date
) date_series
WHERE type = 'shared'
ON CONFLICT (node_id, cost_date, dimension) DO NOTHING;

-- Verify the data was added
SELECT 
    'node_costs_by_dimension' as table_name,
    COUNT(*) as row_count,
    SUM(amount) as total_amount,
    MIN(cost_date) as min_date,
    MAX(cost_date) as max_date
FROM node_costs_by_dimension;

-- Show cost breakdown by node type
SELECT 
    n.type,
    COUNT(DISTINCT n.id) as node_count,
    COUNT(*) as cost_records,
    SUM(c.amount) as total_cost
FROM node_costs_by_dimension c
JOIN cost_nodes n ON c.node_id = n.id
WHERE c.cost_date BETWEEN '2024-01-01' AND '2024-01-31'
GROUP BY n.type
ORDER BY total_cost DESC;

