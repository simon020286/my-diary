package utils

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

const menu1Name string = "menu1"
const menu2Name string = "menu2"
const mainName string = "main"
const msgName string = "msg"

type ViewUtils struct {
	storyView []string
	menuItems []string
	g         *gocui.Gui
	maxX      int
	maxY      int
}

func NewViewUtils(initialView string, g *gocui.Gui) *ViewUtils {
	maxX, maxY := g.Size()
	return &ViewUtils{
		storyView: []string{initialView},
		menuItems: []string{menu1Name, menu2Name},
		g:         g,
		maxX:      maxX - 1,
		maxY:      maxY,
	}
}

func (vu *ViewUtils) Size() (int, int) {
	return vu.maxX, vu.maxY
}

func (vu *ViewUtils) menuDimension(name string, totWidth int, totHeigth int) (x1 int, y1 int, x2 int, y2 int) {
	menuLength := len(vu.menuItems)
	indexEl := 0
	witdhEl := totWidth / menuLength
	y1 = 0
	y2 = totHeigth
	for i, v := range vu.menuItems {
		if v == name {
			indexEl = i
			break
		}
	}

	if indexEl == 0 {
		x1 = 0
	} else {
		x1 = indexEl * witdhEl
	}

	if indexEl == menuLength-1 {
		x2 = totWidth
	} else {
		x2 = (indexEl + 1) * witdhEl
	}

	return
}

func (vu *ViewUtils) CreateMenu(name string, title string, current bool, onInit func(v *gocui.View)) error {
	x1, y1, x2, y2 := vu.menuDimension(name, vu.maxX, vu.maxY/2)
	if v, err := vu.g.SetView(name, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = title
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		if onInit != nil {
			onInit(v)
		}

		if current {
			_, err := vu.g.SetCurrentView(menu1Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vu *ViewUtils) CreateContent(name string) error {
	if v, err := vu.g.SetView(mainName, 0, vu.maxY/2, vu.maxX, vu.maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Editable = true
		v.Wrap = true
	}
	return nil
}

func (vu *ViewUtils) GetCurrentViewName() string {
	return vu.storyView[len(vu.storyView)-1]
}

func (vu *ViewUtils) CursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		l, _ := v.Line(cy + 1)
		if l == "" {
			return nil
		}
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (vu *ViewUtils) CursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (vu *ViewUtils) NavigateBack(cleanCurrent bool) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		prevView := vu.storyView[len(vu.storyView)-2]
		if cleanCurrent {
			g.CurrentView().Clear()
			g.CurrentView().Title = ""
		}
		vu.storyView = vu.storyView[:len(vu.storyView)-1]
		_, err := g.SetCurrentView(prevView)
		if err != nil {
			return err
		}

		return nil
	}
}

func (vu *ViewUtils) NavigateTo(nextView string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		vu.storyView = append(vu.storyView, nextView)
		currentView, err := g.SetCurrentView(nextView)
		ox, oy := currentView.Origin()
		currentView.SetCursor(ox, oy)
		return err
	}
}

func (vu *ViewUtils) AddUpAndDown(viewName string) error {
	if err := vu.g.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, vu.CursorDown); err != nil {
		return err
	}
	if err := vu.g.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, vu.CursorUp); err != nil {
		return err
	}
	return nil
}

func (vu *ViewUtils) CurrentRowText(v *gocui.View) string {
	var l string
	var err error
	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	return l
}

func (vu *ViewUtils) ShowMessage(g *gocui.Gui, message string) error {
	return vu.ShowDialog(g, message, nil)
}

func (vu *ViewUtils) ShowConfirm(g *gocui.Gui, message string, onClose func(yesNo bool)) error {
	if _, err := vu.newDialogView(g, message); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		onDialogClose := func(yesNo bool) func() {
			return func() {
				if onClose != nil {
					onClose(yesNo)
				}
			}
		}
		if err := g.SetKeybinding(msgName, gocui.KeyCtrlY, gocui.ModNone, vu.closeDialog(onDialogClose(true))); err != nil {
			return err
		}
		if err := g.SetKeybinding(msgName, gocui.KeyCtrlN, gocui.ModNone, vu.closeDialog(onDialogClose(false))); err != nil {
			return err
		}

		if _, err := g.SetCurrentView(msgName); err != nil {
			return err
		}
	}

	return nil
}

func (vu *ViewUtils) newDialogView(g *gocui.Gui, message string) (*gocui.View, error) {
	messageLine := strings.Split(message, "\n")
	xAdjuster := 30
	if (vu.maxX/2 - xAdjuster) < 0 {
		xAdjuster = vu.maxX / 2
	}
	yAdgjuster := len(messageLine) / 2
	if yAdgjuster < 1 {
		yAdgjuster = 1
	} else if (vu.maxY/2 - yAdgjuster) < 0 {
		yAdgjuster = vu.maxY / 2
	}

	addLine := func() int {
		if (len(messageLine) % 2) == 0 {
			return yAdgjuster + 1
		}
		return yAdgjuster
	}

	if (len(messageLine) % 2) == 0 {

	}

	if v, err := g.SetView(msgName, vu.maxX/2-xAdjuster, vu.maxY/2-yAdgjuster, vu.maxX/2+xAdjuster, vu.maxY/2+addLine()); err != nil {
		if err != gocui.ErrUnknownView {
			return v, err
		}
		fmt.Fprintf(v, "%s", message)
		return v, err
	}

	return nil, nil
}

func (vu *ViewUtils) ShowDialog(g *gocui.Gui, message string, onClose func(text string)) error {
	if v, err := vu.newDialogView(g, message); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetKeybinding(msgName, gocui.KeyEnter, gocui.ModNone, vu.closeDialog(func() {
			if onClose != nil {
				onClose(v.Buffer())
			}
		})); err != nil {
			return err
		}
		if err := g.SetKeybinding(msgName, gocui.KeyCtrlQ, gocui.ModNone, vu.closeDialog(nil)); err != nil {
			return err
		}
		// fmt.Fprintf(v, "%s", message)
		v.Editable = true
		if _, err := g.SetCurrentView(msgName); err != nil {
			return err
		}
	}
	return nil
}

func (vu *ViewUtils) closeDialog(onClose func()) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if err := g.DeleteView(msgName); err != nil {
			return err
		}
		g.DeleteKeybindings(msgName)
		if _, err := g.SetCurrentView(vu.GetCurrentViewName()); err != nil {
			return err
		}
		if onClose != nil {
			onClose()
		}
		return nil
	}
}
