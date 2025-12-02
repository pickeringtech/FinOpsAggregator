package graph

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore for testing
type MockStore struct {
	mock.Mock
}

type MockNodeRepository struct {
	mock.Mock
}

type MockEdgeRepository struct {
	mock.Mock
}

func (m *MockNodeRepository) List(ctx context.Context, filters interface{}) (map[uuid.UUID]*models.CostNode, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*models.CostNode), args.Error(1)
}

func (m *MockEdgeRepository) GetActiveEdgesForDate(ctx context.Context, date time.Time) ([]models.DependencyEdge, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DependencyEdge), args.Error(1)
}

func TestGraph_BasicOperations(t *testing.T) {
	// Create test nodes
	node1 := &models.CostNode{
		ID:   uuid.New(),
		Name: "node1",
		Type: "product",
	}
	node2 := &models.CostNode{
		ID:   uuid.New(),
		Name: "node2",
		Type: "shared_resource",
	}

	nodes := map[uuid.UUID]*models.CostNode{
		node1.ID: node1,
		node2.ID: node2,
	}

	// Create test edge
	edge := models.DependencyEdge{
		ID:                uuid.New(),
		ParentID:          node2.ID, // shared_resource -> product
		ChildID:           node1.ID,
		DefaultStrategy:   "proportional_on",
		DefaultParameters: map[string]interface{}{"dimension": "instance_hours"},
		ActiveFrom:        time.Now().AddDate(0, 0, -1),
		ActiveTo:          nil,
	}

	// Create graph using helper function
	g := createGraphFromEdges(nodes, []models.DependencyEdge{edge})

	t.Run("nodes access", func(t *testing.T) {
		retrievedNodes := g.Nodes()
		assert.Equal(t, 2, len(retrievedNodes))
		assert.Contains(t, retrievedNodes, node1.ID)
		assert.Contains(t, retrievedNodes, node2.ID)
	})

	t.Run("edge access", func(t *testing.T) {
		parentEdges := g.GetOutgoingEdges(node2.ID)
		assert.Equal(t, 1, len(parentEdges))
		assert.Equal(t, edge.ChildID, parentEdges[0].ChildID)

		childEdges := g.GetIncomingEdges(node1.ID)
		assert.Equal(t, 1, len(childEdges))
		assert.Equal(t, edge.ParentID, childEdges[0].ParentID)
	})

	t.Run("roots and leaves", func(t *testing.T) {
		roots := g.GetRoots()
		assert.Equal(t, 1, len(roots))
		assert.Contains(t, roots, node2.ID) // shared_resource has no incoming edges

		leaves := g.GetLeaves()
		assert.Equal(t, 1, len(leaves))
		assert.Contains(t, leaves, node1.ID) // product has no outgoing edges
	})
}

func TestGraph_CycleDetection(t *testing.T) {
	// Create nodes for cycle test
	nodeA := &models.CostNode{ID: uuid.New(), Name: "A", Type: "product"}
	nodeB := &models.CostNode{ID: uuid.New(), Name: "B", Type: "product"}
	nodeC := &models.CostNode{ID: uuid.New(), Name: "C", Type: "product"}

	nodes := map[uuid.UUID]*models.CostNode{
		nodeA.ID: nodeA,
		nodeB.ID: nodeB,
		nodeC.ID: nodeC,
	}

	t.Run("no cycle", func(t *testing.T) {
		// A -> B -> C (no cycle)
		edges := []models.DependencyEdge{
			{ID: uuid.New(), ParentID: nodeA.ID, ChildID: nodeB.ID},
			{ID: uuid.New(), ParentID: nodeB.ID, ChildID: nodeC.ID},
		}

		g := createGraphFromEdges(nodes, edges)
		hasCycle := g.HasCycle()
		assert.False(t, hasCycle, "Graph should not have a cycle")
	})

	t.Run("with cycle", func(t *testing.T) {
		// A -> B -> C -> A (cycle)
		edges := []models.DependencyEdge{
			{ID: uuid.New(), ParentID: nodeA.ID, ChildID: nodeB.ID},
			{ID: uuid.New(), ParentID: nodeB.ID, ChildID: nodeC.ID},
			{ID: uuid.New(), ParentID: nodeC.ID, ChildID: nodeA.ID},
		}

		g := createGraphFromEdges(nodes, edges)
		hasCycle := g.HasCycle()
		assert.True(t, hasCycle, "Graph should have a cycle")
	})
}

