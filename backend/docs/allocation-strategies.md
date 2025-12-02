# Allocation Strategy Catalogue

This document provides detailed specifications for all allocation strategies in the FinOps Aggregator system.

## 1. Implemented Strategies

### 1.1 `equal` (`StrategyEqual`)

**Description:** Parent cost is split equally among all children.

**Parameters:** None

**Formula:** `share = 1 / N` where N = number of children

**Example:**
```
Shared DB ($300/day) → [Product A, Product B, Product C]
Result: Each product receives $100/day
```

### 1.2 `proportional_on` (`StrategyProportionalOn`)

**Description:** Parent cost is split proportionally based on a usage metric.

**Parameters:**
- `metric` (string, required): The usage metric to use (e.g. `cpu_hours`, `requests`)

**Formula:** `share_i = usage_i / Σ usage_j` for all children j

**Example:**
```
Platform K8s ($500/day) → [Product A (1000 CPU-hrs), Product B (4000 CPU-hrs)]
Result: A receives $100/day (20%), B receives $400/day (80%)
```

### 1.3 `fixed_percent` (`StrategyFixedPercent`)

**Description:** A fixed percentage of the parent's cost is allocated to the child.

**Parameters:**
- `percent` (float, required): The percentage to allocate (0-100 or 0-1)

**Formula:** `share = percent / 100` (if percent > 1)

**Example:**
```
Shared Service ($1000/day) → Product A (fixed_percent: 25)
Result: A receives $250/day
```

### 1.4 `capped_proportional` (`StrategyCappedProp`)

**Description:** Proportional allocation with a per-child cap.

**Parameters:**
- `metric` (string, required): The usage metric to use
- `cap` (float, required): Maximum share per child (0-100 or 0-1)

**Formula:** `share = min(proportional_share, cap)`

**Example:**
```
Platform ($1000/day) → [A (90% usage), B (10% usage)] with cap=50%
Result: A receives $500/day (capped), B receives $100/day
Note: $400 remains unallocated due to cap
```

### 1.5 `residual_to_max` (`StrategyResidualToMax`)

**Description:** Assigns the residual share to the parent with maximum usage.

**Parameters:**
- `metric` (string, required): The usage metric to use

**Example:**
```
Multiple parents allocating to same child:
Parent A (100 usage) → Child X
Parent B (400 usage) → Child X
Result: B (max usage) receives residual allocation
```

## 2. Planned Strategies

### 2.1 `weighted_average` (`StrategyWeightedAverage`)

**Description:** Uses a look-back window over usage metrics to compute smoothed proportional shares.

**Parameters:**
- `metric` (string, required): The usage metric to use
- `window_days` (int, required): Number of days to look back
- `decay` (float, optional): Exponential decay factor (default: 1.0 = no decay)

**Formula:** `share_i = weighted_avg(usage_i, window) / Σ weighted_avg(usage_j, window)`

**Example:**
```
Platform ($1000/day) with window_days=7:
Product A: Day usage [100, 200, 150, 180, 160, 140, 170] → avg 157
Product B: Day usage [300, 100, 350, 120, 340, 160, 330] → avg 243
Result: A receives $392.50/day (39.25%), B receives $607.50/day (60.75%)
```

### 2.2 `hybrid_fixed_proportional` (`StrategyHybridFixedProp`)

**Description:** Splits cost into a fixed baseline portion (equal across children) and a variable portion (proportional to usage).

**Parameters:**
- `metric` (string, required): The usage metric for proportional portion
- `fixed_percent` (float, required): Percentage allocated equally (0-100)

**Formula:**
- `fixed_share = fixed_percent / (100 * N)` per child
- `variable_share = (1 - fixed_percent/100) * (usage_i / Σ usage_j)`
- `total_share = fixed_share + variable_share`

**Example:**
```
Platform ($1000/day) with fixed_percent=40:
Children: [A (100 usage), B (300 usage), C (0 usage)]
Fixed portion ($400): A=$133.33, B=$133.33, C=$133.33
Variable portion ($600): A=$150 (25%), B=$450 (75%), C=$0 (0%)
Result: A=$283.33, B=$583.33, C=$133.33
```

### 2.3 `min_floor_proportional` (`StrategyMinFloorProp`)

**Description:** Guarantees each child a minimum share, then distributes remainder proportionally.

**Parameters:**
- `metric` (string, required): The usage metric for proportional portion
- `min_floor_percent` (float, required): Minimum percentage per child (0-100)

**Formula:**
- If `N * min_floor_percent >= 100`: each child gets `100/N`%
- Otherwise:
  - `floor_share = min_floor_percent / 100` per child
  - `remainder = 1 - (N * min_floor_percent / 100)`
  - `proportional_share = remainder * (usage_i / Σ usage_j)`
  - `total_share = floor_share + proportional_share`

**Example:**
```
Platform ($1000/day) with min_floor_percent=10:
Children: [A (100 usage), B (300 usage), C (0 usage)]
Floor: A=$100, B=$100, C=$100 (30% total)
Remainder ($700): A=$175 (25%), B=$525 (75%), C=$0 (0%)
Result: A=$275, B=$625, C=$100
```

## 3. Strategy Selection Guidelines

| Scenario | Recommended Strategy | Rationale |
|----------|---------------------|-----------|
| Unknown usage patterns | `equal` | Fair baseline when no data available |
| Clear usage metrics | `proportional_on` | Direct correlation to resource consumption |
| Fixed contractual splits | `fixed_percent` | Matches business agreements |
| Prevent single-tenant dominance | `capped_proportional` | Ensures fair distribution |
| Noisy daily metrics | `weighted_average` | Smooths out spikes |
| Base fee + usage model | `hybrid_fixed_proportional` | Matches SaaS pricing models |
| Minimum viable allocation | `min_floor_proportional` | Ensures all products carry baseline |

## 4. Implementation Notes

### 4.1 Share Validation

All strategies must ensure:
- Individual shares are in range [0, 1]
- Sum of shares for a parent's children should be <= 1
- Sum of shares > 1 indicates a configuration error

### 4.2 Zero Usage Handling

When total usage is zero:
- `proportional_on`: Falls back to `equal` allocation
- `weighted_average`: Falls back to `equal` allocation
- `hybrid_fixed_proportional`: Only fixed portion is allocated
- `min_floor_proportional`: Only floor portion is allocated

### 4.3 Edge Cases

- **Single child**: All strategies allocate 100% to the single child
- **No children**: No allocation occurs (parent retains cost)
- **Negative usage**: Treated as zero (logged as warning)

