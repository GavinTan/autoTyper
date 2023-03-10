package main

import (
	"encoding/json"
	"errors"
	"flag"
	"image/color"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	list               *widget.List
	hotkeyItem         HotkeyData
	inputKeys          []string
	firstKey           string
	lastKey            string
	autostart          bool
	specialModeShow    bool
	specialModeWindow  fyne.Window
	addShow            bool
	addWindow          fyne.Window
	editShow           bool
	editWindow         fyne.Window
	registerHotkeyList = map[string]*hotkey.Hotkey{}
)

type Input struct {
	widget.Entry
	Wrapping fyne.TextWrap
}

type HotkeyData struct {
	gorm.Model
	HotkeyText string
	Hotkey     string `gorm:"uniqueIndex"`
}

type HotkeyObj struct {
	db *gorm.DB
}

func NewInput() *Input {
	e := &Input{Wrapping: fyne.TextTruncate}
	e.ExtendBaseWidget(e)
	return e
}

func (e *Input) KeyDown(key *fyne.KeyEvent) {
	keyName := string(key.Name)
	if name, ok := ModKeyName[string(key.Name)]; ok {
		keyName = name
	}

	if len(inputKeys) > 0 {
		if _, ok := Modkeys[inputKeys[len(inputKeys)-1]]; ok {
			inputKeys = append(inputKeys, keyName)
		}
	} else {
		firstKey = keyName
		inputKeys = append(inputKeys, keyName)
	}

	lastKey = keyName

	e.CursorColumn = len(strings.Join(inputKeys, " + "))
	e.SetText(strings.Join(inputKeys, " + "))
}

func (e *Input) KeyUp(key *fyne.KeyEvent) {
	keyName := string(key.Name)
	if name, ok := ModKeyName[string(key.Name)]; ok {
		keyName = name
	}

	_, ok := Modkeys[lastKey]

	for i := 0; i < len(inputKeys); i++ {
		if inputKeys[i] == keyName {
			inputKeys = append(inputKeys[:i], inputKeys[i+1:]...)
		}
	}

	if len(inputKeys) == 0 && ok {
		e.SetText("")
	}

	k := strings.Split(e.Text, "+")
	if len(inputKeys) == 0 && len(k) == 2 && firstKey == "Alt" {
		e.SetText("")
	}
}

func (e *Input) TypedRune(r rune) {
}

func Hotkey() *HotkeyObj {
	userConfigDir, _ := os.UserConfigDir()
	dataPath := path.Join(userConfigDir, "autoTyper")
	dbPath := path.Join(dataPath, "data.db")
	os.MkdirAll(dataPath, os.ModePerm)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		notification("???????????????????????????????????????")
	}
	db.Table("autoTyper").AutoMigrate(&HotkeyData{})

	return &HotkeyObj{db: db.Table("autoTyper")}
}

func (h HotkeyObj) ListData() binding.StringList {
	var data []HotkeyData
	h.db.Find(&data)
	dataList := binding.NewStringList()
	for _, item := range data {
		d, _ := json.Marshal(item)
		dataList.Append(string(d))
	}

	return dataList
}

func (h HotkeyObj) All() []HotkeyData {
	var data []HotkeyData
	h.db.Find(&data)
	return data
}

func (h HotkeyObj) Create(data *HotkeyData) error {
	if err := h.db.Create(&data).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("?????????????????????")
		}
	}

	return nil
}

func (h HotkeyObj) Update(id uint, data HotkeyData) error {
	if err := h.db.Exec("UPDATE hotkey SET text=?, hotkey=? WHERE id=?", data.HotkeyText, data.Hotkey, id).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("?????????????????????")
		}
	}

	return nil
}

func (h HotkeyObj) Delete(id uint) {
	h.db.Exec("DELETE FROM hotkey WHERE id=?", id)
}

func (h HotkeyObj) Rload(dataList binding.StringList) {
	newData, _ := h.ListData().Get()
	dataList.Set(newData)
	list.Refresh()
}

func (h HotkeyObj) Close() {
	if db, err := h.db.DB(); err == nil {
		db.Close()
	}
}

func showErrorMessage(message string, w fyne.Window) {
	d := dialog.NewError(errors.New(message), w)
	d.Resize(fyne.NewSize(200, 150))
	d.Show()
}

func showInformationMessage(message string, w fyne.Window) {
	d := dialog.NewInformation("Information", message, w)
	d.Resize(fyne.NewSize(200, 150))
	d.Show()
}

func notification(message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "",
		Content: message,
	})
}

