-- Migration to fix inverted dependency relationships
--
-- Problem: Current edges have products as parents of resources/shared/platform services
-- Solution: Reverse the edges so resources/shared/platform are parents of products
--
-- This migration:
-- 1. Identifies all inverted edges (products → resources/shared/platform)
-- 2. Creates new edges with correct relationships (resources/shared/platform → products)
-- 3. Removes the old inverted edges
--
-- Expected impact: ~514 edges will be reversed

BEGIN;

-- Step 1: Create a temporary table to store the corrected edges
CREATE TEMP TABLE corrected_edges AS
SELECT
    uuid_generate_v4() as new_id,
    de.id as old_id,
    de.child_id as new_parent_id,  -- What was child becomes parent
    de.parent_id as new_child_id,  -- What was parent becomes child
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
    -- Find inverted edges: products as parents of infrastructure
    parent_node.type = 'product'
    AND child_node.type IN ('resource', 'shared', 'platform');

-- Step 2: Log what we're about to fix
DO $$
DECLARE
    edge_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO edge_count FROM corrected_edges;
    RAISE NOTICE 'Found % inverted dependency edges to fix', edge_count;

    -- Log breakdown by child type
    FOR edge_count IN
        SELECT COUNT(*) FROM corrected_edges WHERE child_type = 'resource'
    LOOP
        RAISE NOTICE '  - % edges: product → resource (will become resource → product)', edge_count;
    END LOOP;

    FOR edge_count IN
        SELECT COUNT(*) FROM corrected_edges WHERE child_type = 'shared'
    LOOP
        RAISE NOTICE '  - % edges: product → shared (will become shared → product)', edge_count;
    END LOOP;

    FOR edge_count IN
        SELECT COUNT(*) FROM corrected_edges WHERE child_type = 'platform'
    LOOP
        RAISE NOTICE '  - % edges: product → platform (will become platform → product)', edge_count;
    END LOOP;
END $$;

-- Step 3: Insert the corrected edges
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
FROM corrected_edges;

-- Step 4: Delete the old inverted edges
DELETE FROM dependency_edges
WHERE id IN (SELECT old_id FROM corrected_edges);

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

    RAISE NOTICE 'Migration completed successfully';
    RAISE NOTICE 'Total dependency edges: %', total_edges;
    RAISE NOTICE 'Remaining edges with products as parents: %', product_parent_edges;

    IF product_parent_edges > 0 THEN
        RAISE WARNING 'Still have % edges with products as parents - these may be legitimate product-to-product dependencies', product_parent_edges;
    END IF;
END $$;

COMMIT;