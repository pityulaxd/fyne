package widget

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
)

type vuRenderer struct {
	label *canvas.Text
	background, bar,
	optimumBar, peakBar,
	lowalphaGreen,
	lowalphaAmber,
	lowalphaRed *canvas.Rectangle

	objects []fyne.CanvasObject
	meter   *vuMeter
}

// MinSize calculates the minimum size of the VU meter.  Code shamelessly stolen from progressbar for now.
func (v *vuRenderer) MinSize() fyne.Size {
	var tsize fyne.Size
	if text := v.meter.TextFormatter; text != nil {
		tsize = fyne.MeasureText(text(), v.label.TextSize, v.label.TextStyle)
	} else {
		tsize = fyne.MeasureText("100%", v.label.TextSize, v.label.TextStyle)
	}

	return fyne.NewSize(tsize.Width+theme.Padding()*4, tsize.Height+theme.Padding()*2)
}

func (v *vuRenderer) updateBars() {
	if v.meter.Value < v.meter.Min {
		v.meter.Value = v.meter.Min
	}
	if v.meter.Value > v.meter.Max {
		v.meter.Value = v.meter.Max
	}

	delta := float64(v.meter.Max - v.meter.Min)
	ratio := float64(v.meter.Value-v.meter.Min) / delta

	if text := v.meter.TextFormatter; text != nil {
		v.label.Text = text()
	} else {
		v.label.Text = strconv.Itoa(int(ratio*100)) + "%"
	}

	size := v.meter.Size()
	var ratioSize float64
	if v.meter.VUMeterDirection == VUMeterHorizontal {
		ratioSize = float64(size.Width)
	} else {
		ratioSize = float64(size.Height)
	}

	greenSize := 0.0
	amberSize := 0.0
	redSize := 0.0
	if ratio > (v.meter.OptimumValueMin / 100) {
		if ratio > (v.meter.OptimumValueMax / 100) {
			greenSize = ratioSize * v.meter.OptimumValueMin / 100
			amberSize = ratioSize * v.meter.OptimumValueMax / 100
			redSize = ratioSize * ratio
		} else {
			greenSize = ratioSize * v.meter.OptimumValueMin / 100
			amberSize = ratioSize * ratio
		}
	} else {
		greenSize = ratioSize * ratio
	}
	redSize = redSize - amberSize
	if redSize < 0 {
		redSize = 0
	}
	amberSize = amberSize - greenSize
	if amberSize < 0 {
		amberSize = 0
	}

	lowalphaGreenSize := ratioSize * v.meter.OptimumValueMin / 100
	lowalphaAmberSize := ratioSize * v.meter.OptimumValueMax / 100
	lowalphaRedSize := ratioSize
	lowalphaRedSize = lowalphaRedSize - lowalphaAmberSize
	lowalphaAmberSize = lowalphaAmberSize - lowalphaGreenSize

	if v.meter.VUMeterDirection == VUMeterHorizontal {
		v.lowalphaGreen.Move(fyne.NewPos(0, 0))
		v.lowalphaGreen.Resize(fyne.NewSize(float32(lowalphaGreenSize), size.Height))
		v.lowalphaAmber.Resize(fyne.NewSize(float32(lowalphaAmberSize), size.Height))
		v.lowalphaAmber.Move(fyne.NewPos(float32(lowalphaGreenSize), 0))
		v.lowalphaRed.Resize(fyne.NewSize(float32(lowalphaRedSize), size.Height))
		v.lowalphaRed.Move(fyne.NewPos(float32(lowalphaAmberSize+lowalphaGreenSize), 0))
	} else {
		v.lowalphaGreen.Move(fyne.NewPos(0, size.Height-float32(lowalphaGreenSize)))
		v.lowalphaGreen.Resize(fyne.NewSize(size.Width, float32(lowalphaGreenSize)))
		v.lowalphaAmber.Resize(fyne.NewSize(size.Width, float32(lowalphaAmberSize)))
		v.lowalphaAmber.Move(fyne.NewPos(0, size.Height-float32(lowalphaGreenSize+lowalphaAmberSize)))
		v.lowalphaRed.Resize(fyne.NewSize(size.Width, float32(lowalphaRedSize)))
		v.lowalphaRed.Move(fyne.NewPos(0, size.Height-float32(lowalphaGreenSize+lowalphaAmberSize+lowalphaRedSize)))
	}

	if v.meter.VUMeterDirection == VUMeterHorizontal {
		v.bar.Move(fyne.NewPos(0, 0))
		v.bar.Resize(fyne.NewSize(float32(greenSize), size.Height))
		v.optimumBar.Resize(fyne.NewSize(float32(amberSize), size.Height))
		v.optimumBar.Move(fyne.NewPos(float32(greenSize), 0))
		v.peakBar.Resize(fyne.NewSize(float32(redSize), size.Height))
		v.peakBar.Move(fyne.NewPos(float32(greenSize+amberSize), 0))
	} else {
		v.bar.Resize(fyne.NewSize(size.Width, float32(greenSize)))
		v.bar.Move(fyne.NewPos(0, size.Height-float32(greenSize)))
		v.optimumBar.Resize(fyne.NewSize(size.Width, float32(amberSize)))
		v.optimumBar.Move(fyne.NewPos(0, size.Height-float32(greenSize+amberSize)))
		v.peakBar.Resize(fyne.NewSize(size.Width, float32(redSize)))
		v.peakBar.Move(fyne.NewPos(0, size.Height-float32(greenSize+amberSize+redSize)))
	}
}

