# Worked Examples: Cost Allocation Scenarios

This document provides detailed worked examples for common cost allocation scenarios.

## Example 1: Basic Infrastructure → Product Allocation

### Scenario
A compute resource allocates its costs to a single product.

### Graph
```
┌─────────────────┐         ┌─────────────────┐
│    Resource     │         │     Product     │
│   "compute"     │────────▶│    "web_app"    │
│   $100/day      │  equal  │    $0 direct    │
└─────────────────┘         └─────────────────┘
```

### Input Data
| Node | Type | Direct Cost |
|------|------|-------------|
| compute | resource | $100/day |
| web_app | product | $0/day |

### Edge Configuration
| Parent | Child | Strategy |
|--------|-------|----------|
| compute | web_app | equal |

### Allocation Result
| Node | Direct | Indirect | Holistic |
|------|--------|----------|----------|
| compute | $100 | $0 | $100 |
| web_app | $0 | $100 | $100 |

### Invariants Check
- Raw Infra Cost: $100
- Allocated Product Cost: $100 (web_app is final cost centre)
- Unallocated: $0
- Coverage: 100%

---

## Example 2: Shared Service → Multiple Products (Equal)

### Scenario
A shared database cluster allocates costs equally to three products.

### Graph
```
                              ┌─────────────────┐
                         ┌───▶│   "web_app"     │
                         │    │   $0 direct     │
┌─────────────────┐      │    └─────────────────┘
│     Shared      │      │    ┌─────────────────┐
│   "database"    │──────┼───▶│     "api"       │
│   $300/day      │ equal│    │   $0 direct     │
└─────────────────┘      │    └─────────────────┘
                         │    ┌─────────────────┐
                         └───▶│    "mobile"     │
                              │   $0 direct     │
                              └─────────────────┘
```

### Input Data
| Node | Type | Direct Cost |
|------|------|-------------|
| database | shared | $300/day |
| web_app | product | $0/day |
| api | product | $0/day |
| mobile | product | $0/day |

### Allocation Result
| Node | Direct | Indirect | Holistic |
|------|--------|----------|----------|
| database | $300 | $0 | $300 |
| web_app | $0 | $100 | $100 |
| api | $0 | $100 | $100 |
| mobile | $0 | $100 | $100 |

### Invariants Check
- Raw Infra Cost: $300
- Allocated Product Cost: $300 ($100 × 3 final cost centres)
- Unallocated: $0
- Coverage: 100%

---

## Example 3: Platform → Products (Proportional on CPU)

### Scenario
A Kubernetes platform allocates costs proportionally based on CPU usage.

### Graph
```
                                    ┌─────────────────┐
                               ┌───▶│   "web_app"     │
                               │    │ 1000 CPU-hrs    │
┌─────────────────┐            │    └─────────────────┘
│    Platform     │ proportional    ┌─────────────────┐
│  "kubernetes"   │────────────┼───▶│     "api"       │
│   $500/day      │  on: cpu   │    │ 4000 CPU-hrs    │
└─────────────────┘            │    └─────────────────┘
```

### Input Data
| Node | Type | Direct Cost | CPU Usage |
|------|------|-------------|-----------|
| kubernetes | platform | $500/day | - |
| web_app | product | $0/day | 1000 hrs |
| api | product | $0/day | 4000 hrs |

### Share Calculation
- Total CPU: 1000 + 4000 = 5000 hrs
- web_app share: 1000/5000 = 20%
- api share: 4000/5000 = 80%

### Allocation Result
| Node | Direct | Indirect | Holistic |
|------|--------|----------|----------|
| kubernetes | $500 | $0 | $500 |
| web_app | $0 | $100 | $100 |
| api | $0 | $400 | $400 |

### Invariants Check
- Raw Infra Cost: $500
- Allocated Product Cost: $500 ($100 + $400)
- Unallocated: $0
- Coverage: 100%

---

## Example 4: Product → Product Full Roll-up

### Scenario
Product A is fully rolled up into Product B (100% allocation).

### Graph
```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│    Platform     │         │   Product A     │         │   Product B     │
│  "kubernetes"   │────────▶│  "component"    │────────▶│   "suite"       │
│   $500/day      │  equal  │   $0 direct     │ 100%    │   $0 direct     │
└─────────────────┘         └─────────────────┘         └─────────────────┘
```

### Input Data
| Node | Type | Direct Cost |
|------|------|-------------|
| kubernetes | platform | $500/day |
| component | product | $0/day |
| suite | product | $0/day |

### Edge Configuration
| Parent | Child | Strategy | Params |
|--------|-------|----------|--------|
| kubernetes | component | equal | - |
| component | suite | fixed_percent | percent: 100 |

### Allocation Flow
1. kubernetes ($500) → component: $500 (100% of single child)
2. component ($500 holistic) → suite: $500 (100% roll-up)

### Allocation Result
| Node | Direct | Indirect | Holistic | Final Cost Centre? |
|------|--------|----------|----------|-------------------|
| kubernetes | $500 | $0 | $500 | No (infra) |
| component | $0 | $500 | $500 | No (has outgoing edge) |
| suite | $0 | $500 | $500 | Yes |

### Invariants Check
- Raw Infra Cost: $500
- Allocated Product Cost: $500 (only suite is final cost centre)
- Unallocated: $0
- Coverage: 100%
- **No amplification**: Total final cost centre cost ($500) = Raw Infra ($500)

---

## Example 5: Partial and Multi-Parent Roll-ups

### Scenario
Product A rolls up 30% to Product B and 50% to Product C, retaining 20%.

### Graph
```
                                         ┌─────────────────┐
                                    30%  │   Product B     │
                               ┌────────▶│   "enterprise"  │
┌─────────────────┐            │         └─────────────────┘
│   Product A     │────────────┤
│   "base"        │            │         ┌─────────────────┐
│ $1000 holistic  │            └────────▶│   Product C     │
└─────────────────┘               50%    │   "premium"     │
                                         └─────────────────┘
```

### Input Data
Assume Product A has received $1000 from infrastructure allocations.

| Node | Holistic Cost (before roll-up) |
|------|-------------------------------|
| base | $1000 |
| enterprise | $0 |
| premium | $0 |

### Edge Configuration
| Parent | Child | Strategy | Params |
|--------|-------|----------|--------|
| base | enterprise | fixed_percent | percent: 30 |
| base | premium | fixed_percent | percent: 50 |

### Share Calculation
- To enterprise: $1000 × 30% = $300
- To premium: $1000 × 50% = $500
- Residual on base: $1000 × 20% = $200

### Allocation Result
| Node | Holistic | Final Cost Centre? |
|------|----------|-------------------|
| base | $200 (residual) | Yes (partial) |
| enterprise | $300 | Yes |
| premium | $500 | Yes |

### Invariants Check
- Sum of final cost centres: $200 + $300 + $500 = $1000
- **No amplification**: Total equals original holistic cost
- All three products are final cost centres (base has residual)

---

## Summary: Key Invariants

For all examples above, these invariants hold:

1. **Conservation of Cost**
   ```
   AllocatedProductCost + UnallocatedInfraCost = RawInfraCost
   ```

2. **Final Cost Centre Sum**
   ```
   AllocatedProductCost = Σ Holistic(p) for all final cost centre products p
   ```

3. **No Amplification**
   ```
   Sum of all final cost centre costs ≤ Raw Infrastructure Cost
   ```

4. **Share Validity**
   ```
   For each parent: Σ shares to children ≤ 1
   ```

