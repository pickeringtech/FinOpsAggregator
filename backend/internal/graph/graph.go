package graph

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
)

// Graph represents the cost attribution DAG
type Graph struct {
	nodes     map[uuid.UUID]*models.CostNode
	edges     map[uuid.UUID][]models.DependencyEdge // parent_id -> []edges
	incoming  map[uuid.UUID][]models.DependencyEdge // child_id -> []edges
	date      time.Time
	hash      string
}

// GraphBuilder builds a graph for a specific date
type GraphBuilder struct {
	store *store.Store
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(store *store.Store) *GraphBuilder {
	return &GraphBuilder{
		store: store,
	}
}

// BuildForDate builds a graph for a specific date
func (gb *GraphBuilder) BuildForDate(ctx context.Context, date time.Time) (*Graph, error) {
	log.Info().Time("date", date).Msg("Building graph for date")

	// Get all active nodes (non-archived)
	nodes, err := gb.store.Nodes.List(ctx, store.NodeFilters{
		IncludeArchived: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get all active edges for the date
	edges, err := gb.store.Edges.GetActiveEdgesForDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	// Build the graph
	g := &Graph{
		nodes:    make(map[uuid.UUID]*models.CostNode),
		edges:    make(map[uuid.UUID][]models.DependencyEdge),
		incoming: make(map[uuid.UUID][]models.DependencyEdge),
		date:     date,
	}

	// Add nodes
	for i := range nodes {
		g.nodes[nodes[i].ID] = &nodes[i]
	}

	// Add edges
	for _, edge := range edges {
		// Verify both nodes exist
		if _, exists := g.nodes[edge.ParentID]; !exists {
			log.Warn().
				Str("parent_id", edge.ParentID.String()).
				Str("edge_id", edge.ID.String()).
				Msg("Edge references non-existent parent node")
			continue
		}
		if _, exists := g.nodes[edge.ChildID]; !exists {
			log.Warn().
				Str("child_id", edge.ChildID.String()).
				Str("edge_id", edge.ID.String()).
				Msg("Edge references non-existent child node")
			continue
		}

		g.edges[edge.ParentID] = append(g.edges[edge.ParentID], edge)
		g.incoming[edge.ChildID] = append(g.incoming[edge.ChildID], edge)
	}

	// Calculate graph hash
	g.hash = g.calculateHash()

	log.Info().
		Int("nodes", len(g.nodes)).
		Int("edges", len(edges)).
		Str("hash", g.hash).
		Msg("Graph built successfully")

	return g, nil
}

// Nodes returns all nodes in the graph
func (g *Graph) Nodes() map[uuid.UUID]*models.CostNode {
	return g.nodes
}

// Edges returns all outgoing edges for a node
func (g *Graph) Edges(nodeID uuid.UUID) []models.DependencyEdge {
	return g.edges[nodeID]
}

// IncomingEdges returns all incoming edges for a node
func (g *Graph) IncomingEdges(nodeID uuid.UUID) []models.DependencyEdge {
	return g.incoming[nodeID]
}

// Date returns the date this graph was built for
func (g *Graph) Date() time.Time {
	return g.date
}

// Hash returns the graph hash
func (g *Graph) Hash() string {
	return g.hash
}

// ValidateDAG validates that the graph is a valid DAG (no cycles)
func (g *Graph) ValidateDAG() error {
	log.Info().Msg("Validating DAG structure")

	// Use DFS to detect cycles
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if g.hasCycleDFS(nodeID, visited, recStack) {
				return fmt.Errorf("cycle detected in graph")
			}
		}
	}

	log.Info().Msg("DAG validation passed")
	return nil
}

// hasCycleDFS performs DFS to detect cycles
func (g *Graph) hasCycleDFS(nodeID uuid.UUID, visited, recStack map[uuid.UUID]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Visit all children
	for _, edge := range g.edges[nodeID] {
		childID := edge.ChildID
		if !visited[childID] {
			if g.hasCycleDFS(childID, visited, recStack) {
				return true
			}
		} else if recStack[childID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// TopologicalSort returns nodes in topological order
func (g *Graph) TopologicalSort() ([]uuid.UUID, error) {
	log.Debug().Msg("Computing topological sort")

	// Kahn's algorithm
	inDegree := make(map[uuid.UUID]int)
	
	// Initialize in-degree for all nodes
	for nodeID := range g.nodes {
		inDegree[nodeID] = 0
	}
	
	// Calculate in-degrees
	for _, edges := range g.edges {
		for _, edge := range edges {
			inDegree[edge.ChildID]++
		}
	}

	// Find nodes with no incoming edges
	queue := make([]uuid.UUID, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	var result []uuid.UUID
	
	for len(queue) > 0 {
		// Remove node from queue
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// For each child of this node
		for _, edge := range g.edges[nodeID] {
			childID := edge.ChildID
			inDegree[childID]--
			
			// If child has no more incoming edges, add to queue
			if inDegree[childID] == 0 {
				queue = append(queue, childID)
			}
		}
	}

	// Check if all nodes were processed (no cycles)
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("graph contains cycles")
	}

	log.Debug().Int("nodes", len(result)).Msg("Topological sort completed")
	return result, nil
}

// GetRoots returns nodes with no incoming edges (root nodes)
func (g *Graph) GetRoots() []uuid.UUID {
	var roots []uuid.UUID
	for nodeID := range g.nodes {
		if len(g.incoming[nodeID]) == 0 {
			roots = append(roots, nodeID)
		}
	}
	
	// Sort for deterministic output
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].String() < roots[j].String()
	})
	
	return roots
}

// GetLeaves returns nodes with no outgoing edges (leaf nodes)
func (g *Graph) GetLeaves() []uuid.UUID {
	var leaves []uuid.UUID
	for nodeID := range g.nodes {
		if len(g.edges[nodeID]) == 0 {
			leaves = append(leaves, nodeID)
		}
	}
	
	// Sort for deterministic output
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].String() < leaves[j].String()
	})
	
	return leaves
}

