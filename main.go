package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/sphero"
)

func main() {
	deviceName := flag.String("device", "", "path to Sphero device")
	flag.Parse()

	gbot := gobot.NewGobot()

	adaptor := sphero.NewSpheroAdaptor("sphero", *deviceName)
	driver := sphero.NewSpheroDriver(adaptor, "sphero")

	work := func() {
		gobot.Every(3*time.Second, func() {
			driver.Roll(30, uint16(gobot.Rand(360)))
		})
	}

	robot := gobot.NewRobot("sphero",
		[]gobot.Connection{adaptor},
		[]gobot.Device{driver},
		work,
	)

	gbot.AddRobot(robot)

	var errors []error
	go func() {
		errors = gbot.Start()
	}()
	if errors != nil {
		gbot.Robot("sphero").Connection("sphero").Finalize()
	}

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Println(<-ch)

	// Stop the service gracefully.
	log.Println("Shutdown successful, exiting")
}