func TestGraph_TopologicalSort(t *testing.T) {
	// Create a simple DAG: A -> B -> C, A -> C
	nodeA := &models.CostNode{ID: uuid.New(), Name: "A", Type: "platform"}
	nodeB := &models.CostNode{ID: uuid.New(), Name: "B", Type: "shared_resource"}
	nodeC := &models.CostNode{ID: uuid.New(), Name: "C", Type: "product"}

	nodes := map[uuid.UUID]*models.CostNode{
		nodeA.ID: nodeA,
		nodeB.ID: nodeB,
		nodeC.ID: nodeC,
	}

	edges := []models.DependencyEdge{
		{ID: uuid.New(), ParentID: nodeA.ID, ChildID: nodeB.ID},
		{ID: uuid.New(), ParentID: nodeB.ID, ChildID: nodeC.ID},
		{ID: uuid.New(), ParentID: nodeA.ID, ChildID: nodeC.ID},
	}

	g := createGraphFromEdges(nodes, edges)

	order, err := g.TopologicalSort()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(order))

	// Verify topological order: A should come before B and C, B should come before C
	posA := findPosition(order, nodeA.ID)
	posB := findPosition(order, nodeB.ID)
	posC := findPosition(order, nodeC.ID)

	assert.True(t, posA < posB, "A should come before B")
	assert.True(t, posA < posC, "A should come before C")
	assert.True(t, posB < posC, "B should come before C")
}

func TestGraph_Statistics(t *testing.T) {
	// Create test graph with deterministic node IDs
	nodeIDs := make([]uuid.UUID, 5)
	nodes := make(map[uuid.UUID]*models.CostNode)
	for i := 0; i < 5; i++ {
		id := uuid.New()
		nodeIDs[i] = id
		nodes[id] = &models.CostNode{
			ID:   id,
			Name: fmt.Sprintf("node%d", i),
			Type: "product",
		}
	}

	// Create edges: 0 -> 1 -> 2 -> 3, node 4 is isolated
	edges := []models.DependencyEdge{
		{ID: uuid.New(), ParentID: nodeIDs[0], ChildID: nodeIDs[1]},
		{ID: uuid.New(), ParentID: nodeIDs[1], ChildID: nodeIDs[2]},
		{ID: uuid.New(), ParentID: nodeIDs[2], ChildID: nodeIDs[3]},
	}

	g := createGraphFromEdges(nodes, edges)
	stats := g.GetStatistics()

	assert.Equal(t, 5, stats.NodeCount)
	assert.Equal(t, 3, stats.EdgeCount)
	assert.Equal(t, 2, stats.RootCount)    // nodes 0 and 4 have no incoming edges
	assert.Equal(t, 2, stats.LeafCount)    // nodes 3 and 4 have no outgoing edges
	assert.Equal(t, 4, stats.MaxDepth)     // longest path is 4 nodes deep (0->1->2->3)
	assert.False(t, stats.HasCycles)
}

// Helper functions for tests

func createGraphFromEdges(nodes map[uuid.UUID]*models.CostNode, edges []models.DependencyEdge) *Graph {
	g := &Graph{
		nodes:    nodes,
		edges:    make(map[uuid.UUID][]models.DependencyEdge),
		incoming: make(map[uuid.UUID][]models.DependencyEdge),
		date:     time.Now(),
	}

	for _, edge := range edges {
		g.edges[edge.ParentID] = append(g.edges[edge.ParentID], edge)
		g.incoming[edge.ChildID] = append(g.incoming[edge.ChildID], edge)
	}

	return g
}

func findPosition(slice []uuid.UUID, item uuid.UUID) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func getNodeID(nodes map[uuid.UUID]*models.CostNode, index int) uuid.UUID {
	i := 0
	for id := range nodes {
		if i == index {
			return id
		}
		i++
	}
	return uuid.Nil
}
