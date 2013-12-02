package main

import (
	"github.com/polydice/pulse"
	"os"
)

func main() {
	pump := pulse.New(":" + os.Getenv("PORT"))

	// Allow cross domain requset.
	pump.Start(true)
}
