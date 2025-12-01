# Dependency Model Analysis

## Current (Incorrect) Model

The current dependency model has **inverted relationships** that cause the allocation coverage issue:

### Current Edge Direction
```
Product → Resource/Shared/Platform
```

### Example from seed.go
```go
// WRONG: Products as parents of resources
edges = append(edges, models.DependencyEdge{
    ParentID:        nodeMap[product],     // Product is parent
    ChildID:         nodeMap[resource],    // Resource is child
    DefaultStrategy: string(models.StrategyEqual),
})
```

### Current Database State
```sql
-- Current (wrong) edges
product_id → shared_service_id
product_id → platform_service_id  
product_id → resource_id
```

### Problems with Current Model
1. **Products have $0 direct costs** (correct)
2. **Resources/Shared/Platform have $1.7M direct costs** (correct)
3. **Allocation flows from Products → Resources** (WRONG\!)
4. **Products try to allocate $0 to resources** → Almost no allocation
5. **Result: 0.33% allocation coverage**

## Correct Model

### Correct Edge Direction
```
Resource/Shared/Platform → Product
```

### Cost Flow Direction
```
Direct Costs: Resource/Shared/Platform ($1.7M)
       ↓ (allocation)
Allocated Costs: Products (should receive $1.7M)
```

### Correct Database State
```sql
-- Correct edges
shared_service_id → product_id
platform_service_id → product_id
resource_id → product_id
```

### Expected Results with Correct Model
1. **Resources/Shared/Platform have $1.7M direct costs** ✓
2. **Allocation flows from Resources → Products** ✓
3. **Products receive allocated costs from dependencies** ✓
4. **Result: ~100% allocation coverage** ✓

## Test Cases to Validate

### Test Case 1: Basic Resource → Product Allocation
```
Resource "compute" ($100/day) → Product "web_app"
Expected: Product receives $100/day allocated cost
```

### Test Case 2: Shared Service → Multiple Products
```
Shared "database" ($300/day) → Products ["web_app", "api", "mobile"]
Expected: Each product receives $100/day (equal allocation)
```

### Test Case 3: Platform → Products with Proportional Allocation
```
Platform "kubernetes" ($500/day) → Products based on CPU usage
Expected: Products receive costs proportional to their CPU usage
```
