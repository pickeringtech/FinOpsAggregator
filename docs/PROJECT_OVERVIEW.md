# FinOps Aggregator: Project Overview

## Executive Summary

FinOps Aggregator is a **prototype platform** designed to explore and validate approaches to cloud cost attribution, allocation, and financial visibility. It addresses a fundamental challenge in modern cloud-native organisations: understanding the true cost of products, services, and features when infrastructure is shared, dependencies are complex, and costs flow through multiple layers of abstraction.

This document outlines the nature, purpose, and value of the project, and establishes its role as an early-stage prototype intended to surface challenges and inform future investment decisions.

---

## The Problem Space

### Why Cost Attribution is Hard

Modern cloud infrastructure creates a fundamental disconnect between where costs are incurred and where business value is delivered:

1. **Shared Infrastructure**: Kubernetes clusters, databases, message queues, and observability platforms serve multiple products simultaneously. A single database instance might support five different products, but the cloud bill shows only one line item.

2. **Complex Dependencies**: A customer-facing product might depend on an API gateway, which depends on a service mesh, which depends on a Kubernetes platform, which depends on compute instances. Each layer adds cost, but tracing that cost back to the product requires understanding the entire dependency chain.

3. **Multi-Dimensional Costs**: Cloud costs aren't monolithic. They span compute hours, storage gigabytes, network egress, API calls, and dozens of other dimensions—each requiring different allocation logic.

4. **Dynamic Relationships**: Dependencies change over time. New products are launched, services are deprecated, and usage patterns shift. Static allocation models quickly become stale.

5. **Financial Accountability Gap**: Finance teams need product-level P&L statements. Engineering teams see infrastructure costs. The translation between these views is often manual, error-prone, and delayed.

### The Business Impact

Without accurate cost attribution:

- **Product profitability is unknown**: Is that new feature actually profitable, or is it consuming more infrastructure than it generates in revenue?
- **Optimisation is blind**: Teams can't prioritise cost reduction efforts without knowing which products drive which costs.
- **Budgeting is guesswork**: Forecasting product costs requires understanding how infrastructure costs flow to products.
- **Accountability is diffuse**: When no one owns a cost, no one optimises it.

---

## The Solution Approach

### DAG-Based Cost Attribution

FinOps Aggregator models cost relationships as a **Directed Acyclic Graph (DAG)**, where:

- **Nodes** represent cost entities: resources, shared services, platform components, and products
- **Edges** represent cost flow relationships with configurable allocation strategies
- **Costs flow from infrastructure to products**, following the natural direction of dependency

This approach enables:

- **Holistic cost visibility**: See both direct costs (what a node incurs) and indirect costs (what flows to it from dependencies)
- **Flexible allocation strategies**: Equal splits, proportional allocation based on usage metrics, fixed percentages, capped allocations, and more
- **Multi-dimensional tracking**: Handle different cost types (compute, storage, network) with appropriate allocation logic for each
- **Temporal analysis**: Track how costs and allocations change over time

### Key Concepts

| Concept | Definition |
|---------|------------|
| **Direct Cost** | Cost that originates on a node (e.g., cloud bill for a resource) |
| **Indirect Cost** | Cost received from parent nodes via allocation |
| **Holistic Cost** | Direct + Indirect cost—the true total cost of a node |
| **Final Cost Centre** | A product node with no further outgoing allocations—where costs ultimately land |
| **Coverage** | Percentage of infrastructure cost successfully allocated to products |

### Allocation Strategies

The system supports multiple allocation strategies to handle different cost distribution scenarios:

- **Equal**: Split costs evenly among all children
- **Proportional**: Distribute based on a usage metric (CPU hours, requests, storage)
- **Fixed Percent**: Allocate a specific percentage to each child
- **Capped Proportional**: Proportional with upper bounds
- **Residual to Max**: Allocate remainder to the highest consumer

---

## Prototype Status and Purpose

### This is a Prototype

FinOps Aggregator is explicitly designed as a **prototype and proof-of-concept**, not a production-ready system. Its primary purposes are:

1. **Validate the DAG-based approach**: Does modelling cost relationships as a graph provide the flexibility and accuracy needed for real-world scenarios?

2. **Surface data challenges early**: What data is required? What's available? What's the gap? How do we handle missing or inconsistent data?

3. **Explore integration patterns**: How do we ingest cost data from cloud providers? How do we capture usage metrics for proportional allocation? What observability data is needed?

