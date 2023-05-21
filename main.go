package main

import (
	"timer/app"
)

func main() {
	app.GetMigratorApp().Start()

	app.GetSchedulerApp().Start()

	app.GetWebApp().Start()

	var c chan struct{}
	<-c
}
