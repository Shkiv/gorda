package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

type session struct {
	Start time.Time
	End   time.Time
}

var clockLayout string = "15:04"

func main() {
	gtk.Init(nil)

	builder, err := gtk.BuilderNew()
	handleFatalError(err)

	err = builder.AddFromFile("gorda.glade")
	handleFatalError(err)

	windowObject, err := builder.GetObject("main_window")
	handleFatalError(err)

	window := windowObject.(*gtk.Window)
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	startButtonObject, err := builder.GetObject("start_button")
	handleFatalError(err)

	stopButtonObject, err := builder.GetObject("stop_button")
	handleFatalError(err)

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
	handleFatalError(err)

	sessionLabel := labelObject.(*gtk.Label)
	durationLabelObject, err := builder.GetObject("duration_label")
	handleFatalError(err)

	durationLabel := durationLabelObject.(*gtk.Label)
	if string(body) != "null" {
		newSessionLabel := "Started: " + message.Format(clockLayout)
		sessionLabel.SetLabel(newSessionLabel)

		duration := time.Since(message).Round(time.Minute)
		newDurationLabel := "Duration: " + duration.String()
		durationLabel.SetLabel(newDurationLabel)
	} else {
		sessionLabel.SetLabel("No active interval")
		durationLabel.SetLabel("")
	}
}

func getSessions() ([]session, error) {
	resp, err := http.Get("http://localhost:8090/intervals")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sessions []session
	err = json.Unmarshal(body, &sessions)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return sessions, nil
}

func updateSessions(builder *gtk.Builder) {
	sessions, err := getSessions()
	if err != nil {
		log.Print(err)
	}

	listBoxObject, err := builder.GetObject("session_list")
	handleFatalError(err)

	listBox := listBoxObject.(*gtk.ListBox)
	listBox.GetChildren().Foreach(func(item interface{}) {
		item.(*gtk.Widget).Destroy()
	})

	for i, session := range sessions {
		box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
		handleFatalError(err)

		startLabel, err := gtk.LabelNew("Start: " + session.Start.Format(clockLayout))
		handleFatalError(err)

		endLabel, err := gtk.LabelNew("End: " + session.End.Format(clockLayout))
		handleFatalError(err)

		separator, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
		handleFatalError(err)

		box.Add(startLabel)
		box.Add(separator)
		box.Add(endLabel)
		listBox.Insert(box, i)
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

func handleFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