func openWindowInput() {
	w := fyne.CurrentApp().NewWindow("????????????????????????")
	w.SetFixedSize(true)

	textInput := widget.NewMultiLineEntry()
	textInput.Validator = validation.NewRegexp(`\S`, "??????")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "??????", Widget: textInput, HintText: "??????????????????????????????????????????"},
		},
		OnCancel: func() {
			w.Close()
			specialModeShow = false
		},
		OnSubmit: func() {
			d := dialog.NewConfirm("Confirmation", "??????Yes???3?????????????????????", func(b bool) {
				if b {
					w.Hide()
					robotgo.Sleep(3)
					shiftStr := "~!@#$%^&*()_+|{}:\"<>?"
					for _, str := range textInput.Text {
						robotgo.KeySleep = 100
						if strings.Contains(shiftStr, string(str)) || unicode.IsUpper(str) {
							robotgo.KeyTap(string(str), []string{"shift"})
						} else {
							robotgo.KeyTap(string(str))
						}
					}
					robotgo.KeyTap("enter")
					w.Close()
					specialModeShow = false
				}
			}, w)
			d.Show()
		},
		SubmitText: "??????",
		CancelText: "??????",
	}

	w.SetCloseIntercept(func() {
		w.Close()
		specialModeShow = false
	})

	w.SetContent(form)
	w.Resize(fyne.NewSize(360, 160))
	w.Show()

	specialModeShow = true
	specialModeWindow = w
}

func registerWindowInputHotkey() {
	go func() {
		hk := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF10)
		if err := hk.Register(); err != nil {
			notification("?????????????????????")
		}

		if runtime.GOOS == "darwin" {
			robotgo.KeyTap("1")
			robotgo.KeyTap("backspace")
		}

		for range hk.Keydown() {
			if specialModeShow {
				specialModeWindow.Show()
			} else {
				openWindowInput()
			}
		}
	}()
}

func registerHotkey(data HotkeyData) {
	k := strings.Split(data.Hotkey, " + ")

	go func() {
		if len(k) > 1 {
			var mk []hotkey.Modifier

			for i, v := range k {
				if i+1 != len(k) {
					mk = append(mk, hotkey.Modifier(Modkeys[v]))
				}
			}

			hk := hotkey.New(mk, Keys[k[len(k)-1]])
			if err := hk.Register(); err != nil {
				notification("?????????????????????")
			}

			registerHotkeyList[data.Hotkey] = hk

			for range hk.Keydown() {
				robotgo.TypeStr(data.HotkeyText)
			}

			hk.Unregister()
		}

	}()
}

func unRegisterHotkey(data HotkeyData) {
	go func() {
		if hk, ok := registerHotkeyList[data.Hotkey]; ok {
			hk.Unregister()
		}
	}()
}

func registerAllHotkey() {
	for _, v := range Hotkey().All() {
		v := v
		k := strings.Split(v.Hotkey, " + ")

		go func() {
			if len(k) > 1 {
				var mk []hotkey.Modifier

				for i, v := range k {
					if i+1 != len(k) {
						mk = append(mk, hotkey.Modifier(Modkeys[v]))
					}
				}

				hk := hotkey.New(mk, Keys[k[len(k)-1]])
				if err := hk.Register(); err != nil {
					notification("?????????????????????")
				}

				registerHotkeyList[v.Hotkey] = hk

				for range hk.Keydown() {
					robotgo.TypeStr(v.HotkeyText)
				}
			}

		}()
	}
}

