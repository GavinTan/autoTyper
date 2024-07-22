package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
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

	"github.com/gavintan/autoTyper/config"
	"github.com/gavintan/autoTyper/ipc"
	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	h                  *Hotkey
	list               *widget.List
	hotkeyItem         HotkeyData
	inputKeys          []string
	firstKey           string
	lastKey            string
	autostart          bool
	inputWindow        fyne.Window
	registerHotkeyList = map[string]*hotkey.Hotkey{}
)

type Input struct {
	widget.Entry
}

type MultiLineInput struct {
	widget.Entry
}

type HotkeyData struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	HotkeyText string
	Hotkey     string `gorm:"uniqueIndex"`
}

type Hotkey struct {
	db *gorm.DB
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

func (me *MultiLineInput) TypedKey(key *fyne.KeyEvent) {
	// 禁止换行
	keyName := string(key.Name)
	if strings.Contains(keyName, "Enter") || strings.Contains(keyName, "Return") {
		return
	}

	me.Entry.TypedKey(key)
}

func NewInput() *Input {
	e := &Input{}
	e.Wrapping = fyne.TextTruncate
	e.ExtendBaseWidget(e)
	return e
}

func NewMultiLineInput() *MultiLineInput {
	e := &MultiLineInput{}
	e.Wrapping = fyne.TextWrapBreak
	e.MultiLine = true
	e.ExtendBaseWidget(e)
	return e
}

func NewHotkey() *Hotkey {
	userConfigDir, _ := os.UserConfigDir()
	dataPath := path.Join(userConfigDir, config.Name)
	dbPath := path.Join(dataPath, "data.db")
	os.MkdirAll(dataPath, os.ModePerm)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		notification("打开数据文件失败，请检查！")
	}
	db.Table(config.Name).AutoMigrate(&HotkeyData{})

	return &Hotkey{db: db.Table(config.Name).WithContext(context.Background())}
}

func (h Hotkey) ListData() binding.StringList {
	dataList := binding.NewStringList()
	for _, item := range h.All() {
		d, _ := json.Marshal(item)
		dataList.Append(string(d))
	}

	return dataList
}

func (h Hotkey) All() []HotkeyData {
	var data []HotkeyData
	h.db.Find(&data)
	return data
}

func (h Hotkey) Create(data *HotkeyData) error {
	if err := h.db.Create(&data).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("热键已被使用！")
		}
	}

	return nil
}

func (h Hotkey) Update(data HotkeyData) error {
	if err := h.db.Model(&data).Updates(data).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("热键已被使用！")
		}
	}

	return nil
}

func (h Hotkey) Delete(id uint) {
	h.db.Unscoped().Delete(&HotkeyData{}, id)
}

func (h Hotkey) Rload(dataList binding.StringList) {
	newData, _ := h.ListData().Get()
	dataList.Set(newData)
	list.Refresh()
}

func (h Hotkey) Close() {
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

func openInputWindow() {
	inputWindow = fyne.CurrentApp().NewWindow("模拟按键输入模式")
	inputWindow.SetFixedSize(true)

	textInput := widget.NewMultiLineEntry()
	textInput.Validator = validation.NewRegexp(`\S`, "必填")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "文本", Widget: textInput, HintText: "不支持中文以及非键盘上的符号"},
		},
		OnCancel: func() {
			inputWindow.Close()
			inputWindow = nil
		},
		OnSubmit: func() {
			d := dialog.NewConfirm("Confirmation", "请按Yes后3秒激活接收窗口", func(b bool) {
				if b {
					inputWindow.Hide()
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
					inputWindow.Close()
					inputWindow = nil
				}
			}, inputWindow)
			d.Show()
		},
		SubmitText: "提交",
		CancelText: "取消",
	}

	inputWindow.SetCloseIntercept(func() {
		inputWindow.Close()
		inputWindow = nil
	})

	inputWindow.SetContent(form)
	inputWindow.Resize(fyne.NewSize(360, 160))
	inputWindow.CenterOnScreen()
	inputWindow.Show()
}

