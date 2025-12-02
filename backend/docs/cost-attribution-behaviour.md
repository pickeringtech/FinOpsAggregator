# Cost Attribution Behaviour (Canonical Spec)

This document defines the **canonical behaviour** for cost attribution in the system. It should be treated as the source of truth for how costs flow through the dependency graph and how product-level totals and coverage are calculated.

It is intended to be read together with `backend/docs/dependency-model-analysis.md`, which focuses on the dependency edge direction and a few concrete test cases.

## 1. Goals and scope

- Model how **infrastructure and shared costs** flow into **products**.
- Support **product roll-ups and partial roll-ups** over time without breaking financial invariants.
- Provide precise definitions for:
  - Raw Infrastructure Cost
  - Allocated Product Cost
  - Product Holistic Cost
  - Unallocated Infrastructure Cost
  - Coverage
  - Final cost centres
- Describe how **allocation strategies** are used for infra → product and product → product allocations.

## 2. Node types and graphs

### 2.1 Node types

The system models nodes in a cost attribution graph:

- **Resource** – raw infrastructure components (e.g. compute instances, disks).
- **Shared Service** – shared internal services (e.g. shared database clusters).
- **Platform Service** – shared platform or tooling (e.g. Kubernetes, CI/CD, observability).
- **Product** – business-facing products you ultimately want to report on.

Some internal, shared Products may act similarly to Shared/Platform nodes from a cost-flow perspective.

### 2.2 Allocation graph (cost-flow graph)

The **allocation graph** is the directed graph used by the allocation engine. Its edges represent **cost flowing from a source node to a receiver node**:

- **Direction:**
  - `Resource/Shared/Platform → Product`
  - `Shared/Platform → Shared/Product`
  - `Product → Product` (for product roll-ups)
- In all cases, **ParentID = cost source**, **ChildID = cost receiver**.

Costs always flow **downstream along edges** using allocation strategies. For any edge `parent → child`, a share of the parent’s cost is allocated into the child.

### 2.3 Product hierarchy (business view)

Separately, the **product hierarchy** is a tree/DAG used for navigation and display in the UI:

- Edges represent **"is part of" / "rolls up under"** relationships between Products.
- A "root" in this hierarchy is a Product that has no parent in the hierarchy table.
- The hierarchy is used for **display and aggregation**, not for driving raw allocation mathematics.

The allocation graph and hierarchy often align, but **cost invariants must be defined in terms of the allocation graph**, not the hierarchy alone.

## 3. Cost definitions and invariants

For a given reporting period (e.g. a day, month, or arbitrary date range):

### 3.1 Base quantities

- **Direct Cost (per node)** – the cost that originates on a node before allocation.
  - For Resources/Shared/Platform: cloud bills, infra costs, etc.
  - For Products: typically zero in this model, unless explicitly modelled.
- **Indirect Cost (per node)** – cost received from **parent nodes** via allocation along edges.
- **Total (Holistic) Cost (per node)** –
  - `Holistic(node) = Direct(node) + Indirect(node)`.

### 3.2 Global aggregates

- **Raw Infrastructure Cost** –
  - `RawInfraCost = Σ Holistic(node)` over all cost-bearing infra-like nodes
    (Resources, Shared Services, Platform Services, and any internal Product nodes that we choose to treat as infra sources).
  - In practice this should equal the sum of **Direct** costs on those nodes before allocation.

- **Allocated Product Cost** –
  - Sum of **Holistic Cost** over all **final cost centre Products** (definition in section 4).

- **Unallocated Infrastructure Cost** –
  - `UnallocatedInfraCost = RawInfraCost − AllocatedProductCost` (clamped at ≥ 0, within rounding tolerance).

- **Coverage (%)** –
  - If `RawInfraCost > 0`:
    - `Coverage = AllocatedProductCost / RawInfraCost` (expressed as a percentage, 0–100%).
  - If `RawInfraCost == 0`, coverage is defined as 0.

### 3.3 Core invariants

For any allocation run over a given period:

1. **Conservation of cost** (no amplification):
   - `AllocatedProductCost + UnallocatedInfraCost = RawInfraCost` (up to rounding).
2. **Final cost centre sum**:
   - `AllocatedProductCost = Σ Holistic(p)` over all final cost centre Products `p`.
3. **Hierarchy consistency**:
   - A hierarchy view that rolls up costs from leaves must be consistent with `AllocatedProductCost` and must not double count intermediate Product nodes.