4. **Inform UX requirements**: What views do finance teams need? What do engineering teams need? How do we present complex allocation logic in an understandable way?

5. **Identify edge cases**: What happens with circular dependencies? Partial allocations? Missing usage data? Negative costs (credits)?

### What This Prototype Demonstrates

- **Core allocation engine**: Processes cost data through the DAG and computes holistic costs
- **Multiple allocation strategies**: Implements the most common distribution patterns
- **API-first design**: RESTful API for integration with other systems
- **Interactive frontend**: Dashboard for exploring product hierarchies and cost breakdowns
- **Seed data generation**: Realistic demo data for testing and demonstration

### What This Prototype Does Not Include

- Production-grade security and authentication
- High-availability deployment configurations
- Comprehensive error handling for all edge cases
- Performance optimisation for very large datasets
- Complete integration with cloud provider billing APIs
- Audit logging and compliance features

---

## Value Proposition

### For Finance Teams

- **Product-level cost visibility**: See the true cost of each product, including allocated shared infrastructure
- **Margin analysis**: Understand gross margin by product when combined with revenue data
- **Budget accuracy**: Base forecasts on actual cost attribution, not estimates

### For Engineering Teams

- **Cost-aware architecture**: Understand the cost implications of technical decisions
- **Optimisation targeting**: Identify which services drive the most cost to which products
- **Dependency visibility**: See how infrastructure costs cascade through the system

### For Platform Teams

- **Showback/chargeback**: Provide accurate cost data for internal billing
- **Capacity planning**: Understand how product growth affects infrastructure costs
- **Efficiency metrics**: Track cost per request, cost per user, and other unit economics

---

## Technical Architecture

### Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (Next.js)                       │
│  Dashboard │ Product Hierarchy │ Cost Explorer │ Recommendations │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Backend (Go)                             │
│  REST API │ Allocation Engine │ Analysis │ Chart Generation     │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                       PostgreSQL Database                        │
│  Nodes │ Edges │ Costs │ Usage │ Allocation Results             │
└─────────────────────────────────────────────────────────────────┘
```

### Data Model

- **cost_nodes**: Entities in the cost graph (resources, shared services, products)
- **dependency_edges**: Relationships between nodes with allocation strategies
- **node_costs_by_dimension**: Raw cost data per node, date, and dimension
- **node_usage_by_dimension**: Usage metrics for proportional allocation
- **allocation_results_by_dimension**: Computed allocation results

---

## Insights and Challenges Discovered

Through building this prototype, several key insights have emerged:

### Data Quality is Critical

Cost attribution is only as good as the input data. Challenges include:
- Incomplete tagging of cloud resources
- Missing usage metrics for proportional allocation
- Inconsistent granularity between cost and usage data
- Delayed availability of billing data

### Graph Complexity Grows Quickly

Real-world dependency graphs are more complex than anticipated:
- Products depend on products (roll-ups, bundles)
- Shared services depend on other shared services
- Allocation strategies vary by dimension within the same edge

### Financial Invariants Must Be Preserved

The system must guarantee:
- **Conservation**: All infrastructure cost is accounted for (allocated + unallocated = total)
- **No amplification**: Allocated cost never exceeds raw infrastructure cost
- **Auditability**: Every allocation decision can be traced and explained

### User Experience Matters

Technical accuracy is necessary but not sufficient:
- Finance users need familiar views (P&L, cost centres)
- Engineering users need actionable insights (what to optimise)
- Both need confidence in the numbers

---

## Next Steps

This prototype provides a foundation for evaluating whether to invest further in this approach. Recommended next steps:

1. **Pilot with real data**: Test with actual cloud billing and usage data from a subset of products
2. **Validate with stakeholders**: Review outputs with finance and engineering teams
3. **Identify integration requirements**: Determine what data sources and systems need to connect
4. **Assess build vs. buy**: Compare this approach with commercial FinOps platforms
5. **Define production requirements**: Security, scale, reliability, and compliance needs

---

## Conclusion

FinOps Aggregator represents an exploration of how DAG-based cost attribution can provide the financial visibility that modern cloud-native organisations need. As a prototype, it prioritises learning and validation over production readiness.

The insights gained from this work—about data requirements, allocation complexity, and user needs—will inform decisions about how to achieve accurate, actionable cost attribution at scale, whether through continued development of this platform or through alternative approaches.

---

*This document describes a prototype system. Features, approaches, and conclusions are subject to change as the project evolves.*

