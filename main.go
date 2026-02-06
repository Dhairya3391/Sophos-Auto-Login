package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ANSI color codes for cleaner output
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
)

// Configuration constants
const (
	BaseURL   = "http://10.255.1.1:8090/"
	LoginURL  = BaseURL + "login.xml"
	LogoutURL = BaseURL + "logout.xml"
	Username  = "24010101602"
	Password  = "dhairya3391"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

// HTTP client with connection pooling
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
	},
}

// Config holds application configuration
type Config struct {
	Duration time.Duration
	Forever  bool
}

// formatDuration formats a duration as human-readable string
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// login performs captive portal login with optimized HTTP client
func login() error {
	payload := url.Values{
		"mode":        {"191"},
		"username":    {Username},
		"password":    {Password},
		"a":           {"1"},
		"producttype": {"0"},
	}

	req, err := http.NewRequest("POST", LoginURL, bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()
	// Drain the response body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("%s✓%s Login successful\n", colorGreen, colorReset)
		return nil
	}
	return fmt.Errorf("login failed (status %d)", resp.StatusCode)
}

// logout performs captive portal logout
func logout() error {
	payload := url.Values{
		"mode":        {"193"},
		"username":    {Username},
		"a":           {"1"},
		"producttype": {"0"},
	}

	req, err := http.NewRequest("POST", LogoutURL, bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer resp.Body.Close()
	// Drain the response body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("%s✓%s Logout successful\n", colorYellow, colorReset)
		return nil
	}
	return fmt.Errorf("logout failed (status %d)", resp.StatusCode)
}

// checkInternet verifies actual internet connectivity (not just captive portal reachability)
func checkInternet() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", "http://www.gstatic.com/generate_204", nil)
	if err != nil {
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 204
}

// parseFlags parses command line arguments
func parseFlags() *Config {
	config := &Config{}

	var minutes int
	var hours int

	flag.BoolVar(&config.Forever, "forever", false, "Run forever until user quits")
	flag.IntVar(&minutes, "minutes", 0, "Run for specified minutes")
	flag.IntVar(&hours, "hours", 0, "Run for specified hours")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Captive Portal Auto Login Script\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Convert duration flags properly
	if minutes > 0 {
		config.Duration = time.Duration(minutes) * time.Minute
	}
	if hours > 0 {
		config.Duration = time.Duration(hours) * time.Hour
	}

	return config
}

// setupSignalHandler handles graceful shutdown on interrupt signals
func setupSignalHandler(quitChan chan<- struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("\n%s!%s Interrupt received. Logging out...\n", colorRed, colorReset)
		quitChan <- struct{}{}
	}()
}

// runMainLoop executes the main application loop
func runMainLoop(config *Config) error {
	quitChan := make(chan struct{}, 1)
	setupSignalHandler(quitChan)

	// Cleanup on exit
	defer fmt.Print(colorReset)
	defer httpClient.CloseIdleConnections()

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	statusCheckTicker := time.NewTicker(30 * time.Second)
	defer statusCheckTicker.Stop()

	for {
		select {
		case <-quitChan:
			fmt.Println()
			return logout()
		case <-ticker.C:
			now := time.Now()
			elapsed := now.Sub(startTime).Round(time.Second)

			// Check duration limit
			if !config.Forever && config.Duration > 0 {
				if elapsed >= config.Duration {
					fmt.Printf("\n%s⏱%s Timer expired. Logging out...\n", colorYellow, colorReset)
					return logout()
				}
				remaining := config.Duration - elapsed
				fmt.Printf("\033[2K\r%s●%s Connected for %s | Time left: %s | Ctrl+C to logout",
					colorGreen, colorGray, formatDuration(elapsed), formatDuration(remaining))
			} else {
				fmt.Printf("\033[2K\r%s●%s Connected for %s | Ctrl+C to logout",
					colorGreen, colorGray, formatDuration(elapsed))
			}

		case <-statusCheckTicker.C:
			if !checkInternet() {
				fmt.Printf("\n%s!%s Connection lost. Reconnecting...\n", colorRed, colorReset)
				if err := login(); err != nil {
					fmt.Printf("%s✗%s Re-login failed: %v\n", colorRed, colorReset, err)
				}
			}
		}
	}
}

func main() {
	config := parseFlags()

	// Set default to forever if no flags provided
	if !config.Forever && config.Duration == 0 {
		config.Forever = true
		fmt.Printf("%sℹ%s No duration flag given. Defaulting to forever.\n", colorGray, colorReset)
	} else if config.Forever {
		fmt.Printf("%sℹ%s Running forever.\n", colorGray, colorReset)
	} else if config.Duration > 0 {
		hours := int(config.Duration.Hours())
		minutes := int(config.Duration.Minutes()) % 60
		if hours > 0 {
			fmt.Printf("%sℹ%s Running for %d hours", colorGray, colorReset, hours)
			if minutes > 0 {
				fmt.Printf(" and %d minutes", minutes)
			}
		} else {
			fmt.Printf("%sℹ%s Running for %d minutes", colorGray, colorReset, minutes)
		}
		fmt.Println()
	}

	// Perform initial login with retry
	maxRetries := 3
	var loginErr error
	for i := 0; i < maxRetries; i++ {
		if loginErr = login(); loginErr == nil {
			break
		}
		if i < maxRetries-1 {
			delay := time.Duration(1<<i) * time.Second
			fmt.Printf("%s⚠%s Retrying in %v...\n", colorYellow, colorReset, delay)
			time.Sleep(delay)
		}
	}
	if loginErr != nil {
		log.Fatalf("Initial login failed after %d attempts: %v", maxRetries, loginErr)
	}

	// Run main loop
	if err := runMainLoop(config); err != nil {
		log.Printf("Error in main loop: %v", err)
		os.Exit(1)
	}
}
