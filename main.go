package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/bushtalkradio/xplane-client/bushtalk"
	"github.com/bushtalkradio/xplane-client/config"
	"github.com/bushtalkradio/xplane-client/ui"
	"github.com/bushtalkradio/xplane-client/xplane"
)

const (
	trackInterval = 5 * time.Second
	reconnectDelay = 5 * time.Second
)

type App struct {
	fyneApp        fyne.App
	cfg            *config.Config
	bushtalkClient *bushtalk.Client
	xplaneClient   *xplane.Client
	loginWindow    *ui.LoginWindow
	statusWindow   *ui.StatusWindow
	stopCh         chan struct{}
}

func main() {
	// Load configuration first (before any UI)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Hide console immediately, before Fyne app starts
	if !cfg.ShowConsole {
		HideConsole()
	}

	fyneApp := app.New()
	fyneApp.SetIcon(AppIcon())
	fyneApp.Settings().SetTheme(&BushtalkTheme{})

	a := &App{
		fyneApp: fyneApp,
		cfg:     cfg,
	}

	// Initialize Bushtalk client
	a.bushtalkClient = bushtalk.NewClient(cfg.ApiURL)

	// Check if we have saved credentials
	if cfg.HasCredentials() {
		a.bushtalkClient.SetToken(cfg.ApiToken)
		a.showStatusWindow()
		a.startTracking()
	} else {
		a.showLoginWindow()
	}

	a.fyneApp.Run()
}

func (a *App) showLoginWindow() {
	a.loginWindow = ui.NewLoginWindow(a.fyneApp, a.cfg, a.bushtalkClient, func(token string) {
		// Login successful
		a.bushtalkClient.SetToken(token)
		a.loginWindow.Hide()
		a.showStatusWindow()
		a.startTracking()
	})
	a.loginWindow.Window().SetOnClosed(func() {
		a.fyneApp.Quit()
	})
	a.loginWindow.Show()
}

func (a *App) showStatusWindow() {
	a.statusWindow = ui.NewStatusWindow(a.fyneApp,
		// onDisconnect - stop tracking but stay logged in
		func() {
			a.stopTracking()
			if a.xplaneClient != nil {
				a.xplaneClient.Disconnect()
			}
			// Restart tracking (will reconnect to X-Plane)
			a.startTracking()
		},
	)
	a.statusWindow.Window().SetOnClosed(func() {
		a.stopTracking()
		if a.xplaneClient != nil {
			a.xplaneClient.Disconnect()
		}
		a.fyneApp.Quit()
	})
	a.statusWindow.Show()
}

func (a *App) startTracking() {
	a.stopCh = make(chan struct{})

	// Connect to X-Plane
	go a.connectXPlane()

	// Start position sending loop
	go a.trackingLoop()
}

func (a *App) stopTracking() {
	if a.stopCh != nil {
		close(a.stopCh)
	}
}

func (a *App) connectXPlane() {
	for {
		select {
		case <-a.stopCh:
			return
		default:
		}

		a.xplaneClient = xplane.NewClient(a.cfg.XPlanePort)
		a.xplaneClient.SetCallbacks(
			func() {
				// Connected
				if a.statusWindow != nil {
					a.statusWindow.SetXPlaneConnected(true)
				}
			},
			func() {
				// Disconnected - will trigger reconnect
				if a.statusWindow != nil {
					a.statusWindow.SetXPlaneConnected(false)
				}
			},
		)

		err := a.xplaneClient.Connect()
		if err != nil {
			log.Printf("X-Plane connection failed: %v, retrying in %v", err, reconnectDelay)
			select {
			case <-a.stopCh:
				return
			case <-time.After(reconnectDelay):
				continue
			}
		}

		// Wait for disconnect or stop
		select {
		case <-a.stopCh:
			a.xplaneClient.Disconnect()
			return
		case <-a.xplaneClient.Done():
			// X-Plane disconnected, reconnect after delay
			log.Printf("X-Plane disconnected, reconnecting in %v", reconnectDelay)
			select {
			case <-a.stopCh:
				return
			case <-time.After(reconnectDelay):
				continue
			}
		}
	}
}

func (a *App) trackingLoop() {
	ticker := time.NewTicker(trackInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.sendPosition()
		}
	}
}

func (a *App) sendPosition() {
	if a.xplaneClient == nil || !a.xplaneClient.IsConnected() {
		return
	}

	pos := a.xplaneClient.GetPosition()
	if !pos.IsValid() {
		return
	}

	// Update status window
	if a.statusWindow != nil {
		a.statusWindow.UpdatePosition(pos)
	}

	// Convert to Bushtalk format
	payload := &bushtalk.TrackPayload{
		Latitude:       pos.Latitude,
		Longitude:      pos.Longitude,
		AltitudeAGL:    pos.AltitudeAGL * 3.28084, // meters to feet
		GroundVelocity: pos.Groundspeed * 1.94384, // m/s to knots
		Heading:        pos.Heading,
		TailNumber:     pos.TailNumber,
		OnGround:       pos.AltitudeAGL < 1.0, // Below 1 meter AGL
	}

	log.Printf("Sending: lat=%.4f lon=%.4f alt=%.0fft spd=%.0fkts hdg=%.0f° tail=%s ground=%v",
		payload.Latitude, payload.Longitude, payload.AltitudeAGL,
		payload.GroundVelocity, payload.Heading, payload.TailNumber, payload.OnGround)

	err := a.bushtalkClient.SendPosition(payload)
	if err != nil {
		log.Printf("Failed to send position: %v", err)
		return
	}

	if a.statusWindow != nil {
		a.statusWindow.SetLastSent(time.Now())
	}
}
