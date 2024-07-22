package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type cDialog struct {
	widget.BaseWidget

	Title      string
	Hotkey     string
	HotkeyText string
	OnSubmit   func()
	OnCancel   func()

	container   *fyne.Container
	textInput   *MultiLineInput
	hotkeyInput *Input
	confirmBtn  *widget.Button
}

func (c *cDialog) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

func (a *cDialog) onSubmit() {
	if a.OnSubmit != nil {
		a.OnSubmit()
	}
}

func (a *cDialog) onCancel() {
	if a.OnCancel != nil {
		a.OnCancel()
	}
}

func NewDialog(title string, item HotkeyData) *cDialog {
	c := &cDialog{}
	c.ExtendBaseWidget(c)

	c.textInput = NewMultiLineInput()
	c.textInput.SetText(item.HotkeyText)
	c.textInput.Validator = validation.NewRegexp(`\S`, "required")
	c.textInput.SetOnValidationChanged(func(err error) {
		if err != nil || c.hotkeyInput.Validate() != nil {
			c.confirmBtn.Disable()
		} else {
			c.confirmBtn.Enable()
		}
	})

	c.hotkeyInput = NewInput()
	c.hotkeyInput.SetText(item.Hotkey)
	c.hotkeyInput.Validator = validation.NewRegexp(`\S`, "required")
	c.hotkeyInput.SetOnValidationChanged(func(err error) {
		if err != nil || c.textInput.Validate() != nil {
			c.confirmBtn.Disable()
		} else {
			c.confirmBtn.Enable()
		}
	})

	c.confirmBtn = widget.NewButton("确认", c.onSubmit)
	c.confirmBtn.Importance = widget.HighImportance
	c.confirmBtn.Icon = theme.ConfirmIcon()

	cancelBtn := widget.NewButton("取消", c.onCancel)
	cancelBtn.Icon = theme.CancelIcon()

	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle.Bold = true

	c.container = container.NewVBox(
		container.NewHBox(layout.NewSpacer(), titleLabel, layout.NewSpacer()),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("文本"),
			c.textInput,
			widget.NewLabel("热键"),
			c.hotkeyInput,
		),
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			cancelBtn,
			c.confirmBtn,
		),
	)

	return c
}
