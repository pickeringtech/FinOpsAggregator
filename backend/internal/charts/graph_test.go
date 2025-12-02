package charts

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphRenderer_RenderNoDataChart(t *testing.T) {
	renderer := &GraphRenderer{}

	var buf bytes.Buffer
	err := renderer.RenderNoDataChart(context.Background(), "Test message", &buf, "png")

	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0, "Chart should generate some data")
}

func TestGraphRenderer_RenderCostTrend_NoData(t *testing.T) {
	// Skip this test as it requires a real store connection
	// The RenderCostTrend method requires a *store.Store which cannot be easily mocked
	t.Skip("Skipping test that requires database connection")
}

func TestGraphRenderer_RenderCostTrend_WithData(t *testing.T) {
	// Skip this test as it requires a real store connection
	// The RenderCostTrend method requires a *store.Store which cannot be easily mocked
	t.Skip("Skipping test that requires database connection")
}

func TestGraphRenderer_UnsupportedFormat(t *testing.T) {
	// The RenderNoDataChart function defaults to PNG for unsupported formats
	// so this test verifies that behavior
	renderer := &GraphRenderer{}

	var buf bytes.Buffer
	err := renderer.RenderNoDataChart(context.Background(), "Test", &buf, "unsupported")

	// Function defaults to PNG for unsupported formats, so no error expected
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0, "Should generate PNG output for unsupported format")
}

func TestSupportedFormats(t *testing.T) {
	renderer := &GraphRenderer{}

	supportedFormats := []string{"png", "svg"}

	for _, format := range supportedFormats {
		t.Run(format, func(t *testing.T) {
			var buf bytes.Buffer
			err := renderer.RenderNoDataChart(context.Background(), "Test", &buf, format)
			assert.NoError(t, err, "Format %s should be supported", format)
			assert.Greater(t, buf.Len(), 0, "Should generate data for format %s", format)
		})
	}
}