// GetAncestors returns all ancestor nodes of a given node
func (g *Graph) GetAncestors(nodeID uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	var ancestors []uuid.UUID
	
	g.getAncestorsDFS(nodeID, visited, &ancestors)
	
	// Sort for deterministic output
	sort.Slice(ancestors, func(i, j int) bool {
		return ancestors[i].String() < ancestors[j].String()
	})
	
	return ancestors
}

func (g *Graph) getAncestorsDFS(nodeID uuid.UUID, visited map[uuid.UUID]bool, ancestors *[]uuid.UUID) {
	for _, edge := range g.incoming[nodeID] {
		parentID := edge.ParentID
		if !visited[parentID] {
			visited[parentID] = true
			*ancestors = append(*ancestors, parentID)
			g.getAncestorsDFS(parentID, visited, ancestors)
		}
	}
}

// GetDescendants returns all descendant nodes of a given node
func (g *Graph) GetDescendants(nodeID uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	var descendants []uuid.UUID
	
	g.getDescendantsDFS(nodeID, visited, &descendants)
	
	// Sort for deterministic output
	sort.Slice(descendants, func(i, j int) bool {
		return descendants[i].String() < descendants[j].String()
	})
	
	return descendants
}

func (g *Graph) getDescendantsDFS(nodeID uuid.UUID, visited map[uuid.UUID]bool, descendants *[]uuid.UUID) {
	for _, edge := range g.edges[nodeID] {
		childID := edge.ChildID
		if !visited[childID] {
			visited[childID] = true
			*descendants = append(*descendants, childID)
			g.getDescendantsDFS(childID, visited, descendants)
		}
	}
}

// calculateHash computes a deterministic hash of the graph structure
func (g *Graph) calculateHash() string {
	hasher := sha256.New()
	
	// Hash date
	hasher.Write([]byte(g.date.Format("2006-01-02")))
	
	// Hash nodes (sorted by ID for determinism)
	nodeIDs := make([]string, 0, len(g.nodes))
	for nodeID := range g.nodes {
		nodeIDs = append(nodeIDs, nodeID.String())
	}
	sort.Strings(nodeIDs)
	
	for _, nodeIDStr := range nodeIDs {
		nodeID := uuid.MustParse(nodeIDStr)
		node := g.nodes[nodeID]
		hasher.Write([]byte(fmt.Sprintf("node:%s:%s:%s", nodeID, node.Name, node.Type)))
	}
	
	// Hash edges (sorted by parent then child for determinism)
	type edgeKey struct {
		parent, child string
	}
	var edgeKeys []edgeKey
	
	for parentID, edges := range g.edges {
		for _, edge := range edges {
			edgeKeys = append(edgeKeys, edgeKey{
				parent: parentID.String(),
				child:  edge.ChildID.String(),
			})
		}
	}
	
	sort.Slice(edgeKeys, func(i, j int) bool {
		if edgeKeys[i].parent != edgeKeys[j].parent {
			return edgeKeys[i].parent < edgeKeys[j].parent
		}
		return edgeKeys[i].child < edgeKeys[j].child
	})
	
	for _, key := range edgeKeys {
		hasher.Write([]byte(fmt.Sprintf("edge:%s:%s", key.parent, key.child)))
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// Stats returns statistics about the graph
func (g *Graph) Stats() GraphStats {
	roots := g.GetRoots()
	leaves := g.GetLeaves()
	
	totalEdges := 0
	for _, edges := range g.edges {
		totalEdges += len(edges)
	}
	
	return GraphStats{
		NodeCount:  len(g.nodes),
		EdgeCount:  totalEdges,
		RootCount:  len(roots),
		LeafCount:  len(leaves),
		Date:       g.date,
		Hash:       g.hash,
	}
}

// GraphStats represents statistics about a graph
type GraphStats struct {
	NodeCount int       `json:"node_count"`
	EdgeCount int       `json:"edge_count"`
	RootCount int       `json:"root_count"`
	LeafCount int       `json:"leaf_count"`
	Date      time.Time `json:"date"`
	Hash      string    `json:"hash"`
}
