package main

import (
	"fmt"
	"log"
	"sync"
	"net/http"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/bmxx80"
	"periph.io/x/periph/host"
)

func main() {

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	// Open a handle to a bme280/bmp280 connected on the I²C bus using default
	// settings:
	dev, err := bmxx80.NewI2C(bus, 0x76, &bmxx80.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Halt()

	var mutex sync.Mutex
	metrics := func(res http.ResponseWriter, req *http.Request) {

		mutex.Lock()
		defer mutex.Unlock()

		// Read temperature from the sensor:
		var env physic.Env
		if err = dev.Sense(&env); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%9.2fF %10s %9s\n", env.Temperature.Fahrenheit(), env.Pressure, env.Humidity)

		fmt.Fprintf(res, "# Help room_temperature_fahrenheit Current temperature of the room in Fahrenheit\n")
		fmt.Fprintf(res, "# TYPE room_temperature_fahrenheit gauge\n")
		fmt.Fprintf(res, "room_temperature_fahrenheit %9.2f\n", env.Temperature.Fahrenheit())

		fmt.Fprintf(res, "# Help room_pressure_kilopascal Current pressure of the room in KiloPascal\n")
		fmt.Fprintf(res, "# TYPE room_pressure_kilopascal gauge\n")
		fmt.Fprintf(res, "room_pressure_kilopascal %9.2f\n", float64(env.Pressure) / float64(physic.KiloPascal))

		fmt.Fprintf(res, "# Help room_humidity_percent Current relative humidity of the room\n")
		fmt.Fprintf(res, "# TYPE room_humidity_percent gauge\n")
		fmt.Fprintf(res, "room_humidity_percent %9.2f\n", float64(env.Humidity)/float64(physic.PercentRH))
	}

	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":8080", nil)

}
