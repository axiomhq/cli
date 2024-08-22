package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var messages = []string{
	"Log generation job timed out. Please check the configuration and try again.",
	"Insufficient resources to continue log generation. Check system resources and configuration.",
	"Invalid configuration detected. Verify the log generator settings and update accordingly.",
	"Unable to write logs to the specified file. Check file permissions and disk space availability.",
	"Required dependencies not found. Install the necessary libraries and try again.",
	"Failed to establish a connection with the log storage server. Check network settings and connectivity.",
	"Log data format is not valid. Ensure that the log entries adhere to the specified format.",
	"Security violation detected during log generation. Review security policies and configurations.",
	"Memory allocation failure. Check system memory and adjust log generator settings if necessary.",
	"An unexpected exception occurred during log generation. Review logs for more details or contact support.",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
	)
	defer stop()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	attrs := slog.Group("system",
		slog.String("hostname", hostname),
		slog.Int("pid", os.Getpid()),
	)

	logger.InfoContext(ctx, "Starting log generator", attrs)

	for {
		r := rand.IntN(1000) + 1 //nolint:gosec // not used for security purposes
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Duration(r) * time.Millisecond):
		}

		level := slog.Level(r % 3 * 4)
		msg := messages[r%len(messages)]

		logger.Log(ctx, level, msg, attrs)
	}
}
