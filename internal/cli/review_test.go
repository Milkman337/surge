package cli

import (
	"testing"

	"github.com/AtomicWasTaken/surge/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestApplyReviewFlagOverrides(t *testing.T) {
	cmd := &cobra.Command{Use: "review"}
	cmd.Flags().String("github-token", "", "")
	cmd.Flags().String("owner", "", "")
	cmd.Flags().String("repo", "", "")
	cmd.Flags().Int("pr", 0, "")
	cmd.Flags().String("ai-provider", "", "")
	cmd.Flags().String("ai-model", "", "")
	cmd.Flags().String("ai-base-url", "", "")
	cmd.Flags().String("ai-api-key", "", "")
	cmd.Flags().String("context-depth", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().Int("max-inline", 0, "")
	cmd.Flags().Int("max-tokens", 0, "")
	cmd.Flags().Float64("temperature", 0, "")

	assert.NoError(t, cmd.Flags().Set("github-token", "gh-token"))
	assert.NoError(t, cmd.Flags().Set("owner", "octo"))
	assert.NoError(t, cmd.Flags().Set("repo", "surge"))
	assert.NoError(t, cmd.Flags().Set("pr", "12"))
	assert.NoError(t, cmd.Flags().Set("ai-provider", "litellm"))
	assert.NoError(t, cmd.Flags().Set("ai-model", "model-x"))
	assert.NoError(t, cmd.Flags().Set("ai-base-url", "https://litellm.example"))
	assert.NoError(t, cmd.Flags().Set("ai-api-key", "ai-key"))
	assert.NoError(t, cmd.Flags().Set("context-depth", "relevant"))
	assert.NoError(t, cmd.Flags().Set("output", "json"))
	assert.NoError(t, cmd.Flags().Set("max-inline", "7"))
	assert.NoError(t, cmd.Flags().Set("max-tokens", "4096"))
	assert.NoError(t, cmd.Flags().Set("temperature", "0.6"))

	cfg := &config.Config{}
	applyReviewFlagOverrides(cmd, cfg)

	assert.Equal(t, "gh-token", cfg.GitHub.Token)
	assert.Equal(t, "octo", cfg.GitHub.Owner)
	assert.Equal(t, "surge", cfg.GitHub.Repo)
	assert.Equal(t, 12, cfg.GitHub.PRNumber)
	assert.Equal(t, "litellm", cfg.AI.Provider)
	assert.Equal(t, "model-x", cfg.AI.Model)
	assert.Equal(t, "https://litellm.example", cfg.AI.BaseURL)
	assert.Equal(t, "ai-key", cfg.AI.APIKey)
	assert.Equal(t, "relevant", cfg.ContextDepth)
	assert.Equal(t, "json", cfg.Output.Format)
	assert.Equal(t, 7, cfg.MaxInlineComments)
	assert.Equal(t, 4096, cfg.MaxTokens)
	assert.Equal(t, 0.6, cfg.Temperature)
}
