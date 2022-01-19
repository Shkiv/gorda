package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gotk3/gotk3/gtk"
)

type session struct {
	UUID  uuid.UUID
	Start time.Time
	End   time.Time
}

const (
	SPACING              int           = 5
	UPDATE_INTERVAL      time.Duration = 5 * time.Second
	DURATION_LABEL_WIDTH int           = 20
)

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

	ticker := time.NewTicker(UPDATE_INTERVAL)
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
		name, err := item.(*gtk.Widget).GetName()
		handleFatalError(err)
		_, err = getSessionById(sessions, name)
		if err != nil {
			item.(*gtk.Widget).Destroy()
		}
	})

	for i, session := range sessions {
		_, err := getRowByName(listBox, session.UUID.String())
		if err != nil {
			box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, SPACING)
			startLabel, _ := gtk.LabelNew("Start: " + session.Start.Format(clockLayout))
			endLabel, _ := gtk.LabelNew("End: " + session.End.Format(clockLayout))
			separator0, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)

			duration := session.End.Sub(session.Start).Round(time.Second)
			durationLabel, _ := gtk.LabelNew("Duration: " + duration.String())
			durationLabel.SetXAlign(0)
			durationLabel.SetWidthChars(DURATION_LABEL_WIDTH)

			separator1, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
			editButton, _ := gtk.ButtonNewFromIconName("gtk-edit", gtk.ICON_SIZE_BUTTON)

			box.Add(startLabel)
			box.Add(separator0)
			box.Add(endLabel)
			box.Add(separator1)
			box.Add(durationLabel)
			box.Add(editButton)
			listBox.Insert(box, i)
			listBox.GetRowAtIndex(i).SetName(session.UUID.String())
		}
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

func getSessionById(sessions []session, id string) (session, error) {
	var err error
	for _, session := range sessions {
		if session.UUID.String() == id {
			return session, err
		}
	}
	err = errors.New("no session found")
	return session{}, err
}

func getRowByName(listBox *gtk.ListBox, name string) (*gtk.Widget, error) {
	var (
		result *gtk.Widget
		err    error
	)
	listBox.GetChildren().Foreach(func(item interface{}) {
		widgetName, err := item.(*gtk.Widget).GetName()
		handleFatalError(err)
		if widgetName == name {
			result = item.(*gtk.Widget)
			return
		}
	})
	if result == nil {
		err = errors.New("no listbox found")
	}
	return result, err
}

func handleFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