func main() {
	title := "?????????????????????"
	a := app.NewWithID(title)
	a.Settings().SetTheme(&mainTheme{})
	a.SetIcon(resourceIconPng)
	mainWindow := a.NewWindow(title)

	h := Hotkey()
	dataList := h.ListData()

	openAddWindow := func() {
		w := a.NewWindow("??????")
		w.SetFixedSize(true)

		textInput := widget.NewMultiLineEntry()
		textInput.Validator = validation.NewRegexp(`\S`, "??????")
		hotkeyInput := NewInput()
		hotkeyInput.Validator = validation.NewRegexp(`\S`, "??????")

		form := &widget.Form{
			Items: []*widget.FormItem{
				widget.NewFormItem("??????", textInput),
				widget.NewFormItem("??????", hotkeyInput),
			},
			OnCancel: func() {
				w.Close()
				addShow = false
			},
			OnSubmit: func() {
				d := HotkeyData{HotkeyText: textInput.Text, Hotkey: hotkeyInput.Text}
				if err := h.Create(&d); err != nil {
					showErrorMessage(err.Error(), mainWindow)
				} else {
					dj, _ := json.Marshal(&d)
					dataList.Append(string(dj))
					registerHotkey(d)
				}

				w.Close()
				addShow = false
			},
			SubmitText: "??????",
			CancelText: "??????",
		}

		w.SetCloseIntercept(func() {
			w.Close()
			addShow = false
		})

		w.SetContent(form)
		w.Resize(fyne.NewSize(360, 150))
		w.Show()
		addShow = true
		addWindow = w
	}

	openEditWindow := func(item HotkeyData) {
		w := a.NewWindow("??????")
		w.SetFixedSize(true)

		textInput := widget.NewMultiLineEntry()
		textInput.SetText(item.HotkeyText)
		hotkeyInput := NewInput()
		hotkeyInput.SetText(item.Hotkey)

		form := &widget.Form{
			Items: []*widget.FormItem{
				{Text: "??????", Widget: textInput},
				{Text: "??????", Widget: hotkeyInput},
			},
			OnCancel: func() {
				w.Close()
				editShow = false
			},
			OnSubmit: func() {
				if item.HotkeyText == textInput.Text && item.Hotkey == hotkeyInput.Text {

				} else {
					if err := h.Update(item.ID, HotkeyData{HotkeyText: textInput.Text, Hotkey: hotkeyInput.Text}); err != nil {
						showErrorMessage(err.Error(), mainWindow)
					} else {
						h.Rload(dataList)
						unRegisterHotkey(hotkeyItem)
						hotkeyItem.HotkeyText = textInput.Text
						hotkeyItem.Hotkey = hotkeyInput.Text
						registerHotkey(hotkeyItem)
						showInformationMessage("????????????", mainWindow)
					}
				}

				w.Close()
				editShow = false
			},
			SubmitText: "??????",
			CancelText: "??????",
		}

		w.SetCloseIntercept(func() {
			w.Close()
			editShow = false
		})

		w.SetContent(form)
		w.Resize(fyne.NewSize(360, 150))
		w.Show()
		editShow = true
		editWindow = w
	}

	list = widget.NewListWithData(
		dataList,
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, widget.NewLabel(""), canvas.NewText("", color.RGBA{204, 204, 204, 255}))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			var hotkeyItem HotkeyData
			item, _ := i.(binding.String).Get()

			json.Unmarshal([]byte(item), &hotkeyItem)

			text := o.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(hotkeyItem.HotkeyText)

			text1 := o.(*fyne.Container).Objects[1].(*canvas.Text)
			text1.Text = hotkeyItem.Hotkey
		})

	list.OnSelected = func(id widget.ListItemID) {
		selItem, _ := dataList.GetValue(id)
		json.Unmarshal([]byte(selItem), &hotkeyItem)
	}

	addBtn := widget.NewButtonWithIcon("??????", theme.ContentAddIcon(), func() {
		if addShow {
			addWindow.Show()
		} else {
			openAddWindow()
		}
	})
	editBtn := widget.NewButtonWithIcon("??????", theme.DocumentCreateIcon(), func() {
		if hotkeyItem.Hotkey != "" {
			if editShow {
				editWindow.Show()
			} else {
				openEditWindow(hotkeyItem)
			}
		} else {
			d := dialog.NewError(errors.New("????????????"), mainWindow)
			d.Resize(fyne.NewSize(200, 150))
			d.Show()
		}
	})
	delBtn := widget.NewButtonWithIcon("??????", theme.DeleteIcon(), func() {
		if hotkeyItem.Hotkey != "" {
			d := dialog.NewConfirm("Confirmation", "??????????????????", func(b bool) {
				if b {
					list.UnselectAll()
					h.Delete(hotkeyItem.ID)
					h.Rload(dataList)
					unRegisterHotkey(hotkeyItem)
					hotkeyItem = HotkeyData{}
				}
			}, mainWindow)
			d.SetConfirmText("???")
			d.SetDismissText("???")
			d.Resize(fyne.NewSize(250, 150))
			d.Show()
		} else {
			showErrorMessage("????????????", mainWindow)
		}
	})

	content := container.NewBorder(nil, container.NewGridWithColumns(3, addBtn, editBtn, delBtn), nil, nil, list)

	f10 := &desktop.CustomShortcut{KeyName: fyne.KeyF10}
	toolsItem := fyne.NewMenuItem("????????????", func() {
		if specialModeShow {
			specialModeWindow.Show()
		} else {
			openWindowInput()
		}
	})
	toolsItem.Shortcut = f10

	settingsItem := fyne.NewMenuItem("????????????", nil)
	if GetAutoStartStatus() {
		settingsItem.Checked = true
	}

	helpMenu := fyne.NewMenu("??????",
		fyne.NewMenuItem("??????", func() {
			u, _ := url.Parse("https://github.com/gavintan/autoTyper")
			_ = a.OpenURL(u)
		}))

	// mainWindow.Canvas().AddShortcut(f10, func(shortcut fyne.Shortcut) {
	// 	openWindowInput()
	// })

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("??????"),
		fyne.NewMenu("??????", settingsItem),
		fyne.NewMenu("??????", toolsItem),
		helpMenu,
	)

	settingsItem.Action = func() {
		settingsItem.Checked = !settingsItem.Checked
		AutoStart(settingsItem.Checked)
		mainMenu.Refresh()
	}

	mainWindow.SetMainMenu(mainMenu)

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("autoTyper",
			fyne.NewMenuItem("Show", func() {
				mainWindow.Show()
			}))

		desk.SetSystemTrayMenu(m)
	}

	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})

	mainWindow.SetContent(content)

	registerWindowInputHotkey()
	registerAllHotkey()

	mainWindow.Resize(fyne.NewSize(500, 400))

	flag.BoolVar(&autostart, "autostart", false, "Run without window.")
	flag.Parse()

	if autostart {
		a.Run()
	} else {
		mainWindow.ShowAndRun()
	}

}
