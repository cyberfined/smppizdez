package main

import (
	"os"
	"smppizdez/glade"
	"smppizdez/json_storage"
	"smppizdez/smpp"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var mainWindow *gtk.Window

func activate(app *gtk.Application) {
	accRepo, err := json_storage.Open("data.json")
	if err != nil {
		errorDialogFatal("Failed to open database: %v", err)
	}

	cssProvider, _ := gtk.CssProviderNew()
	err = cssProvider.LoadFromData(".invalid-entry{border:1px red solid;}")
	if err != nil {
		errorDialogFatal("Failed to create CSS provider: %v", err)
	}
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

	builder, err := gtk.BuilderNewFromString(glade.Source)
	if err != nil {
		errorDialogFatal("Failed to initialize GTK builder: %v", err)
	}

	mainWindowI, _ := builder.GetObject("main_window")
	mainWindow = mainWindowI.(*gtk.Window)
	initAccountsList(builder, accRepo)

	sender := smpp.Sender{
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		EnquireLink:  30 * time.Second,
	}
	initSubmitSmForm(builder, sender)

	mainWindow.Present()
	app.AddWindow(mainWindow)
}

func main() {
	const appID = "org.smppizdez"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		panic(err)
	}
	application.Connect("activate", func() { activate(application) })
	os.Exit(application.Run(os.Args))
}
