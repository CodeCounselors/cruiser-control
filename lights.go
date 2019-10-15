package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"strings"

	"periph.io/x/periph/host"
)

/**
Omega GPIO Rules

IMPORTANT: You will damage your Omega if your GPIO is set to output and you try to drive external current to the pin.

GPIO7|8|9 - Reserved for SPI
GPIO38	FW_RST	    Exposed: Yes
GPIO44	Omega LED	Exposed: No

- Since the Omega’s storage uses SPI, the SPI communication pins - GPIOs 7, 8, and 9 - must be used for the SPI protocol and cannot be used as regular GPIOs.
 */

/**
The RGB LED on the Expansion Dock consists of three LEDs that give you the ability to create RGB colors on the Expansion Dock.
GPIO 17 controls the red LED.
GPIO 16 controls the green LED.
GPIO 15 controls the blue LED.
The RGB LED is active-low, meaning that setting the GPIO to 0 will turn on the designated LED.
 */

const ROOF_BAR = "15"
const ROOF_RIGHT = "16"
const ROOF_LEFT = "17"

type PinState struct {
	LedState  string
	NextState string
	PinNum    string
	Pin       gpio.PinIO
	Name      string
}

var (
	pins []PinState
	pinByNumber map[string]PinState
)

/**
<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
<link rel="stylesheet" href="https://code.getmdl.io/1.3.0/material.indigo-pink.min.css">
 */
var homeTemplate = `
<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Cruiser Controller</title>
		<style>
			button {
				width: 300px;
				height: 100px;
				margin-top: 5px;
				border-radius: 10px;
				background: lightgoldenrodyellow;
				font-size: 30px;
			}
			.On {
				background-color: gray;
			}
			.Off {
				background-color: green;
			}
		</style>
	</head>
	<body class="container">
		<center>		
			<h3>Cruiser Controller</h3>
			<br/>
			{{range .Pins}}
			<button
				class="mdl-button mdl-js-button mdl-button--raised mdl-js-ripple-effect mdl-button--accent {{ .NextState }}"
				onClick="javascript:location.href = '/switch?pin={{ .PinNum }}';">
				{{ .Name }}
			</button>
			<br/>
			{{end}}
		</center>
	</body>
</html>
`



func main() {
	_, err := host.Init() // Init periph.io
	if err != nil {
		log.Fatal(err)
	}

	setupPins()

	// Close when the program ends
	for i := range pins {
		defer pins[i].Pin.Halt()
	}

	// Route Handlers
	http.HandleFunc("/switch", handleSwitchRequest)
	http.HandleFunc("/", handleHomeRequest)

	fmt.Println("Server running at port 9090...")
	http.ListenAndServe(":9090", nil) // This blocks execution
}

func setupPins() {
	pinByNumber = make(map[string]PinState)
	pinByNumber[ROOF_BAR] = PinState{
		Pin:    gpioreg.ByName(ROOF_BAR),
		PinNum: ROOF_BAR,
		Name:   "Roof Bar",
	}
	pinByNumber[ROOF_RIGHT] = PinState{
		Pin:    gpioreg.ByName(ROOF_RIGHT),
		PinNum: ROOF_RIGHT,
		Name:   "Roof (Right)",
	}
	//pinByNumber[ROOF_LEFT] = PinState{
	//	Pin:  gpioreg.ByName(ROOF_LEFT),
	//	PinNum:    ROOF_LEFT,
	//	Name:      "Roof (Left)",
	//}
	pins = [] PinState{
		pinByNumber[ROOF_BAR],
		pinByNumber[ROOF_RIGHT],
		//pinByNumber[ROOF_LEFT]
	}

	for i := range pins {
		pins[i].NextState = "On"
	}
}

// Endpoint that handle the home page render
func handleHomeRequest(res http.ResponseWriter, req *http.Request) {
	t := template.New("main")
	t, _ = t.Parse(homeTemplate)
	_ = t.Execute(res, struct {
		Pins []PinState
	}{pins})
}

func handleSwitchRequest(res http.ResponseWriter, req *http.Request) {
	pinParam, ok := req.URL.Query()["pin"]

	if !ok || len(pinParam[0]) < 1 {
		log.Println("Url Param 'pin' is missing")
		return
	}

	pinState := pinByNumber[pinParam[0]]
	led := pinState.Pin

	nextState := switchLED(led)
	pinState.NextState = nextState
	log.Println("nextState is", nextState,
		"Toggled state of pin", pinState.Name,
		"number:", pinState.PinNum,
		"nextState:", pinState.NextState)

	if isJsonReq(req) {
		sendJsonState(res, "ok", "On")
	} else {
		http.Redirect(res, req, "/", 303)
	}
}

func sendJsonState(res http.ResponseWriter, message, state string) {
	body := struct {
		Message string `json:"message"`
		State   string `json:"state"`
	}{
		Message: message,
		State:   state,
	}
	json.NewEncoder(res).Encode(body)
}

func isJsonReq(req *http.Request) bool {
	return strings.Contains(req.Header.Get("content-type"), ("json"))
}

func switchLED(led gpio.PinIO) string {
	pinReading := led.Read()
	nextState := "On"
	if (pinReading == gpio.Low) {
		led.Out(gpio.High)
		nextState = "Off"
	} else {
		led.Out(gpio.Low)
	}
	return nextState
}