## 4. Final cost centres

### 4.1 Definition

For any given day or reporting window, a **final cost centre** is:

> A Product node that does **not** allocate any of its cost into another Product during that period.

In terms of the allocation graph:

- A Product `P` is a final cost centre if there is **no active allocation edge**
  `P → Q` where `Q` is also a Product in the same period.
- Equivalently, `P` does **not** appear as `ParentID` on any Product→Product `DependencyEdge` that is active for that date range.

Infra-like nodes (Resources, Shared, Platform, and internal services) are **not** final cost centres; they are sources of cost that should be fully or partially allocated into Products.

### 4.2 Rationale

- Cost should always "stop" at a set of **ultimate Products** that represent what you report P&L on.
- Intermediate Products (or internal shared Products) may receive cost and then allocate it onwards, but they are not final destinations.
- By defining final cost centres via the **absence of outgoing Product→Product edges**, the model automatically adapts as product structures evolve.

### 4.3 Behaviour over time

This definition supports an evolving product catalogue:

- Initially, a Product has no Product→Product edges ⇒ it is a final cost centre.
- Later, if we add edges `P → Q` (Product roll-ups):
  - From that point forward, `P` stops being a final cost centre for the portion of cost that is allocated onwards.
  - Products like `Q` that do not allocate further become the new final cost centres.
- Historical periods remain unchanged because edges are time-bounded with `ActiveFrom` / `ActiveTo`.

## 5. Product roll-ups and partial roll-ups

### 5.1 Product→Product edges

Product roll-ups are modelled using the same allocation mechanics as infra allocation:

- Edges: `Product_source → Product_parent`.
- Interpretation: "Product_source allocates some share of its cost into Product_parent".

A Product may have **multiple parents**; the system must allow `Product_source → P1`, `Product_source → P2`, etc.

### 5.2 Partial vs full roll-ups

For a Product `P` with total holistic cost `C_P` in a period and outgoing edges to parents `{Q1...Qk}` with shares `{s1...sk}`:

- **Allocated to parents:** each `Qi` receives `C_P * si` as indirect cost.
- **Residual on P:** `C_residual = C_P * (1 − Σ si)`.

Rules:

- The sum of shares `Σ si` **does not have to equal 1**:
  - If `Σ si == 1`: P becomes a pure intermediate (all its cost is rolled up).
  - If `Σ si < 1`: P retains `C_residual` and remains (partly) a final cost centre.
- `Σ si > 1` is invalid for financial consistency and must be prevented or validated.

This supports scenarios where only a **percentage of a Product** is rolled up into parents.

### 5.3 Interaction with final cost centres

- If `P` has **no outgoing Product → Product** edges in the period, it is a final cost centre.
- If `P` has outgoing Product → Product edges but `Σ si < 1`, then:
  - The **residual** part of `P` is effectively still a final cost centre.
  - Globally, we can still treat `P` as a final cost centre and include its holistic cost, because the allocations to parents are part of its outgoing contributions.
- Final cost centres therefore form the set of Products that **do not allocate further** in the graph, even if their own cost includes allocations from other Products or infra.

## 6. Allocation strategies

Allocation strategies determine **how much** of a parent node’s cost is allocated to each child along an edge.

### 6.1 Existing strategy types (implemented)

Defined in `internal/models/types.go` and implemented in `internal/allocate/strategies.go`:

- **`equal`** (`StrategyEqual`)
  - Parent cost is split equally among all children.
- **`proportional_on`** (`StrategyProportionalOn`)
  - Parent cost is split proportionally based on a usage metric (e.g. CPU, requests).
- **`fixed_percent`** (`StrategyFixedPercent`)
  - A fixed percentage of the parent’s cost is allocated to the child, based on a `percent` parameter.
- **`capped_proportional`** (`StrategyCappedProp`)
  - Proportional allocation with a per-child cap; share is `min(proportional_share, cap)`.
- **`residual_to_max`** (`StrategyResidualToMax`)
  - In multi-parent scenarios, assigns the residual share to the parent with maximum usage after other parents have received proportional shares.

These strategies apply uniformly to:

- **Infra/Shared/Platform → Product** allocations.
- **Product → Product** allocations (for roll-ups), where usage data may come from application metrics.

### 6.2 Metrics and Dynatrace integration

