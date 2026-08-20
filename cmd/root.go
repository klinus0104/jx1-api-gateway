package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/klinus0104/jx1-api-gateway/cmd/server"
	"github.com/klinus0104/jx1-api-gateway/internal/config"
	dbpool "github.com/klinus0104/jx1-api-gateway/pkg/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initConfig() {
	config.LoadEnv()
	if err := runMigrations(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func runMigrations() error {
	database, err := dbpool.Open(context.Background(), dbpool.Config{URL: config.Cfg.DatabaseURL, MaxOpenConns: 2, MaxIdleConns: 1, ConnMaxLifetime: 5 * time.Minute})
	if err != nil {
		return err
	}
	defer database.Close()
	if err = dbpool.Migrate(context.Background(), database); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(server.Cmd)
}

var rootCmd = &cobra.Command{
	Use:   "jx-api-gateway",
	Short: "CMD working with jx-api-gateway",
	Long:  "CMD working with jx-api-gateway",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Execute command error: %+v", err)
	}
}
