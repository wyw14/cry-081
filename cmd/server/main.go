package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/wyw14/cry-081/internal/bootstrap"
)

func main() {
	lifetime, release := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer release()
	os.Exit(bootstrap.Execute(lifetime))
}
