package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bushtalkradio/xplane-client/bushtalk"
	"github.com/bushtalkradio/xplane-client/config"
)

// LoginWindow creates the login form window
type LoginWindow struct {
	window    fyne.Window
	cfg       *config.Config
	client    *bushtalk.Client
	onSuccess func(token string)

	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	rememberCheck *widget.Check
	portEntry     *widget.Entry
	apiURLEntry   *widget.Entry
	consoleCheck  *widget.Check
	loginButton   *widget.Button
	statusLabel   *widget.Label
}

// NewLoginWindow creates a new login window
func NewLoginWindow(app fyne.App, cfg *config.Config, client *bushtalk.Client, onSuccess func(token string)) *LoginWindow {
	l := &LoginWindow{
		window:    app.NewWindow("Bushtalk Radio"),
		cfg:       cfg,
		client:    client,
		onSuccess: onSuccess,
	}
	l.buildUI()
	return l
}

func (l *LoginWindow) buildUI() {
	// Header with icon and title
	icon := canvas.NewImageFromResource(l.window.Icon())
	icon.SetMinSize(fyne.NewSize(64, 64))
	icon.FillMode = canvas.ImageFillContain

	title := widget.NewLabelWithStyle("Bushtalk Radio",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true})

	subtitle := widget.NewLabelWithStyle("X-Plane Flight Tracker",
		fyne.TextAlignCenter,
		fyne.TextStyle{})

	header := container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(title),
		container.NewCenter(subtitle),
	)

	// Form fields with better styling
	l.usernameEntry = widget.NewEntry()
	l.usernameEntry.SetPlaceHolder("Username")
	if l.cfg.Username != "" {
		l.usernameEntry.SetText(l.cfg.Username)
	}

	l.passwordEntry = widget.NewPasswordEntry()
	l.passwordEntry.SetPlaceHolder("Password")
	l.passwordEntry.OnSubmitted = func(_ string) {
		l.handleLogin()
	}

	l.rememberCheck = widget.NewCheck("Remember me", nil)
	l.rememberCheck.SetChecked(l.cfg.HasCredentials())

	l.loginButton = widget.NewButtonWithIcon("Connect", theme.LoginIcon(), l.handleLogin)
	l.loginButton.Importance = widget.HighImportance

	l.statusLabel = widget.NewLabel("")
	l.statusLabel.Wrapping = fyne.TextWrapWord
	l.statusLabel.Alignment = fyne.TextAlignCenter

	// Credentials form
	credentialsCard := widget.NewCard("", "", container.NewVBox(
		l.usernameEntry,
		l.passwordEntry,
		l.rememberCheck,
	))

	// Advanced settings (collapsed by default)
	l.portEntry = widget.NewEntry()
	l.portEntry.SetPlaceHolder("8086")
	l.portEntry.SetText(fmt.Sprintf("%d", l.cfg.XPlanePort))

	l.apiURLEntry = widget.NewEntry()
	l.apiURLEntry.SetPlaceHolder("https://bushtalkradio.com")
	l.apiURLEntry.SetText(l.cfg.ApiURL)

	l.consoleCheck = widget.NewCheck("Show debug console (requires restart)", nil)
	l.consoleCheck.SetChecked(l.cfg.ShowConsole)

	advancedContent := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("X-Plane Port", l.portEntry),
			widget.NewFormItem("API URL", l.apiURLEntry),
		),
		l.consoleCheck,
	)

	advancedAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Advanced Settings", advancedContent),
	)

	// Main layout with proper spacing
	content := container.NewVBox(
		layout.NewSpacer(),
		header,
		widget.NewSeparator(),
		credentialsCard,
		l.loginButton,
		l.statusLabel,
		layout.NewSpacer(),
		advancedAccordion,
	)

	// Add padding
	padded := container.NewPadded(content)

	l.window.SetContent(padded)
	l.window.Resize(fyne.NewSize(380, 480))
	l.window.CenterOnScreen()
	l.window.SetFixedSize(true)
}

func (l *LoginWindow) handleLogin() {
	username := l.usernameEntry.Text
	password := l.passwordEntry.Text

	if username == "" || password == "" {
		l.statusLabel.SetText("Please enter username and password")
		return
	}

	// Parse and update port
	var port int
	if _, err := fmt.Sscanf(l.portEntry.Text, "%d", &port); err != nil || port <= 0 {
		l.statusLabel.SetText("Invalid X-Plane port")
		return
	}
	l.cfg.XPlanePort = port
	l.cfg.ApiURL = l.apiURLEntry.Text
	l.cfg.ShowConsole = l.consoleCheck.Checked

	// Update client base URL if changed
	l.client = bushtalk.NewClient(l.cfg.ApiURL)

	l.loginButton.Disable()
	l.statusLabel.SetText("Connecting...")

	go func() {
		authResp, err := l.client.Authenticate(username, password)

		if err != nil {
			l.loginButton.Enable()
			l.statusLabel.SetText("Login failed: " + err.Error())
			return
		}

		// Save credentials if remember is checked
		if l.rememberCheck.Checked {
			l.cfg.Username = username
			l.cfg.ApiToken = authResp.IDToken
			if err := l.cfg.Save(); err != nil {
				dialog.ShowError(err, l.window)
			}
		}

		l.onSuccess(authResp.IDToken)
	}()
}

// Show displays the login window
func (l *LoginWindow) Show() {
	l.window.Show()
}

// Hide hides the login window
func (l *LoginWindow) Hide() {
	l.window.Hide()
}

// Close closes the login window
func (l *LoginWindow) Close() {
	l.window.Close()
}

// Window returns the underlying Fyne window
func (l *LoginWindow) Window() fyne.Window {
	return l.window
}
