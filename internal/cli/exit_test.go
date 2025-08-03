package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLI_WithoutExitOverride(t *testing.T) {
	var buf bytes.Buffer
	cli := NewCLI()
	cli.SetOutput(&buf)
	
	// Remove exit override from CLI temporarily
	cli.SetExitOverride(false)
	
	args := []string{"status"}
	err := cli.Parse(args)
	
	assert.NoError(t, err)
}
