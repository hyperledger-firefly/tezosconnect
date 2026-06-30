//go:build docs

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGenerateConfigDocs(t *testing.T) {
	// Initialize config of all plugins
	InitConfig()
	f, err := os.Create(filepath.Join("..", "config.md"))
	assert.NoError(t, err)
	generatedConfig, err := config.GenerateConfigMarkdown(context.Background(), "", config.GetKnownKeys())
	assert.NoError(t, err)
	_, err = f.Write(generatedConfig)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
}
