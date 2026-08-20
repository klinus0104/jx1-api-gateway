package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/klinus0104/jx1-api-gateway/internal/config"
	"github.com/klinus0104/jx1-api-gateway/internal/http_handler"
	"github.com/klinus0104/jx1-api-gateway/internal/repository"
	"github.com/klinus0104/jx1-api-gateway/internal/service"
	dbpool "github.com/klinus0104/jx1-api-gateway/pkg/db"
	"github.com/klinus0104/jx1-api-gateway/pkg/heaven"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{Use: "web-service", Short: "manage the API gateway web service"}
var start = &cobra.Command{Use: "start", Short: "start the API gateway", RunE: func(cmd *cobra.Command, args []string) error { return Start(config.Cfg) }}

func init() { Cmd.AddCommand(start) }

func Start(cfg config.Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("MSSQL_URL is required")
	}
	db, err := dbpool.Open(context.Background(), dbpool.Config{URL: cfg.DatabaseURL, MaxOpenConns: 10, MaxIdleConns: 5, ConnMaxLifetime: 30 * time.Minute})
	if err != nil {
		return err
	}
	defer db.Close()
	accountRepo := repository.NewAccountRepository(db)
	gmRepo := repository.NewGMUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	var heavenClient *heaven.Client
	if os.Getenv("GM_API_HEAVEN_ACTIONS") == "1" {
		heavenClient, err = heaven.Dial(context.Background(), cfg.RelayTarget, cfg.HeavenTablePath)
		if err != nil {
			return fmt.Errorf("Heaven connection failed: %w", err)
		}
		defer heavenClient.Close()
		if _, err = heavenClient.Verify(context.Background(), cfg.HeavenServerName, cfg.HeavenServerPassword, cfg.HeavenIdentity); err != nil {
			return fmt.Errorf("Heaven verification failed: %w", err)
		}
	}
	mw := http_handler.NewMiddleware(cfg.JWTSecret)
	adminSvc := service.NewAdminService(accountRepo, gmRepo, auditRepo, heavenClient, cfg.JWTSecret, cfg.TokenTTL)
	playerSvc := service.NewPlayerService(accountRepo, cfg.JWTSecret, cfg.TokenTTL)
	authSvc := service.NewAuthService(mw)
	srv := http_handler.NewServer(normalizeAddr(cfg.Addr), mw, adminSvc, playerSvc, authSvc)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() { <-ctx.Done(); _ = srv.Shutdown() }()
	return srv.Listen()
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ":8080"
	}
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	return ":" + addr
}
