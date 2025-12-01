-- Rollback migration to restore inverted dependency relationships
--
-- This rollback:
-- 1. Identifies all corrected edges (resources/shared/platform → products)
-- 2. Creates edges with the old inverted relationships (products → resources/shared/platform)
-- 3. Removes the corrected edges
--
-- WARNING: This rollback restores the INCORRECT dependency model
-- Only use this if you need to rollback for debugging purposes

BEGIN;

-- Step 1: Create a temporary table to store the reverted edges
CREATE TEMP TABLE reverted_edges AS
SELECT
    uuid_generate_v4() as new_id,
    de.id as old_id,
    de.child_id as new_parent_id,  -- What was child becomes parent (reverting)
    de.parent_id as new_child_id,  -- What was parent becomes child (reverting)
    de.default_strategy,
    de.default_parameters,
    de.active_from,
    de.active_to,
    parent_node.type as parent_type,
    child_node.type as child_type
FROM dependency_edges de
JOIN cost_nodes parent_node ON de.parent_id = parent_node.id
JOIN cost_nodes child_node ON de.child_id = child_node.id
WHERE
    -- Find corrected edges: infrastructure as parents of products
    parent_node.type IN ('resource', 'shared', 'platform')
    AND child_node.type = 'product';

-- Step 2: Log what we're about to revert
DO $$
DECLARE
    edge_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO edge_count FROM reverted_edges;
    RAISE NOTICE 'Found % corrected dependency edges to revert', edge_count;
    RAISE WARNING 'This will restore the INCORRECT dependency model!';
END $$;

-- Step 3: Insert the reverted edges
INSERT INTO dependency_edges (
    id,
    parent_id,
    child_id,
    default_strategy,
    default_parameters,
    active_from,
    active_to,
    created_at,
    updated_at
)
SELECT
    new_id,
    new_parent_id,
    new_child_id,
    default_strategy,
    default_parameters,
    active_from,
    active_to,
    now(),
    now()
FROM reverted_edges;

-- Step 4: Delete the corrected edges
DELETE FROM dependency_edges
WHERE id IN (SELECT old_id FROM reverted_edges);

-- Step 5: Log completion
DO $$
DECLARE
    total_edges INTEGER;
    product_parent_edges INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_edges FROM dependency_edges;

    SELECT COUNT(*) INTO product_parent_edges
    FROM dependency_edges de
    JOIN cost_nodes parent_node ON de.parent_id = parent_node.id
    WHERE parent_node.type = 'product';

    RAISE NOTICE 'Rollback completed';
    RAISE NOTICE 'Total dependency edges: %', total_edges;
    RAISE NOTICE 'Edges with products as parents (INCORRECT): %', product_parent_edges;
    RAISE WARNING 'The dependency model is now INCORRECT - products are parents of infrastructure!';
END $$;

COMMIT;