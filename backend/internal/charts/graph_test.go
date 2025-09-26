package charts

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of the store interface for testing
type MockStore struct {
	mock.Mock
}

type MockNodeRepository struct {
	mock.Mock
}

type MockCostRepository struct {
	mock.Mock
}

func (m *MockNodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CostNode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CostNode), args.Error(1)
}

func (m *MockNodeRepository) GetByName(ctx context.Context, name string) (*models.CostNode, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CostNode), args.Error(1)
}

func (m *MockCostRepository) GetByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.NodeCostByDimension, error) {
	args := m.Called(ctx, nodeID, startDate, endDate, dimensions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.NodeCostByDimension), args.Error(1)
}

// MockGraphBuilder for testing
type MockGraphBuilder struct {
	mock.Mock
}

type MockGraph struct {
	nodes map[uuid.UUID]*models.CostNode
}

func (mg *MockGraph) Nodes() map[uuid.UUID]*models.CostNode {
	return mg.nodes
}

func TestGraphRenderer_RenderNoDataChart(t *testing.T) {
	renderer := &GraphRenderer{}
	
	var buf bytes.Buffer
	err := renderer.renderNoDataChart(context.Background(), "Test message", &buf, "png")
	
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0, "Chart should generate some data")
}

func TestGraphRenderer_RenderCostTrend_NoData(t *testing.T) {
	// Create mock store
	mockNodeRepo := &MockNodeRepository{}
	mockCostRepo := &MockCostRepository{}
	
	// Create a mock store struct that contains the repositories
	store := &struct {
		Nodes MockNodeRepository
		Costs MockCostRepository
	}{
		Nodes: *mockNodeRepo,
		Costs: *mockCostRepo,
	}
	
	renderer := &GraphRenderer{store: store}
	
	nodeID := uuid.New()
	testNode := &models.CostNode{
		ID:   nodeID,
		Name: "test-node",
		Type: "product",
	}
	
	// Set up expectations
	mockNodeRepo.On("GetByID", mock.Anything, nodeID).Return(testNode, nil)
	mockCostRepo.On("GetByNodeAndDateRange", mock.Anything, nodeID, mock.Anything, mock.Anything, mock.Anything).Return([]models.NodeCostByDimension{}, nil)
	
	var buf bytes.Buffer
	err := renderer.RenderCostTrend(
		context.Background(),
		nodeID,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		"instance_hours",
		&buf,
		"png",
	)
	
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0, "Should generate a no-data chart")
	
	mockNodeRepo.AssertExpectations(t)
	mockCostRepo.AssertExpectations(t)
}

func TestGraphRenderer_RenderCostTrend_WithData(t *testing.T) {
	// Create mock store
	mockNodeRepo := &MockNodeRepository{}
	mockCostRepo := &MockCostRepository{}
	
	// Create a mock store struct that contains the repositories
	store := &struct {
		Nodes MockNodeRepository
		Costs MockCostRepository
	}{
		Nodes: *mockNodeRepo,
		Costs: *mockCostRepo,
	}
	
	renderer := &GraphRenderer{store: store}
	
	nodeID := uuid.New()
	testNode := &models.CostNode{
		ID:   nodeID,
		Name: "test-node",
		Type: "product",
	}
	
	// Create test cost data
	testCosts := []models.NodeCostByDimension{
		{
			NodeID:    nodeID,
			CostDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Dimension: "instance_hours",
			Amount:    decimal.NewFromFloat(100.0),
			Currency:  "USD",
		},
		{
			NodeID:    nodeID,
			CostDate:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Dimension: "instance_hours",
			Amount:    decimal.NewFromFloat(150.0),
			Currency:  "USD",
		},
	}
	
	// Set up expectations
	mockNodeRepo.On("GetByID", mock.Anything, nodeID).Return(testNode, nil)
	mockCostRepo.On("GetByNodeAndDateRange", mock.Anything, nodeID, mock.Anything, mock.Anything, mock.Anything).Return(testCosts, nil)
	
	var buf bytes.Buffer
	err := renderer.RenderCostTrend(
		context.Background(),
		nodeID,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		"instance_hours",
		&buf,
		"png",
	)
	
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0, "Should generate a chart with data")
	
	mockNodeRepo.AssertExpectations(t)
	mockCostRepo.AssertExpectations(t)
}

func TestGraphRenderer_UnsupportedFormat(t *testing.T) {
	renderer := &GraphRenderer{}
	
	var buf bytes.Buffer
	err := renderer.renderNoDataChart(context.Background(), "Test", &buf, "unsupported")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestSupportedFormats(t *testing.T) {
	renderer := &GraphRenderer{}
	
	supportedFormats := []string{"png", "svg"}
	
	for _, format := range supportedFormats {
		t.Run(format, func(t *testing.T) {
			var buf bytes.Buffer
			err := renderer.renderNoDataChart(context.Background(), "Test", &buf, format)
			assert.NoError(t, err, "Format %s should be supported", format)
			assert.Greater(t, buf.Len(), 0, "Should generate data for format %s", format)
		})
	}
}
