package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bushtalkradio/xplane-client/xplane"
)

// InfoRow holds a label-value pair for display
type InfoRow struct {
	Container *fyne.Container
	Value     *widget.Label
}

// StatusWindow shows connection status and position info
type StatusWindow struct {
	window       fyne.Window
	onDisconnect func()
	onSignOut    func()

	connectionDot *canvas.Circle
	xplaneStatus  *widget.Label
	tailRow       *InfoRow
	positionRow   *InfoRow
	altitudeRow   *InfoRow
	speedRow      *InfoRow
	headingRow    *InfoRow
	lastSentRow   *InfoRow
	disconnectBtn *widget.Button
	signOutBtn    *widget.Button

	stopUpdate chan struct{}
}

// Status colors
var (
	colorConnected    = color.NRGBA{R: 52, G: 211, B: 153, A: 255}  // green
	colorDisconnected = color.NRGBA{R: 248, G: 113, B: 113, A: 255} // red
	colorPending      = color.NRGBA{R: 251, G: 146, B: 60, A: 255}  // orange
)

// NewStatusWindow creates a new status window
func NewStatusWindow(app fyne.App, onDisconnect func(), onSignOut func()) *StatusWindow {
	s := &StatusWindow{
		window:       app.NewWindow("Bushtalk Radio"),
		onDisconnect: onDisconnect,
		onSignOut:    onSignOut,
		stopUpdate:   make(chan struct{}),
	}
	s.buildUI()
	return s
}

func (s *StatusWindow) buildUI() {
	// Header with logo
	icon := canvas.NewImageFromResource(s.window.Icon())
	icon.SetMinSize(fyne.NewSize(48, 48))
	icon.FillMode = canvas.ImageFillContain

	title := widget.NewLabelWithStyle("Bushtalk Radio",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true})

	// Connection status with colored dot
	s.connectionDot = canvas.NewCircle(colorPending)
	s.connectionDot.StrokeWidth = 0

	s.xplaneStatus = widget.NewLabel("Connecting to X-Plane...")
	s.xplaneStatus.TextStyle = fyne.TextStyle{Italic: true}

	// Fixed size dot container using a spacer rectangle
	dotSpacer := canvas.NewRectangle(color.Transparent)
	dotSpacer.SetMinSize(fyne.NewSize(14, 14))

	statusRow := container.NewHBox(
		container.NewStack(dotSpacer, container.NewCenter(s.connectionDot)),
		s.xplaneStatus,
	)

	header := container.NewHBox(
		icon,
		container.NewVBox(
			title,
			statusRow,
		),
	)

	// Aircraft info card
	s.tailRow = createInfoRow("Aircraft", "--")
	s.positionRow = createInfoRow("Position", "--")
	s.altitudeRow = createInfoRow("Altitude", "--")
	s.speedRow = createInfoRow("Speed", "--")
	s.headingRow = createInfoRow("Heading", "--")
	s.lastSentRow = createInfoRow("Last Update", "--")

	flightCard := widget.NewCard("Flight Data", "", container.NewVBox(
		s.tailRow.Container,
		widget.NewSeparator(),
		s.positionRow.Container,
		s.altitudeRow.Container,
		s.speedRow.Container,
		s.headingRow.Container,
		widget.NewSeparator(),
		s.lastSentRow.Container,
	))

	// Buttons
	s.disconnectBtn = widget.NewButtonWithIcon("Disconnect", theme.MediaStopIcon(), func() {
		if s.onDisconnect != nil {
			s.onDisconnect()
		}
	})

	s.signOutBtn = widget.NewButtonWithIcon("Sign Out", theme.LogoutIcon(), func() {
		if s.onSignOut != nil {
			s.onSignOut()
		}
	})

	buttonRow := container.NewGridWithColumns(2, s.disconnectBtn, s.signOutBtn)

	// Info text with links
	audioNote := widget.NewRichTextFromMarkdown(
		"Log in at [bushtalkradio.com](https://bushtalkradio.com) to hear audio.")
	discordNote := widget.NewRichTextFromMarkdown(
		"Need help? [Join our Discord](https://discord.gg/ZcGgw9mUqA)")

	// Main layout
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		flightCard,
		layout.NewSpacer(),
		audioNote,
		discordNote,
		layout.NewSpacer(),
		buttonRow,
	)

	padded := container.NewPadded(content)
	s.window.SetContent(padded)
	s.window.Resize(fyne.NewSize(360, 450))
	s.window.CenterOnScreen()
}

// createInfoRow creates a label-value row for flight data display
func createInfoRow(label, value string) *InfoRow {
	labelWidget := widget.NewLabelWithStyle(label,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true})

	valueWidget := widget.NewLabel(value)
	valueWidget.Alignment = fyne.TextAlignTrailing

	return &InfoRow{
		Container: container.NewHBox(
			labelWidget,
			layout.NewSpacer(),
			valueWidget,
		),
		Value: valueWidget,
	}
}

// SetXPlaneConnected updates the X-Plane connection status
func (s *StatusWindow) SetXPlaneConnected(connected bool) {
	if connected {
		s.connectionDot.FillColor = colorConnected
		s.xplaneStatus.SetText("Connected to X-Plane")
	} else {
		s.connectionDot.FillColor = colorDisconnected
		s.xplaneStatus.SetText("Disconnected from X-Plane")
	}
	s.connectionDot.Refresh()
}

// UpdatePosition updates the displayed position info
func (s *StatusWindow) UpdatePosition(pos xplane.Position) {
	// Update tail number
	if pos.TailNumber != "" && pos.TailNumber != "UNKNOWN" {
		s.tailRow.Value.SetText(pos.TailNumber)
	} else {
		s.tailRow.Value.SetText("--")
	}

	// Update position
	if pos.IsValid() {
		s.positionRow.Value.SetText(fmt.Sprintf("%.4f°, %.4f°", pos.Latitude, pos.Longitude))
		s.altitudeRow.Value.SetText(fmt.Sprintf("%.0f ft AGL", pos.AltitudeAGL*3.28084))
		s.speedRow.Value.SetText(fmt.Sprintf("%.0f kts", pos.Groundspeed*1.94384))
		s.headingRow.Value.SetText(fmt.Sprintf("%.0f°", pos.Heading))
	} else {
		s.positionRow.Value.SetText("--")
		s.altitudeRow.Value.SetText("--")
		s.speedRow.Value.SetText("--")
		s.headingRow.Value.SetText("--")
	}
}

// SetLastSent updates the last sent timestamp
func (s *StatusWindow) SetLastSent(t time.Time) {
	s.lastSentRow.Value.SetText(t.Format("15:04:05"))
}

// Show displays the status window
func (s *StatusWindow) Show() {
	s.window.Show()
}

// Hide hides the status window
func (s *StatusWindow) Hide() {
	s.window.Hide()
}

// Close closes the status window
func (s *StatusWindow) Close() {
	close(s.stopUpdate)
	s.window.Close()
}

// Window returns the underlying Fyne window
func (s *StatusWindow) Window() fyne.Window {
	return s.window
}