// Layout the components of the widget
func (v *vuRenderer) Layout(size fyne.Size) {
	v.background.Resize(size)
	v.label.Resize(size)
	v.updateBars()
}

// ApplyTheme is called when the vuMeter may need to update it's look
func (v *vuRenderer) ApplyTheme() {
	v.label.Color = theme.ForegroundColor()
	v.Refresh()
}

func (v *vuRenderer) BackgroundColor() color.Color {
	return theme.ButtonColor()
}

func (v *vuRenderer) Refresh() {
	v.label.Text = fmt.Sprintf("%f %%", v.meter.Value)
	v.Layout(v.meter.Size())
	canvas.Refresh(v.meter)
}

func (v *vuRenderer) Objects() []fyne.CanvasObject {
	return v.objects
}

func (v *vuRenderer) Destroy() {
}

type VUMeterDirectionEnum uint32

const (
	VUMeterVertical = iota
	VUMeterHorizontal
)

// vuMeter widget is a kind of custom progressbar but has "zones" of different color for peaking.
type vuMeter struct {
	BaseWidget
	TextFormatter func() string
	Value, Min, Max,
	OptimumValueMin, OptimumValueMax float64
	VUMeterDirection VUMeterDirectionEnum

	binder basicBinder
}

func (m *vuMeter) CreateRenderer() fyne.WidgetRenderer {
	m.ExtendBaseWidget(m)
	if m.Min == 0 && m.Max == 0 {
		m.Max = 1.0
	}

	background := canvas.NewRectangle(theme.BackgroundColor())
	lowalphaGreen := canvas.NewRectangle(color.RGBA{0, 255, 0, 64})
	lowalphaAmber := canvas.NewRectangle(color.RGBA{255, 200, 0, 64})
	lowalphaRed := canvas.NewRectangle(color.RGBA{255, 0, 0, 64})
	bar := canvas.NewRectangle(color.RGBA{0, 255, 0, 255})
	optimumBar := canvas.NewRectangle(color.RGBA{255, 200, 0, 255})
	peakBar := canvas.NewRectangle(color.RGBA{255, 0, 0, 255})
	label := canvas.NewText("0%", theme.ForegroundColor())
	label.Alignment = fyne.TextAlignCenter
	objects := []fyne.CanvasObject{
		background,
		lowalphaGreen,
		lowalphaAmber,
		lowalphaRed,
		bar,
		optimumBar,
		peakBar,
		label,
	}

	return &vuRenderer{label, background, bar, optimumBar, peakBar,
		lowalphaGreen,
		lowalphaAmber,
		lowalphaRed, objects, m}
}

// SetValue changes the current value of this progress bar (from p.Min to p.Max).
// The widget will be refreshed to indicate the change.
func (m *vuMeter) SetValue(v float64) {
	m.Value = v
	m.Refresh()
}

func (m *vuMeter) updateFromData(data binding.DataItem) {
	if data == nil {
		return
	}
	floatSource, ok := data.(binding.Float)
	if !ok {
		return
	}

	val, err := floatSource.Get()
	if err != nil {
		fyne.LogError("Error getting current data value", err)
		return
	}
	m.SetValue(val)
}

func (m *vuMeter) MinSize() fyne.Size {
	m.ExtendBaseWidget(m)
	return m.BaseWidget.MinSize()
}

func (m *vuMeter) Bind(data binding.Float) {
	m.binder.SetCallback(m.updateFromData)
	m.binder.Bind(data)
}

func (m *vuMeter) Unbind() {
	m.binder.Unbind()
}

// NewVUMeter creates a new meter widget with the specified value
func NewVUMeter(value float64) *vuMeter {
	meter := &vuMeter{Value: value}
	meter.OptimumValueMin = 75
	meter.OptimumValueMax = 85
	meter.ExtendBaseWidget(meter)
	meter.VUMeterDirection = VUMeterHorizontal
	return meter
}

func NewVUMeterWithData(data binding.Float) *vuMeter {
	f, err := data.Get()
	if err != nil {
		f = 25.0
	}
	m := NewVUMeter(f)
	m.Bind(data)
	return m
}
