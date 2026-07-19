package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	couchdev "github.com/pecodez/couchdev"
	"github.com/pecodez/couchdev/internal/api"
	"github.com/pecodez/couchdev/internal/auth"
	"github.com/pecodez/couchdev/internal/config"
	"github.com/pecodez/couchdev/internal/db"
	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/tmux"
)

var version = "dev"

func main() {
	root := &cobra.Command{Use: "couchdev", Short: "Claude Code RC session launcher"}

	var configPath string
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "/etc/couchdev/config.json", "config file")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			var (
				log *zap.Logger
				err error
			)
			if verbose {
				log, err = zap.NewDevelopment()
			} else {
				log, err = zap.NewProduction()
			}
			if err != nil {
				return fmt.Errorf("init logger: %w", err)
			}
			defer log.Sync()

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			var tokenHash []byte
			if cfg.RequireAuth {
				tokenHash, err = auth.ParseTokenHash(cfg.TokenHash)
				if err != nil {
					return err
				}
			} else {
				log.Warn("require_auth is disabled — /api/ routes are unauthenticated")
			}
			conn, err := db.Open(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer conn.Close()

			runDiscovery(cfg.ProjectsDir, project.NewStore(conn), log)

			handler := api.New(tokenHash, conn, tmux.NewExec(log), couchdev.WebFS, cfg.ProjectsDir, git.Real{}, log)
			log.Info("starting", zap.String("addr", cfg.ListenAddr))
			if cfg.TLSCert != "" && cfg.TLSKey != "" {
				return http.ListenAndServeTLS(cfg.ListenAddr, cfg.TLSCert, cfg.TLSKey, handler)
			}
			log.Warn("TLS not configured — serving plain HTTP (development only)")
			return http.ListenAndServe(cfg.ListenAddr, handler)
		},
	}
	serveCmd.Flags().BoolP("verbose", "v", false, "enable verbose debug logging")

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
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println(version) },
	}

	tokenCmd.AddCommand(tokenGenCmd)
	root.AddCommand(serveCmd, tokenCmd, versionCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// runDiscovery scans projectsDir for flat git repos (not yet in src/ layout),
// prompts for confirmation when running interactively, migrates the filesystem,
// and registers any unregistered repos in the DB.
func runDiscovery(projectsDir string, ps *project.Store, log *zap.Logger) {
	flat, err := project.ScanFlat(projectsDir)
	if err != nil {
		log.Warn("discovery scan failed", zap.Error(err))
		return
	}
	if len(flat) == 0 {
		return
	}

	if !stdinIsTerminal() {
		log.Warn("flat-layout repos found; start interactively to migrate",
			zap.Int("count", len(flat)), zap.String("projects_dir", projectsDir))
		return
	}

	fmt.Printf("\nFound %d repo(s) in %s without src/ layout:\n", len(flat), projectsDir)
	for _, r := range flat {
		fmt.Printf("  %s\n", r.Name)
	}
	fmt.Print("Move each into a src/ subdirectory? [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
		log.Info("discovery migration skipped by user")
		return
	}

	for _, r := range flat {
		if err := project.MigrateFlat(projectsDir, r.Name); err != nil {
			log.Error("migration failed", zap.String("name", r.Name), zap.Error(err))
			continue
		}
		newPath := filepath.Join(projectsDir, r.Name, "src")
		if existing, _ := ps.GetByName(r.Name); existing != nil {
			if err := ps.UpdatePath(r.Name, newPath); err != nil {
				log.Error("update path failed", zap.String("name", r.Name), zap.Error(err))
			}
		} else {
			if _, err := ps.Create(project.Project{
				Name:       r.Name,
				RepoPath:   newPath,
				SourceType: "existing",
			}); err != nil {
				log.Error("register failed", zap.String("name", r.Name), zap.Error(err))
			}
		}
		log.Info("migrated repo", zap.String("name", r.Name), zap.String("src", newPath))
	}
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}
