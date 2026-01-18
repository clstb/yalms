package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/clstb/yalms/internal/server"
	"github.com/clstb/yalms/pkg/logseq"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	app := &cli.App{
		Name:  "yalms",
		Usage: "Logseq MCP Server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "logseq-url",
				Value:   "http://127.0.0.1:12315",
				Usage:   "Logseq API URL",
				EnvVars: []string{"LOGSEQ_URL"},
			},
			&cli.StringFlag{
				Name:    "logseq-token",
				Value:   "auth",
				Usage:   "Logseq API Token",
				EnvVars: []string{"LOGSEQ_TOKEN"},
			},
			&cli.StringFlag{
				Name:    "logseq-mode",
				Value:   "general",
				Usage:   "Logseq Mode (general or ontological)",
				EnvVars: []string{"LOGSEQ_MODE"},
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug logging",
			},
		},
		Action: func(c *cli.Context) error {
			// Configure Logging
			var logger *zap.Logger
			var err error
			cfg := zap.NewProductionConfig()
			if c.Bool("debug") {
				cfg = zap.NewDevelopmentConfig()
			}
			cfg.ErrorOutputPaths = []string{"stderr"}
			cfg.OutputPaths = []string{"stderr"}
			logger, err = cfg.Build()
			if err != nil {
				return err
			}
			defer logger.Sync()

			apiURL := c.String("logseq-url")
			token := c.String("logseq-token")
			mode := server.LogseqMode(c.String("logseq-mode"))

			logger.Info("Starting yalms", zap.String("url", apiURL), zap.String("mode", string(mode)))

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			client := logseq.NewClient(apiURL, token, logger)
			mcpServer := server.NewMCPServer(client, logger, mode)

			errChan := make(chan error, 1)
			go func() {
				errChan <- mcpServer.Serve()
			}()

			select {
			case err := <-errChan:
				if err != nil {
					logger.Error("Server error", zap.Error(err))
					return err
				}
			case <-ctx.Done():
				logger.Info("Shutting down")
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
