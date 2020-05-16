package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"
)

var (
	mux  sync.Mutex
	temp float64
)

func metrics(res http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(res, "# Help room_temperature_fahrenheit Current temperature of the room in Fahrenheit\n")
	fmt.Fprintf(res, "# TYPE room_temperature_fahrenheit gauge\n")

	mux.Lock()
	defer mux.Unlock()
	fmt.Fprintf(res, "room_temperature_fahrenheit %9.2f\n", temp)
}

func main() {

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Create a new temperature sensor.
	sensor, err := mcp9808.New(bus, &mcp9808.DefaultOpts)
	if err != nil {
		log.Fatalln(err)
	}

	// Read values from sensor.

	go func() {

		for {
			measurement, err := sensor.SenseTemp()
			if err != nil {
				log.Fatalln(err)
			}

			mux.Lock()
			temp = measurement.Fahrenheit()
			mux.Unlock()
			log.Printf("room_temperature_fahrenheit %9.2f\n", temp)

			time.Sleep(time.Second)
		}
	}()

	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":8080", nil)

}
