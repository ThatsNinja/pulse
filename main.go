package main

import (
	"github.com/polydice/pulse"
	"os"
)

func main() {
	pump := pulse.New(":" + os.Getenv("PORT"))

	//cafeActivity := messenger.New("cafe_activity")
	//pump.RegisterMessenger(cafeActivity)

	// Allow cross domain requset.
	pump.Start(true)
}
