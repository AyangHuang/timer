package main

import (
	"timer/app"
)

func main() {
	webApp := app.GerWebApp()
	webApp.Start()
	var c chan struct{}
	<-c
}
