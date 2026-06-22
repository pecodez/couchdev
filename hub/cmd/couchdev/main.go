package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	couchdev "github.com/pecodez/couchdev"
	"github.com/pecodez/couchdev/internal/api"
	"github.com/pecodez/couchdev/internal/auth"
	"github.com/pecodez/couchdev/internal/config"
	"github.com/pecodez/couchdev/internal/db"
	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/tmux"
)

func main() {
	root := &cobra.Command{Use: "couchdev", Short: "Claude Code RC session launcher"}

	var configPath string
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "/etc/couchdev/config.json", "config file")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			log, _ := zap.NewProduction()
			defer log.Sync()

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			tokenHash, err := auth.ParseTokenHash(cfg.TokenHash)
			if err != nil {
				return err
			}
			conn, err := db.Open(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer conn.Close()

			handler := api.New(tokenHash, conn, tmux.Exec{}, couchdev.WebFS, cfg.ProjectsDir, git.Real{})
			log.Info("starting", zap.String("addr", cfg.ListenAddr))
			if cfg.TLSCert != "" && cfg.TLSKey != "" {
				return http.ListenAndServeTLS(cfg.ListenAddr, cfg.TLSCert, cfg.TLSKey, handler)
			}
			log.Warn("TLS not configured — serving plain HTTP (development only)")
			return http.ListenAndServe(cfg.ListenAddr, handler)
		},
	}

	tokenCmd := &cobra.Command{Use: "token", Short: "Token management"}
	tokenGenCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a bearer token and print its SHA-256 hash for config",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw := make([]byte, 32)
			if _, err := rand.Read(raw); err != nil {
				return err
			}
			token := hex.EncodeToString(raw)
			sum := sha256.Sum256([]byte(token))
			fmt.Printf("Token (copy to phone app):\n%s\n\ntoken_hash (put in config.json):\n%s\n",
				token, hex.EncodeToString(sum[:]))
			return nil
		},
	}
	tokenCmd.AddCommand(tokenGenCmd)
	root.AddCommand(serveCmd, tokenCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
