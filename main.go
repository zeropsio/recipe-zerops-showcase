package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
	"time"
)

var Commit string

func main() {
	_ = godotenv.Load("replace.env")

	rootCmd := new(cobra.Command)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "serve HTTP",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkCacheDir, err := cmd.Flags().GetBool("check-cache-dir")
			if err != nil {
				return err
			}
			if checkCacheDir {
				var info os.FileInfo
				info, err = os.Stat("./cache")
				if err != nil {
					return err
				}
				if !info.IsDir() {
					return errors.New("cache dir is not a directory")
				}
			}

			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, Commit)
			})

			port := 8080
			if stringPort, ok := os.LookupEnv("APP_PORT"); ok {
				var err error
				port, err = strconv.Atoi(stringPort)
				if err != nil {
					return fmt.Errorf("port must be an integer: %w", err)
				}
			}

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("listening on %s\n", addr)
			err = http.ListenAndServe(addr, http.DefaultServeMux)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("listen and serve: %w", err)
			}

			return nil
		},
	}
	serveCmd.Flags().Bool("check-cache-dir", false, "check the cache directory existence")

	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "print build commit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Commit)
		},
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "simulate migration command by writing a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Create("/var/www/migration.txt")
			if err != nil {
				return err
			}
			defer f.Close()
			_, _ = fmt.Fprintln(f, time.Now().String())
			return nil
		},
	}

	rootCmd.AddCommand(
		serveCmd,
		commitCmd,
		migrateCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