Usage metrics can be derived from application monitoring systems such as Dynatrace. Typical examples:

- HTTP request counts.
- Request durations.
- Metrics labelled by `customer_id`, plan, region, etc.

These metrics can be ingested into `NodeUsageByDimension` (or an extended model), and strategies like `proportional_on` can:

- Use a chosen metric (e.g. `http_requests`, `http_duration_ms`).
- Optionally apply filters or segment definitions (e.g. subsets of `customer_id`s) to define how cost is split between parent Products.

### 6.3 Planned strategy extensions (conceptual)

The following strategy patterns are useful and can be implemented on top of the existing framework:

1. **Time-windowed proportional (weighted average)**
   - Use a look-back window of N days for a usage metric.
   - Allocate based on a weighted average over the window, smoothing out noisy daily data.

2. **Hybrid fixed + proportional**
   - Split parent cost into a base fraction allocated equally to all children and a variable fraction allocated proportionally by usage.
   - Useful for modelling a "base platform fee" plus a variable component.

3. **Minimum floor + proportional remainder**
   - Guarantee each child a minimum share or amount, and allocate the remaining cost proportionally by usage.
   - Useful when every child should carry some baseline overhead.

These strategies are especially relevant when using Dynatrace HTTP metrics with `customer_id` labels to drive Product → Product roll-ups and customer segment allocations.

## 7. Test cases and validation

The examples in `dependency-model-analysis.md` remain the primary **sanity test cases** for the dependency direction:

1. Basic Resource → Product allocation.
2. Shared Service → Multiple Products (equal allocation).
3. Platform → Products with proportional allocation.

Additional test cases should cover:

- Product → Product roll-ups with partial shares and multiple parents.
- Final cost centre identification based on the absence of outgoing Product → Product edges.
- Conservation invariants:
  - `AllocatedProductCost + UnallocatedInfraCost = RawInfraCost`.
  - `AllocatedProductCost = Σ Holistic(p)` over final cost centre Products.
- Consistency between allocation results and the Product hierarchy.

These tests form the **"red" phase** in a red–green–blue development cycle: they should be written first to encode the behaviour in this document and then used to drive implementation and refactoring.

## 8. Developer Runbooks

### 8.1 How to debug allocation issues

When allocation results don't match expectations, follow this debugging workflow:

#### Step 1: Use the debug/reconciliation endpoint

```bash
curl "http://localhost:8080/api/v1/debug/reconciliation?start_date=2024-01-01&end_date=2024-01-31&currency=USD"
```

This endpoint returns:
- **Raw Infrastructure Cost**: Total direct costs on infra-like nodes
- **Allocated Product Cost**: Sum of holistic costs on final cost centres
- **Unallocated Cost**: Gap between raw and allocated
- **Coverage Percent**: Allocation efficiency
- **Conservation Delta**: Should be ~0 if invariants hold
- **Final Cost Centres**: List of product nodes with their holistic costs
- **Infrastructure Nodes**: List of infra nodes with their direct costs
- **Invariant Violations**: Any detected violations

#### Step 2: Check sanity check warnings

The `/api/v1/products/hierarchy` endpoint includes `sanity_check_passed` and `sanity_check_warnings` in the summary. If `sanity_check_passed` is false, review the warnings.

#### Step 3: Enable trace logging

Set log level to `trace` to see individual strategy calculations:

```go
// Each strategy share calculation is logged at trace level:
// "Strategy share calculated" with strategy, parent_id, child_id, dimension, share
```

#### Step 4: Check for common issues

1. **Missing edges**: Ensure all infra nodes have edges to products
2. **Cycle detection**: The graph builder validates for cycles
3. **Invalid shares**: Sum of shares from a parent should be ≤ 1
4. **Missing usage data**: Proportional strategies require usage metrics

#### Step 5: Validate the graph structure

```go
// Use the graph validator
validator := graph.NewValidator()
errors := validator.Validate(g)
```

### 8.2 How to add a new allocation strategy

#### Step 1: Define the strategy constant

In `internal/models/types.go`:

```go
const (
    // ... existing strategies
    StrategyMyNewStrategy AllocationStrategy = "my_new_strategy"
)
```

#### Step 2: Update the validation function

In `internal/models/types.go`, add to `IsValidStrategy()`:

```go
func IsValidStrategy(s string) bool {
    switch AllocationStrategy(s) {
    case StrategyEqual, StrategyProportionalOn, /* ... */, StrategyMyNewStrategy:
        return true
    }
    return false
}
```

