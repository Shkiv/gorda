package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Error:", err)
	}

	err = builder.AddFromFile("gorda.glade")
	if err != nil {
		log.Fatal("Error:", err)
	}

	windowObject, err := builder.GetObject("main_window")
	if err != nil {
		log.Fatal("Error:", err)
	}

	window := windowObject.(*gtk.Window)
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	startButtonObject, err := builder.GetObject("start_button")
	if err != nil {
		log.Fatal("Error:", err)
	}
	stopButtonObject, err := builder.GetObject("stop_button")
	if err != nil {
		log.Fatal("Error:", err)
	}
	startButton := startButtonObject.(*gtk.Button)
	stopButton := stopButtonObject.(*gtk.Button)

	labelObject, err := builder.GetObject("label")
	if err != nil {
		log.Fatal("Error:", err)
	}

	label := labelObject.(*gtk.Label)
	go updateActiveSession(label)
	startButton.Connect("clicked", func() {
		startSession()
	})
	stopButton.Connect("clicked", func() {
		stopSession()
	})

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-ticker.C
			updateActiveSession(label)
		}
	}()

	window.ShowAll()
	gtk.Main()
}

func updateActiveSession(label *gtk.Label) {
	resp, err := http.Get("http://localhost:8090/active-interval")
	if err != nil {
		log.Print(err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	var message time.Time
	err = json.Unmarshal(body, &message)
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()

	if string(body) != "null" {
		hour, minute, _ := message.Clock()
		minuteString := strconv.Itoa(minute)
		if minute < 10 {
			minuteString = "0" + minuteString
		}
		newLabel := "Started: " + strconv.Itoa(hour) + ":" + minuteString
		label.SetLabel(newLabel)
	} else {
		label.SetLabel("No active interval")
	}
}

func startSession() {
	_, err := http.Get("http://localhost:8090/start")
	if err != nil {
		log.Print(err)
	}
}

func stopSession() {
	_, err := http.Get("http://localhost:8090/stop")
	if err != nil {
		log.Print(err)
	}
}
