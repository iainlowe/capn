package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLI_BasicParsing(t *testing.T) {
	var buf bytes.Buffer
	cli := NewCLI()
	cli.SetOutput(&buf)
	
	// Test that we can at least create and parse without running
	args := []string{"status"}
	err := cli.Parse(args)
	
	// Command should run successfully
	assert.NoError(t, err)
}
