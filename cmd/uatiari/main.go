package main

import (
	"context"
	"os"

	"github.com/fernandoguedes/uatiari/internal/app"
)

func main() {
	os.Exit(app.App{}.Run(context.Background(), os.Args[1:]))
}
