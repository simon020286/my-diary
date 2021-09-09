package main

import (
	"mydiary/models"
	"time"

	"github.com/jroimartin/gocui"
)

type KeyBinding struct {
	ViewNames []string
	Key       gocui.Key
	Handler   func(*gocui.Gui, *gocui.View) error
}

func (k *KeyBinding) Set(g *gocui.Gui) error {
	for _, v := range k.ViewNames {
		if err := g.SetKeybinding(v, k.Key, gocui.ModNone, k.Handler); err != nil {
			return err
		}
	}
	return nil
}

func newKeybinding1(viewName string, key gocui.Key, handler func(*gocui.Gui, *gocui.View) error) KeyBinding {
	return KeyBinding{
		ViewNames: []string{viewName},
		Key:       key,
		Handler:   handler,
	}
}

func newKeybinding2(viewNames []string, key gocui.Key, handler func(*gocui.Gui, *gocui.View) error) KeyBinding {
	return KeyBinding{
		ViewNames: viewNames,
		Key:       key,
		Handler:   handler,
	}
}

func GetKeyBindings(content *models.Content) []KeyBinding {
	return []KeyBinding{
		newKeybinding1("", gocui.KeyCtrlC, quit),
		newKeybinding1(mainName, gocui.KeyCtrlA, apply(content)),
		newKeybinding1(mainName, gocui.KeyEnd, toEndLine),
		newKeybinding1(mainName, gocui.KeyHome, toStartLine),
		newKeybinding2([]string{menu1Name, menu2Name}, gocui.KeyArrowDown, viewUtils.CursorDown),
		newKeybinding2([]string{menu1Name, menu2Name}, gocui.KeyArrowUp, viewUtils.CursorUp),
		newKeybinding2([]string{menu1Name, menu2Name, mainName}, gocui.KeyCtrlS, saveText(content)),
		newKeybinding1(menu1Name, gocui.KeyEnter, getLine(content)),
		newKeybinding1(menu1Name, gocui.KeyCtrlA, addLine(content, models.AddLineOptions{DefaultValue: time.Now().Format("02-01-2006")})),
		newKeybinding2([]string{menu1Name, menu2Name}, gocui.KeyCtrlD, removeLine(content)),
		newKeybinding1(menu2Name, gocui.KeyEnter, getContent(content)),
		newKeybinding2([]string{menu2Name, mainName}, gocui.KeyCtrlB, viewUtils.NavigateBack(true)),
		newKeybinding1(menu2Name, gocui.KeyCtrlA, addLine(content, models.AddLineOptions{})),
	}
}
