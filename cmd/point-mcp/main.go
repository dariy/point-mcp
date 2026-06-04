package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dariy/point-mcp/internal/config"
	"github.com/dariy/point-mcp/internal/middleware"
	"github.com/dariy/point-mcp/internal/point"
	"github.com/dariy/point-mcp/internal/prompts"
	"github.com/dariy/point-mcp/internal/resources"
	"github.com/dariy/point-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	var httpAddr string
	flag.StringVar(&httpAddr, "http", "", "listen address for streamable-HTTP transport (e.g. :9000); omit for stdio")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// --http flag takes priority; fall back to config when transport is http.
	if httpAddr == "" && cfg.Transport == "http" {
		httpAddr = cfg.HTTPAddr
	}

	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	client := point.New(cfg.BaseURL, cfg.APIKey, nil)

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "point-mcp",
		Version: "0.0.1",
	}, nil)
	tools.Register(srv, client)
	resources.Register(srv, client)
	prompts.Register(srv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Handle direct tool call if arguments are provided.
	// Syntax: point-mcp <tool_name> [json_args]
	if flag.NArg() > 0 {
		toolName := flag.Arg(0)
		var toolArgs json.RawMessage
		if flag.NArg() > 1 {
			toolArgs = json.RawMessage(flag.Arg(1))
		} else {
			toolArgs = json.RawMessage("{}")
		}

		// Disable logging to stderr to keep output clean, unless a log file is used.
		if cfg.LogFile == "" {
			log.SetOutput(io.Discard)
		}

		res, err := tools.Dispatch(ctx, toolName, toolArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		out, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(out))
		return
	}

	log.Printf("point-mcp starting: base_url=%s transport=%s api_key_set=%v",
		cfg.BaseURL, cfg.Transport, cfg.APIKey != "")
	log.Printf("server initialized: tools, resources, and prompts registered")

	if httpAddr != "" {
		if len(cfg.AuthTokens) == 0 {
			log.Fatal("MCP_AUTH_TOKENS must be set when running in HTTP mode")
		}
		mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
		handler := middleware.BearerAuth(cfg.AuthTokens, mcpHandler)

		// Add a /health endpoint that doesn't require authentication
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		mux.Handle("/", handler)

		httpSrv := &http.Server{Addr: httpAddr, Handler: mux}
		go func() {
			<-ctx.Done()
			_ = httpSrv.Close()
		}()
		log.Printf("point-mcp listening on %s (streamable-HTTP)", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return
	}

	log.Printf("point-mcp ready (stdio transport) — awaiting MCP input on stdin")
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
