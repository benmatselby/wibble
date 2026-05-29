package cmd

import (
	"fmt"

	"github.com/benmatselby/wibble/pkg/config"
	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/spf13/cobra"
)

// NewCleanCommand creates the clean subcommand which removes feeds and their
// articles from the database if they are no longer defined in the config.
func NewCleanCommand(db *dao.DaoClient) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove feeds and articles not defined in the config",
		Long:  "Compares feeds stored in the database against the config and removes any that are no longer defined.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configURLs, err := config.GetFeeds()
			if err != nil {
				return fmt.Errorf("failed to read feeds from config: %w", err)
			}

			inConfig := make(map[string]bool, len(configURLs))
			for _, url := range configURLs {
				inConfig[url] = true
			}

			dbFeeds, err := (*db).GetFeeds()
			if err != nil {
				return fmt.Errorf("failed to get feeds: %w", err)
			}

			var feedsRemoved, articlesRemoved int64

			for _, feed := range dbFeeds {
				if inConfig[feed.URL] {
					continue
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Removing feed: %s (%s)\n", feed.Title, feed.URL)
				feedsRemoved++

				if dryRun {
					articles, err := (*db).GetArticlesByFeedID(feed.ID)
					if err != nil {
						return fmt.Errorf("failed to count articles for feed %q: %w", feed.URL, err)
					}
					articlesRemoved += int64(len(articles))
					continue
				}

				count, err := (*db).DeleteFeedWithArticles(feed.ID)
				if err != nil {
					return fmt.Errorf("failed to delete feed %q: %w", feed.URL, err)
				}
				articlesRemoved += count
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would remove %d feed(s) and %d article(s) (dry run).\n", feedsRemoved, articlesRemoved)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed %d feed(s) and %d article(s).\n", feedsRemoved, articlesRemoved)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be removed without deleting anything")

	return cmd
}
