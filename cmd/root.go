// Package cmd has all the commands for the application.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/benmatselby/wibble/cmd/tui"
	"github.com/benmatselby/wibble/pkg/client"
	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/theme"
	"github.com/benmatselby/wibble/pkg/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// ApplicationName is the name of the cli binary
const ApplicationName = "wibble"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once
func Execute() {
	// db is populated by PersistentPreRunE in NewRootCommand, after flags are parsed.
	var db dao.DaoClient
	var api client.API

	// Build the root command
	cmd := NewRootCommand(&db, &api)

	// Execute the application
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewRootCommand builds the main cli application and
// adds all the child commands
func NewRootCommand(db *dao.DaoClient, api *client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:     ApplicationName,
		Short:   "RSS reader",
		Version: version.GITCOMMIT,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return fmt.Errorf("error loading config: %v", err)
			}

			sqlite, err := dao.NewSQLiteClient()
			if err != nil {
				return err
			}

			rssAPI := client.NewGoFeedClient()

			*db = sqlite
			*api = rssAPI

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if *db != nil {
				return (*db).Close()
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(*db, *api, theme.Load())
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.config/%s/%s.yaml)", ApplicationName, ApplicationName))
	cmd.PersistentFlags().String("database", fmt.Sprintf("$HOME/.config/%s/%s.db", ApplicationName, ApplicationName), "Path to the SQLite database file")
	_ = viper.BindPFlag("database", cmd.PersistentFlags().Lookup("database"))

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(strings.Join([]string{home, fmt.Sprintf(".config/%s", ApplicationName)}, "/"))
		viper.SetConfigName(ApplicationName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
}
