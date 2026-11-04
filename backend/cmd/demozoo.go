package cmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/demozoo"
)

func NewDemozooCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demozoo",
		Short: "Demozoo data commands",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}
	cmd.AddCommand(NewDemozooExportCmd())

	return cmd
}

func NewDemozooExportCmd() *cobra.Command {
	var (
		input, output, url string
		force              bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Demozoo sceners to CSV",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && input == "" {
				input = args[0]
			}

			return runDemozooExport(cmd.Context(), input, output, url, force)
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "Path to a local demozoo SQL dump (.sql or .gz)")

	outputDefault := filepath.Join(config.DataDirname, "import", "demozoo-sceners.csv")
	cmd.Flags().StringVar(&output, "output", outputDefault, "Output CSV path")
	cmd.Flags().StringVar(&url, "url", "https://data.demozoo.org/demozoo-export.sql.gz", "Demozoo dump URL")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output file")

	return cmd
}

func runDemozooExport(ctx context.Context, input, output, url string, force bool) error {
	if input == "" {
		input = filepath.Join(config.DataDirname, "import", "demozoo-export.sql")
	}

	input, cleanup, err := resolveDemozooInput(ctx, input, url)
	if err != nil {
		return err
	}

	if cleanup != nil {
		defer cleanup()
	}

	reader, closeReader, err := openDemozooReader(input)
	if err != nil {
		return err
	}
	defer closeReader()

	sceners, err := demozoo.ParseDump(reader)
	if err != nil {
		return fmt.Errorf("failed to parse demozoo dump: %w", err)
	}

	if err := writeDemozooCSV(output, force, sceners); err != nil {
		return err
	}

	slog.Info("Demozoo export completed",
		"sceners", len(sceners),
		"output", output,
	)

	return nil
}

func resolveDemozooInput(ctx context.Context, input, url string) (path string, cleanup func(), err error) {
	if _, err := os.Stat(input); err == nil {
		return input, nil, nil
	}

	tmp, err := os.CreateTemp("", "demozoo-export-*.sql")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmp.Name()
	tmp.Close()

	if err := downloadDemozooDump(ctx, url, tmpPath); err != nil {
		_ = os.Remove(tmpPath)

		return "", nil, fmt.Errorf("failed to download demozoo dump: %w", err)
	}

	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}

func openDemozooReader(input string) (io.Reader, func(), error) {
	f, err := os.Open(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open input file %s: %w", input, err)
	}

	closeFn := func() { _ = f.Close() }

	if !strings.HasSuffix(strings.ToLower(input), ".gz") {
		return f, closeFn, nil
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		closeFn()

		return nil, nil, fmt.Errorf("failed to open gzip input %s: %w", input, err)
	}

	return gz, func() {
		_ = gz.Close()

		closeFn()
	}, nil
}

func writeDemozooCSV(output string, force bool, sceners []demozoo.Scener) error {
	if !force {
		if _, err := os.Stat(output); err == nil {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", output)
		}
	}

	if dir := filepath.Dir(output); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, config.DataDirPerm); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", dir, err)
		}
	}

	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", output, err)
	}
	defer out.Close()

	if err := demozoo.WriteCSV(out, sceners); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

func downloadDemozooDump(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}

	const downloadTimeoutMinutes = 10

	client := &http.Client{Timeout: downloadTimeoutMinutes * time.Minute}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}
