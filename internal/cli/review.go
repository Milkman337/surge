package cli

import (
	"fmt"

	"github.com/AtomicWasTaken/surge/internal/ai"
	"github.com/AtomicWasTaken/surge/internal/config"
	"github.com/AtomicWasTaken/surge/internal/github"
	"github.com/AtomicWasTaken/surge/internal/review"
	"github.com/spf13/cobra"
)

func runReview(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	applyReviewFlagOverrides(cmd, cfg)

	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Apply explicit flag overrides
	if cmd.Flags().Changed("verbose") {
		cfg.Verbose = flagVerbose
	}
	_ = cmd // mark as used

	// Apply --no-inline and --no-summary flags
	if flagNoInline {
		cfg.DisableInlineComments = true
	}
	if flagNoSummary {
		cfg.DisableSummaryComment = true
	}

	// Detect owner/repo from git if not set
	owner := cfg.GitHub.Owner
	repo := cfg.GitHub.Repo
	if owner == "" || repo == "" {
		detectedOwner, detectedRepo, err := detectGitInfo()
		if err != nil {
			return fmt.Errorf("could not detect owner/repo: %w (use --owner and --repo flags, or set in config)", err)
		}
		if owner == "" {
			owner = detectedOwner
		}
		if repo == "" {
			repo = detectedRepo
		}
	}

	// Get PR number
	prNumber := cfg.GitHub.PRNumber
	if prNumber <= 0 {
		return fmt.Errorf("PR number is required (use --pr flag or set github.prNumber in config)")
	}

	if cfg.Verbose {
		fmt.Printf("[debug] review config owner=%s repo=%s pr=%d provider=%s model=%s base_url=%s context=%s dry_run=%t\n",
			owner, repo, prNumber, cfg.AI.Provider, cfg.AI.Model, cfg.AI.BaseURL, cfg.ContextDepth, flagDryRun)
	}

	// Check for required tokens
	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token is required (set SURGE_GITHUB_TOKEN env var, or use --github-token flag)")
	}
	if cfg.AI.APIKey == "" && cfg.AI.Provider == "claude" {
		return fmt.Errorf("AI API key is required (set SURGE_AI_API_KEY env var, or use --ai-api-key flag)")
	}

	// Create AI client
	var aiClient ai.AIClient
	switch cfg.AI.Provider {
	case "litellm":
		aiClient = ai.NewLiteLLMClient(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Model)
	case "claude":
		aiClient = ai.NewClaudeClient(cfg.AI.APIKey, cfg.AI.Model)
	default:
		return fmt.Errorf("unknown AI provider: %s", cfg.AI.Provider)
	}

	// Create GitHub client
	ghClient := github.NewGitHubClient(cfg.GitHub.Token)

	// Create orchestrator
	orch := review.NewOrchestrator(aiClient, ghClient, cfg)

	// Run the review
	result, err := orch.Review(cmd.Context(), owner, repo, prNumber, flagDryRun)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Print final status
	if flagDryRun {
		fmt.Println("\n[DRY RUN] No changes posted to the PR.")
	} else {
		fmt.Printf("\nReview posted to %s/%s#%d\n", owner, repo, prNumber)
	}

	// Exit code based on approval
	if !result.Approve && !flagDryRun {
		return nil // Don't exit with error, just report
	}

	return nil
}

func applyReviewFlagOverrides(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("github-token") {
		cfg.GitHub.Token, _ = cmd.Flags().GetString("github-token")
	}
	if cmd.Flags().Changed("owner") {
		cfg.GitHub.Owner, _ = cmd.Flags().GetString("owner")
	}
	if cmd.Flags().Changed("repo") {
		cfg.GitHub.Repo, _ = cmd.Flags().GetString("repo")
	}
	if cmd.Flags().Changed("pr") {
		cfg.GitHub.PRNumber, _ = cmd.Flags().GetInt("pr")
	}
	if cmd.Flags().Changed("ai-provider") {
		cfg.AI.Provider, _ = cmd.Flags().GetString("ai-provider")
	}
	if cmd.Flags().Changed("ai-model") {
		cfg.AI.Model, _ = cmd.Flags().GetString("ai-model")
	}
	if cmd.Flags().Changed("ai-base-url") {
		cfg.AI.BaseURL, _ = cmd.Flags().GetString("ai-base-url")
	}
	if cmd.Flags().Changed("ai-api-key") {
		cfg.AI.APIKey, _ = cmd.Flags().GetString("ai-api-key")
	}
	if cmd.Flags().Changed("context-depth") {
		cfg.ContextDepth, _ = cmd.Flags().GetString("context-depth")
	}
	if cmd.Flags().Changed("output") {
		cfg.Output.Format, _ = cmd.Flags().GetString("output")
	}
	if cmd.Flags().Changed("max-inline") {
		cfg.MaxInlineComments, _ = cmd.Flags().GetInt("max-inline")
	}
	if cmd.Flags().Changed("max-tokens") {
		cfg.MaxTokens, _ = cmd.Flags().GetInt("max-tokens")
	}
	if cmd.Flags().Changed("temperature") {
		cfg.Temperature, _ = cmd.Flags().GetFloat64("temperature")
	}
}