#### Step 3: Implement the calculation function

In `internal/allocate/strategies.go`:

```go
func (s *Strategy) calculateMyNewStrategyShare(
    ctx context.Context,
    store *store.Store,
    parentID, childID uuid.UUID,
    dimension string,
    date time.Time,
) (decimal.Decimal, error) {
    // Get parameters from s.Parameters
    param1, ok := s.Parameters["param1"].(string)
    if !ok {
        return decimal.Zero, fmt.Errorf("my_new_strategy requires 'param1' parameter")
    }

    // Get edges from parent to find all children
    edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
    if err != nil {
        return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
    }

    // Calculate and return the share for this child
    // Share should be in range [0, 1]
    return share, nil
}
```

#### Step 4: Wire into the strategy switch

In `internal/allocate/strategies.go`, add to `CalculateShare()`:

```go
case models.StrategyMyNewStrategy:
    share, err = s.calculateMyNewStrategyShare(ctx, store, parentID, childID, dimension, date)
```

#### Step 5: Add tests

In `internal/allocate/canonical_test.go` or a new test file:

```go
func TestMyNewStrategy(t *testing.T) {
    // Test the strategy with various inputs
    // Verify share calculations are correct
    // Verify edge cases (no children, zero usage, etc.)
}
```

#### Step 6: Document the strategy

Add documentation to `docs/allocation-strategies.md` with:
- Strategy name and description
- Required parameters
- Example use cases
- Mathematical formula

### 8.3 How to model product roll-ups

Product roll-ups allow costs to flow from one product to another, enabling hierarchical cost aggregation.

#### Scenario 1: Full roll-up (100% of Product A → Product B)

Use `fixed_percent` strategy with `percent: 100`:

```sql
INSERT INTO dependency_edges (parent_id, child_id, default_strategy, default_parameters)
VALUES (
    'product-a-uuid',
    'product-b-uuid',
    'fixed_percent',
    '{"percent": 100}'
);
```

After allocation:
- Product A's holistic cost flows entirely to Product B
- Product A is NOT a final cost centre (it has an outgoing product edge)
- Product B IS a final cost centre (if it has no outgoing product edges)

#### Scenario 2: Partial roll-up (30% to B, 50% to C)

Create multiple edges with `fixed_percent`:

```sql
-- 30% to Product B
INSERT INTO dependency_edges (parent_id, child_id, default_strategy, default_parameters)
VALUES ('product-a-uuid', 'product-b-uuid', 'fixed_percent', '{"percent": 30}');

-- 50% to Product C
INSERT INTO dependency_edges (parent_id, child_id, default_strategy, default_parameters)
VALUES ('product-a-uuid', 'product-c-uuid', 'fixed_percent', '{"percent": 50}');
```

After allocation:
- 30% of Product A's holistic cost → Product B
- 50% of Product A's holistic cost → Product C
- 20% remains on Product A (unallocated from A's perspective)
- Product A is NOT a final cost centre

#### Scenario 3: Usage-based roll-up (proportional to customer usage)

Use `proportional_on` or `segment_filtered_proportional`:

```sql
INSERT INTO dependency_edges (parent_id, child_id, default_strategy, default_parameters)
VALUES (
    'shared-product-uuid',
    'customer-product-uuid',
    'segment_filtered_proportional',
    '{"metric": "http_requests", "segment_filter": {"label": "customer_id", "values": ["cust-123"]}}'
);
```

#### Key concepts for roll-ups

1. **Final cost centres**: Only products with NO outgoing product→product edges are final cost centres. The `AllocatedProductCost` is the sum of holistic costs over final cost centres only.

2. **Topological order**: The allocation engine processes nodes in topological order, so parent products are fully allocated before their costs flow to child products.

3. **Multi-parent support**: A product can receive costs from multiple parents. The indirect cost is the sum of all contributions.

4. **Diamond patterns**: If Product A → B and A → C, and both B and C → D, the costs flow correctly without double-counting because we only sum final cost centres.

#### Validation

After setting up roll-ups, verify:

1. Conservation: `AllocatedProductCost + UnallocatedCost = RawInfraCost`
2. No amplification: `AllocatedProductCost ≤ RawInfraCost`
3. Final cost centre identification: Use the debug endpoint to list final cost centres

