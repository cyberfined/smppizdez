package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/gtk"
)

func errorDialog(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	dialog, err := gtk.DialogNewWithButtons(
		"Error",
		mainWindow,
		gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL,
		[]any{"Ok", gtk.RESPONSE_ACCEPT},
	)
	if err == nil {
		var box *gtk.Box
		box, err = dialog.GetContentArea()
		if err == nil {
			var label *gtk.Label
			label, err = gtk.LabelNew(text)
			if err == nil {
				box.Add(label)
			}
		}
	}
	if err != nil {
		log.Fatalf(
			"Failed to create error dialog. Original error: %v, "+
				"dialog creation error: %v",
			text,
			err,
		)
	}
	dialog.Connect("response", func() { dialog.Destroy() })
	dialog.ShowAll()
}

func errorDialogFatal(format string, a ...any) {
	errorDialog(format, a...)
	os.Exit(1)
}

func getEntryById(b *gtk.Builder, id string) *gtk.Entry {
	v, _ := b.GetObject(id)
	return v.(*gtk.Entry)
}

func getComboById(b *gtk.Builder, id string) *gtk.ComboBox {
	v, _ := b.GetObject(id)
	return v.(*gtk.ComboBox)
}

func getButtonById(b *gtk.Builder, id string) *gtk.Button {
	v, _ := b.GetObject(id)
	return v.(*gtk.Button)
}

func getSwitchById(b *gtk.Builder, id string) *gtk.Switch {
	v, _ := b.GetObject(id)
	return v.(*gtk.Switch)
}

func getLabelById(b *gtk.Builder, id string) *gtk.Label {
	v, _ := b.GetObject(id)
	return v.(*gtk.Label)
}

func getCheckById(b *gtk.Builder, id string) *gtk.CheckButton {
	v, _ := b.GetObject(id)
	return v.(*gtk.CheckButton)
}

func getRadioById(b *gtk.Builder, id string) *gtk.RadioButton {
	v, _ := b.GetObject(id)
	return v.(*gtk.RadioButton)
}

func getMenuItemById(b *gtk.Builder, id string) *gtk.MenuItem {
	v, _ := b.GetObject(id)
	return v.(*gtk.MenuItem)
}

func getTextViewById(b *gtk.Builder, id string) *gtk.TextView {
	v, _ := b.GetObject(id)
	return v.(*gtk.TextView)
}

func getTreeViewById(b *gtk.Builder, id string) *gtk.TreeView {
	v, _ := b.GetObject(id)
	return v.(*gtk.TreeView)
}

func getMenuById(b *gtk.Builder, id string) *gtk.Menu {
	v, _ := b.GetObject(id)
	return v.(*gtk.Menu)
}

func markInvalidEntry(e *gtk.Widget, tip string) {
	e.SetTooltipText(tip)
	ctx, _ := e.GetStyleContext()
	ctx.AddClass("invalid-entry")
}

func getComboIndex(c *gtk.ComboBox) int {
	iter, _ := c.GetActiveIter()
	model, _ := c.GetModel()
	path, _ := model.ToTreeModel().GetPath(iter)
	return path.GetIndices()[0]
}

func checkEntryPresence(e *gtk.Entry, name string) (string, bool) {
	text, _ := e.GetText()
	if len(text) == 0 {
		markInvalidEntry(&e.Widget, fmt.Sprintf("%s must be set", name))
		return "", false
	}
	return text, true
}

func checkEntryNumerical(e *gtk.Entry, bits int, name string) (uint64, bool) {
	text, ok := checkEntryPresence(e, name)
	if !ok {
		return 0, false
	}

	val, err := strconv.ParseUint(text, 10, bits)
	if err != nil {
		markInvalidEntry(&e.Widget, fmt.Sprintf("%s is invalid", name))
		return 0, false
	}
	return val, true
}
