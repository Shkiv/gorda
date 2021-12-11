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

	go updateActiveSession(builder)
	go updateSessions(builder)

	startButton.Connect("clicked", func() {
		startSession(builder)
	})
	stopButton.Connect("clicked", func() {
		stopSession(builder)
	})

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-ticker.C
			updateActiveSession(builder)
			updateSessions(builder)
		}
	}()

	window.ShowAll()
	gtk.Main()
}

func updateActiveSession(builder *gtk.Builder) {
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

	labelObject, err := builder.GetObject("session_label")
	if err != nil {
		log.Fatal("Error:", err)
	}
	sessionLabel := labelObject.(*gtk.Label)

	durationLabelObject, err := builder.GetObject("duration_label")
	if err != nil {
		log.Fatal("Error:", err)
	}
	durationLabel := durationLabelObject.(*gtk.Label)

	if string(body) != "null" {
		hour, minute, _ := message.Clock()
		minuteString := strconv.Itoa(minute)
		if minute < 10 {
			minuteString = "0" + minuteString
		}
		newSessionLabel := "Started: " + strconv.Itoa(hour) + ":" + minuteString
		sessionLabel.SetLabel(newSessionLabel)

		duration := time.Since(message).Round(time.Minute)
		newDurationLabel := "Duration: " + duration.String()
		durationLabel.SetLabel(newDurationLabel)
	} else {
		sessionLabel.SetLabel("No active interval")
		durationLabel.SetLabel("")
	}
}

func updateSessions(builder *gtk.Builder) {
	type session struct {
		Start time.Time
		End   time.Time
	}

	resp, err := http.Get("http://localhost:8090/intervals")
	if err != nil {
		log.Print(err)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	var sessions []session
	err = json.Unmarshal(body, &sessions)
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()

	listBoxObject, err := builder.GetObject("session_list")
	if err != nil {
		log.Fatal(err)
	}
	listBox := listBoxObject.(*gtk.ListBox)
	listBox.GetChildren().Foreach(func(item interface{}) {
		item.(*gtk.Widget).Destroy()
	})

	for i, session := range sessions {
		label, err := gtk.LabelNew(session.Start.String() + " " + session.End.String())
		if err != nil {
			log.Fatal(err)
		}
		listBox.Insert(label, i)
	}
	listBox.ShowAll()
}

func startSession(builder *gtk.Builder) {
	_, err := http.Get("http://localhost:8090/start")
	if err != nil {
		log.Print(err)
	}
	go updateActiveSession(builder)
	go updateSessions(builder)
}

func stopSession(builder *gtk.Builder) {
	_, err := http.Get("http://localhost:8090/stop")
	if err != nil {
		log.Print(err)
	}
	go updateActiveSession(builder)
	go updateSessions(builder)
}
