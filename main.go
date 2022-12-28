package main

import (
	"encoding/json"
	"errors"
	"flag"
	"image/color"
	"net/url"
	"os"
	"path"
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
	Text   string
	Hotkey string `gorm:"uniqueIndex"`
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
		notification("打开数据文件失败，请检查！")
	}
	db.Table("hotkey").AutoMigrate(&HotkeyData{})

	return &HotkeyObj{db: db.Table("hotkey")}
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
			return errors.New("热键已被使用！")
		}
	}

	return nil
}

func (h HotkeyObj) Update(id uint, data HotkeyData) error {
	if err := h.db.Exec("UPDATE hotkey SET text=?, hotkey=? WHERE id=?", data.Text, data.Hotkey, id).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("热键已被使用！")
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
	w := fyne.CurrentApp().NewWindow("模拟按键输入模式")
	w.SetFixedSize(true)

	textInput := widget.NewMultiLineEntry()
	textInput.Validator = validation.NewRegexp(`\S`, "必填")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "文本", Widget: textInput, HintText: "不支持中文以及非键盘上的符号"},
		},
		OnCancel: func() {
			w.Close()
			specialModeShow = false
		},
		OnSubmit: func() {
			d := dialog.NewConfirm("Confirmation", "请按Yes后3秒激活接收窗口", func(b bool) {
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
				}
			}, w)
			d.Show()
		},
		SubmitText: "提交",
		CancelText: "取消",
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
			notification("注册热键失败！")
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
				notification("注册热键失败！")
			}

			registerHotkeyList[data.Hotkey] = hk

			for range hk.Keydown() {
				robotgo.TypeStr(data.Text)
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
					notification("注册热键失败！")
				}

				registerHotkeyList[v.Hotkey] = hk

				for range hk.Keydown() {
					robotgo.TypeStr(v.Text)
				}
			}

		}()
	}
}

func main() {
	title := "文本自动输入器"
	a := app.NewWithID(title)
	a.Settings().SetTheme(&mainTheme{})
	a.SetIcon(resourceIconPng)
	mainWindow := a.NewWindow(title)

	h := Hotkey()
	dataList := h.ListData()

	openAddWindow := func() {
		w := a.NewWindow("添加")
		w.SetFixedSize(true)

		textInput := widget.NewMultiLineEntry()
		textInput.Validator = validation.NewRegexp(`\S`, "必填")
		hotkeyInput := NewInput()
		hotkeyInput.Validator = validation.NewRegexp(`\S`, "必填")

		form := &widget.Form{
			Items: []*widget.FormItem{
				widget.NewFormItem("文本", textInput),
				widget.NewFormItem("热键", hotkeyInput),
			},
			OnCancel: func() {
				w.Close()
				addShow = false
			},
			OnSubmit: func() {
				d := HotkeyData{Text: textInput.Text, Hotkey: hotkeyInput.Text}
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
			SubmitText: "提交",
			CancelText: "取消",
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
		w := a.NewWindow("编辑")
		w.SetFixedSize(true)

		textInput := widget.NewMultiLineEntry()
		textInput.SetText(item.Text)
		hotkeyInput := NewInput()
		hotkeyInput.SetText(item.Hotkey)

		form := &widget.Form{
			Items: []*widget.FormItem{
				{Text: "文本", Widget: textInput},
				{Text: "热键", Widget: hotkeyInput},
			},
			OnCancel: func() {
				w.Close()
				editShow = false
			},
			OnSubmit: func() {
				if item.Text == textInput.Text && item.Hotkey == hotkeyInput.Text {

				} else {
					if err := h.Update(item.ID, HotkeyData{Text: textInput.Text, Hotkey: hotkeyInput.Text}); err != nil {
						showErrorMessage(err.Error(), mainWindow)
					} else {
						h.Rload(dataList)
						unRegisterHotkey(hotkeyItem)
						hotkeyItem.Text = textInput.Text
						hotkeyItem.Hotkey = hotkeyInput.Text
						registerHotkey(hotkeyItem)
						showInformationMessage("更新成功", mainWindow)
					}
				}

				w.Close()
				editShow = false
			},
			SubmitText: "提交",
			CancelText: "取消",
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
			text.SetText(hotkeyItem.Text)

			text1 := o.(*fyne.Container).Objects[1].(*canvas.Text)
			text1.Text = hotkeyItem.Hotkey
		})

	list.OnSelected = func(id widget.ListItemID) {
		selItem, _ := dataList.GetValue(id)
		json.Unmarshal([]byte(selItem), &hotkeyItem)
	}

	addBtn := widget.NewButtonWithIcon("添加", theme.ContentAddIcon(), func() {
		if addShow {
			addWindow.Show()
		} else {
			openAddWindow()
		}
	})
	editBtn := widget.NewButtonWithIcon("编辑", theme.DocumentCreateIcon(), func() {
		if hotkeyItem.Hotkey != "" {
			if editShow {
				editWindow.Show()
			} else {
				openEditWindow(hotkeyItem)
			}
		} else {
			d := dialog.NewError(errors.New("请选择！"), mainWindow)
			d.Resize(fyne.NewSize(200, 150))
			d.Show()
		}
	})
	delBtn := widget.NewButtonWithIcon("删除", theme.DeleteIcon(), func() {
		if hotkeyItem.Hotkey != "" {
			d := dialog.NewConfirm("Confirmation", "确认删除吗？", func(b bool) {
				if b {
					list.UnselectAll()
					h.Delete(hotkeyItem.ID)
					h.Rload(dataList)
					unRegisterHotkey(hotkeyItem)
					hotkeyItem = HotkeyData{}
				}
			}, mainWindow)
			d.SetConfirmText("是")
			d.SetDismissText("否")
			d.Resize(fyne.NewSize(250, 150))
			d.Show()
		} else {
			showErrorMessage("请选择！", mainWindow)
		}
	})

	content := container.NewBorder(nil, container.NewGridWithColumns(3, addBtn, editBtn, delBtn), nil, nil, list)

	f10 := &desktop.CustomShortcut{KeyName: fyne.KeyF10}
	toolsItem := fyne.NewMenuItem("特殊模式", func() {
		if specialModeShow {
			specialModeWindow.Show()
		} else {
			openWindowInput()
		}
	})
	toolsItem.Shortcut = f10

	settingsItem := fyne.NewMenuItem("开机自启", nil)
	if GetAutoStartStatus() {
		settingsItem.Checked = true
	}

	helpMenu := fyne.NewMenu("帮助",
		fyne.NewMenuItem("文档", func() {
			u, _ := url.Parse("https://github.com/gavintan/autoTyper")
			_ = a.OpenURL(u)
		}))

	// mainWindow.Canvas().AddShortcut(f10, func(shortcut fyne.Shortcut) {
	// 	openWindowInput()
	// })

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("文件"),
		fyne.NewMenu("设置", settingsItem),
		fyne.NewMenu("工具", toolsItem),
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
