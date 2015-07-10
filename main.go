package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
		var previousStockPrice float64

		gobot.Every(3*time.Second, func() {
			// retrieve: opening price, last price
			resp, err := http.Get("http://download.finance.yahoo.com/d/quotes.csv?s=GOOG&f=ol1")
			defer resp.Body.Close()
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				quoteParts := strings.Split(string(body), ",")
				if len(quoteParts) > 1 {
					openingPrice, _ := strconv.ParseFloat(strings.TrimSpace(quoteParts[0]), 32)
					lastPrice, _ := strconv.ParseFloat(strings.TrimSpace(quoteParts[1]), 32)

					// set color based on status of stock
					if lastPrice > openingPrice {
						driver.SetRGB(0, 255, 0) //green
					} else {
						driver.SetRGB(255, 0, 0) //red
					}
					fmt.Printf("Last Price=%f, Opening Price=%f\n", lastPrice, openingPrice)

					// calculate percentage change for the day
					if previousStockPrice == 0.0 {
						previousStockPrice = lastPrice
					}
					percentageChange := uint16(10000 * ((lastPrice - previousStockPrice) / previousStockPrice))
					fmt.Printf("Percentage change: %d\n", percentageChange)

					var heading uint16
					if percentageChange < 0 {
						heading = 360 + percentageChange
					} else {
						heading = percentageChange
					}

					fmt.Printf("Heading: %d\n", heading)
					driver.Roll(100, heading)
					time.Sleep(1 * time.Second)
					driver.Stop()

					previousStockPrice = lastPrice
				}
			}
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
	fmt.Println(<-ch)

	// Stop the service gracefully.
	fmt.Println("Shutdown successful, exiting")
}