func registerWindowInputHotkey() {
	// 解决darwin下robotgo panic ？
	if runtime.GOOS == "darwin" {
		robotgo.KeyTap("1")
	}

	go func() {
		hk := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF10)
		if err := hk.Register(); err != nil {
			notification("注册「F10」热键失败！")
		}

		for range hk.Keydown() {
			if inputWindow != nil {
				inputWindow.Show()
			} else {
				openInputWindow()
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
				notification(fmt.Sprintf("注册「%s」热键失败！", data.Hotkey))
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

func registerAllHotkey(data []HotkeyData) {
	for _, v := range data {
		registerHotkey(v)
	}
}

func unRegisterAllHotkey() {
	for _, v := range h.All() {
		unRegisterHotkey(v)
	}
}

func init() {
	c, _ := ipc.Connect()
	if c != nil {
		c.Show()
		os.Exit(0)
	}
}

func main() {
	a := app.NewWithID(config.Title)
	a.SetIcon(resourceIconPng)
	mainWindow := a.NewWindow(config.Title)

	go ipc.NewServer(mainWindow)

	h = NewHotkey()
	dataList := h.ListData()

	openAddDialog := func() {
		d := NewDialog("添加", HotkeyData{})
		pop := widget.NewModalPopUp(d, mainWindow.Canvas())

		d.OnSubmit = func() {
			d := HotkeyData{HotkeyText: d.textInput.Text, Hotkey: d.hotkeyInput.Text}
			if err := h.Create(&d); err != nil {
				showErrorMessage(err.Error(), mainWindow)
			} else {
				dj, _ := json.Marshal(&d)
				dataList.Append(string(dj))
				list.Refresh()
				registerHotkey(d)
				pop.Hide()
			}
		}

		d.OnCancel = func() {
			pop.Hide()
		}

		pop.Resize(fyne.NewSize(360, 150))
		pop.Show()

		w := a.NewWindow("添加")
		w.SetFixedSize(true)

		textInput := widget.NewMultiLineEntry()
		textInput.Validator = validation.NewRegexp(`\S`, "必填")
		hotkeyInput := NewInput()
		hotkeyInput.Validator = validation.NewRegexp(`\S`, "必填")
	}

	openEditDialog := func(item HotkeyData) {
		d := NewDialog("编辑", item)
		pop := widget.NewModalPopUp(d, mainWindow.Canvas())

		d.OnSubmit = func() {
			if item.HotkeyText == d.textInput.Text && item.Hotkey == d.hotkeyInput.Text {
			} else {
				if err := h.Update(HotkeyData{ID: item.ID, HotkeyText: d.textInput.Text, Hotkey: d.hotkeyInput.Text}); err != nil {
					showErrorMessage(err.Error(), mainWindow)
				} else {
					h.Rload(dataList)
					unRegisterHotkey(hotkeyItem)
					hotkeyItem.HotkeyText = d.textInput.Text
					hotkeyItem.Hotkey = d.hotkeyInput.Text
					registerHotkey(hotkeyItem)
					pop.Hide()
					showInformationMessage("更新成功", mainWindow)
				}
			}
		}

		d.OnCancel = func() {
			pop.Hide()
		}

		pop.Resize(fyne.NewSize(360, 150))
		pop.Show()
	}

	list = widget.NewListWithData(
		dataList,
		func() fyne.CanvasObject {
			return container.NewGridWithRows(1, widget.NewLabel(""), canvas.NewText("", color.RGBA{204, 204, 204, 255}))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			var hotkeyItem HotkeyData
			item, _ := i.(binding.String).Get()
			json.Unmarshal([]byte(item), &hotkeyItem)

			l := o.(*fyne.Container).Objects[0].(*widget.Label)
			l.SetText(hotkeyItem.HotkeyText)
			l.Truncation = fyne.TextTruncateEllipsis

			t := o.(*fyne.Container).Objects[1].(*canvas.Text)
			t.Text = hotkeyItem.Hotkey + "  "
			t.Alignment = fyne.TextAlignTrailing
		})

	list.OnSelected = func(id widget.ListItemID) {
		selItem, _ := dataList.GetValue(id)
		json.Unmarshal([]byte(selItem), &hotkeyItem)
	}

	addBtn := widget.NewButtonWithIcon("添加", theme.ContentAddIcon(), func() {
		openAddDialog()
	})
	editBtn := widget.NewButtonWithIcon("编辑", theme.DocumentCreateIcon(), func() {
		if hotkeyItem.Hotkey != "" {
			openEditDialog(hotkeyItem)
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
	toolsItem := fyne.NewMenuItem("模拟按键", func() {
		if inputWindow != nil {
			inputWindow.Show()
		} else {
			openInputWindow()
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
	registerAllHotkey(h.All())

	mainWindow.Resize(fyne.NewSize(500, 400))
	mainWindow.CenterOnScreen()

	flag.BoolVar(&autostart, "autostart", false, "Run without window.")
	flag.Parse()

	if autostart {
		a.Run()
	} else {
		mainWindow.ShowAndRun()
	}

}
