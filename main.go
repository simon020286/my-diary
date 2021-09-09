package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jroimartin/gocui"

	"mydiary/models"
	"mydiary/utils"
)

const menu1Name string = "menu1"
const menu2Name string = "menu2"
const mainName string = "main"

var currentDate string
var currentSection string
var viewUtils *utils.ViewUtils
var saveKey []string = []string{menu1Name, menu2Name, mainName}
var navigationKey []string = []string{menu1Name, menu2Name}

func getLine(content *models.Content) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		var l string = viewUtils.CurrentRowText(v)
		var err error

		sections, ok := content.GetSectionNames(l)
		if !ok {
			return nil
		}

		err = viewUtils.NavigateTo(menu2Name)(g, v)
		v, err = g.View(menu2Name)
		if err != nil {
			return err
		}

		v.Title = l
		v.Clear()

		for _, section := range sections {
			fmt.Fprintln(v, section)
		}

		currentDate = l
		currentSection = ""

		return nil
	}
}

func getContent(content *models.Content) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		var l string = viewUtils.CurrentRowText(v)
		var err error

		section, ok := content.GetSection(currentDate, l)
		if !ok {
			return fmt.Errorf("Key not found")
		}

		err = viewUtils.NavigateTo(mainName)(g, v)
		if err != nil {
			return err
		}

		v, err = g.View(mainName)
		if err != nil {
			return err
		}

		v.Title = l

		v.Clear()

		currentSection = section.Name

		fmt.Fprintf(v, "%s", section.Value)
		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func saveText(content *models.Content) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		err := utils.WriteFile(content)
		if err != err {
			return err
		}
		viewUtils.ShowMessage(g, "Text saved succesflully!")
		return nil
	}
}

func apply(content *models.Content) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Rewind()
		section, ok := content.GetSection(currentDate, currentSection)
		if !ok {
			return fmt.Errorf("Section not found")
		}

		section.Value = v.Buffer()
		return nil
	}
}

func addLine(content *models.Content, options models.AddLineOptions) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return viewUtils.ShowDialog(g, options.DefaultValue, func(text string) {
			text = utils.RemoveNewLine(text)
			currentView := v.Name()
			isNew := false
			if currentView == menu1Name {
				isNew = content.AddKey(text)
			} else if currentView == menu2Name {
				_, isNew = content.AddSection(currentDate, text)
			}

			if isNew {
				fmt.Fprintln(v, text)
				v.SetCursor(0, len(v.BufferLines())-2)
			} else {
				viewUtils.ShowMessage(g, "Item alredy exists!")
			}
		})
	}
}

func removeLine(content *models.Content) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		var l string = viewUtils.CurrentRowText(v)
		return viewUtils.ShowConfirm(g, fmt.Sprintf("Vuoi eliminare %s?\n^Y Yes ^N No", l), func(yesNo bool) {
			if yesNo {
				if v.Name() == menu1Name {
					content.RemoveKey(l)
					x1, x2 := v.Origin()
					v.SetCursor(x1, x2)
					v.Clear()
					for _, key := range content.Keys() {
						fmt.Fprintln(v, key)
					}
				} else if v.Name() == menu2Name {
					content.RemoveSection(currentDate, l)
					x1, x2 := v.Origin()
					v.SetCursor(x1, x2)
					v.Clear()
					sections, ok := content.GetSectionNames(currentDate)
					if !ok {
						return
					}
					for index, section := range sections {
						fmt.Fprintln(v, section)
						if index == 0 {
							currentSection = section
						}
					}
				}
			}
		})
	}
}

func toEndLine(g *gocui.Gui, v *gocui.View) error {
	l := viewUtils.CurrentRowText(v)
	_, y := v.Cursor()
	v.SetCursor(len(l), y)
	return nil
}

func toStartLine(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()
	v.SetCursor(0, y)
	return nil
}

func keybindings(g *gocui.Gui, content *models.Content) error {
	keys := GetKeyBindings(content)
	for _, key := range keys {
		key.Set(g)
	}
	return nil
}

func layout(content *models.Content) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		if err := viewUtils.CreateMenu(menu1Name, "Date", true, func(v *gocui.View) {
			for _, key := range content.Keys() {
				fmt.Fprintln(v, key)
			}
		}); err != nil {
			return err
		}
		if err := viewUtils.CreateMenu(menu2Name, "", false, nil); err != nil {
			return err
		}
		if err := viewUtils.CreateContent(mainName); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
		return nil
	}
}

func main() {
	args := os.Args
	date := time.Now()
	filename := fmt.Sprintf("%02d-%d.md", date.Month(), date.Year())
	if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			if args[i] == "-f" || args[i] == "--file" {
				filename = args[i+1]
				i++
			}
		}
	}

	content, err := utils.ReadFile(filename)
	if err != nil {
		log.Panicln(err)
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	viewUtils = utils.NewViewUtils(menu1Name, g)

	g.Cursor = true

	g.SetManagerFunc(layout(content))

	if err := keybindings(g, content); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
