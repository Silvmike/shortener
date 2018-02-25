package main

import (
	"github.com/atotto/clipboard"
	"io/ioutil"
	"bytes"
	"net/http"
	"net/url"
	"github.com/gvalkov/golang-evdev"
	"fmt"
	"log"
	"strings"
	"time"
	"github.com/MarinX/keylogger"
)

type ResponseParsingError struct {
	When     time.Time
	What     string
	Response *[]byte
}

func (e ResponseParsingError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func parseResponse(response *[]byte) (string, error) {
	short := string(*response)
	idIndex := strings.Index(short, "id")
	if idIndex != -1 {
		short = short[(idIndex + 3):]
		short = short[(strings.Index(short, "\"") + 1):]
		short = short[:strings.Index(short, "\"")]
		return short, nil
	} else {
		return "", ResponseParsingError{time.Now(), "Bad response", response}
	}
}

func formatRequest(address string) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("{\"longUrl\":\"")
	buffer.WriteString(address)
	buffer.WriteString("\"}")
	return buffer.Bytes()
}

func obtainShortUrl(address string, callback func(long string, short string)) {

	// validating URL
	_, err := url.ParseRequestURI(address)
	if err != nil {
		return
	}

	const shortenerURL = "https://www.googleapis.com/urlshortener/v1/url?key=YOUR_API_KEY_HERE"

	res, err := http.Post(shortenerURL, "application/json", bytes.NewReader(formatRequest(address)))
	if err != nil {
		log.Fatal(err)
		return
	}

	defer res.Body.Close()
	shortBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	if res.StatusCode == 200 {

		short, err := parseResponse(&shortBytes)
		if err == nil {
			callback(address, short)
		} else {
			log.Fatal(err)
		}

	} else {

		panic(fmt.Sprintf("Service returned with a bad status code[%d] and body:\n%s\n", res.StatusCode, string(shortBytes)))

	}

}

func processKeyboardEvents(inputChannel chan string) {
	for {
		key := <-inputChannel
		if key == "KEY_F10" {
			data, err := clipboard.ReadAll()
			if err == nil {
				go obtainShortUrl(data, func(long string, short string) {
					log.Printf("Converted [ %s ] to [ %s ]\n", long, short)
					clipboard.WriteAll(short)
				})
			}
		}
	}
}

/*
 Returns slice of pointers to evdev.InputDevice
	which is identified as a keyboard device based on assumption
	that every keyboard have an escape button and led indicators
 */
func getKeyboards() []*evdev.InputDevice {

	var result []*evdev.InputDevice

	devices, _ := evdev.ListInputDevices()

	for _, dev := range devices {
		if isKeyboard(dev) {
			result = append(result, dev)
		}
	}

	return result

}

func isKeyboard(device *evdev.InputDevice) bool {
	hasLeds := false
	hasEscButton := false
	for key, value := range device.Capabilities {
		for _, capabilityCode := range value {
			if capabilityCode.Name == "KEY_ESC" {
				hasEscButton = true
			}
		}
		if key.Name == "EV_LED" {
			hasLeds = true
		}
		if hasEscButton && hasLeds {
			break
		}
	}
	return hasLeds && hasEscButton
}

func main() {

	inputChannel := make(chan string)
	go processKeyboardEvents(inputChannel)

	keyboards := getKeyboards()
	if keyboards != nil {
		keyboard := keyboards[0]
		log.Printf("Detected keyboard event bus: %s\n", keyboard.Fn)
		for {
			inputEvent, err := keyboard.ReadOne()
			if err == nil {
				if inputEvent.Type == keylogger.EV_KEY {
					keyEvent := evdev.NewKeyEvent(inputEvent)
					if keyEvent.State == evdev.KeyDown {
						inputChannel <- evdev.KEY[int(keyEvent.Scancode)]
					}
				}
			} else {
				break
			}
		}
	} else {
		log.Fatal("There was no keyboard found. Are you using root?")
	}

}