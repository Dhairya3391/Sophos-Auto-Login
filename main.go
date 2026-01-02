package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// Response structures for XML parsing
type LoginResponse struct {
	XMLName xml.Name `xml:"response"`
	Status  string   `xml:"status,attr"`
	Message string   `xml:"message,attr"`
}

type LogoutResponse struct {
	XMLName xml.Name `xml:"response"`
	Status  string   `xml:"status,attr"`
	Message string   `xml:"message,attr"`
}

// Optimized HTTP client with connection pooling
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	},
}

// Config holds application configuration
type Config struct {
	Duration time.Duration
	Forever  bool
}

// LoginPayload represents the login form data
type LoginPayload struct {
	Mode        string `url:"mode"`
	Username    string `url:"username"`
	Password    string `url:"password"`
	A           string `url:"a"`
	ProductType string `url:"producttype"`
}

// LogoutPayload represents the logout form data
type LogoutPayload struct {
	Mode        string `url:"mode"`
	Username    string `url:"username"`
	A           string `url:"a"`
	ProductType string `url:"producttype"`
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode == 200 {
		fmt.Println("Login successful!")
		fmt.Printf("Response: %s\n", string(body))
	} else {
		return fmt.Errorf("login failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read logout response: %w", err)
	}

	if resp.StatusCode == 200 {
		fmt.Println("Logout successful!")
		fmt.Printf("Response: %s\n", string(body))
	} else {
		return fmt.Errorf("logout failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// checkPing performs optimized connectivity check using raw socket
func checkPing() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use a more efficient approach - try to establish a TCP connection
	// to a reliable endpoint instead of relying on external ping command
	dialer := &net.Dialer{
		Timeout: 2 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", "1.1.1.1:53")
	if err != nil {
		return false
	}
	conn.Close()
	return true
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

// setupTerminalInput handles terminal input for quit functionality
func setupTerminalInput(quitChan chan<- struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			fmt.Println("\nInterrupt signal detected. Logging out and exiting...")
			quitChan <- struct{}{}
		}
	}()

	// For terminal input, we'll use a simpler approach
	// that doesn't require raw terminal manipulation
	go func() {
		var input string
		for {
			_, err := fmt.Scanln(&input)
			if err != nil {
				continue // No input available
			}
			if input == "q" || input == "Q" {
				fmt.Println("\nDetected q/Q. Logging out and exiting...")
				quitChan <- struct{}{}
				return
			}
		}
	}()
}

// runMainLoop executes the main application loop
func runMainLoop(config *Config) error {
	quitChan := make(chan struct{}, 1)
	setupTerminalInput(quitChan)

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	statusCheckTicker := time.NewTicker(30 * time.Second)
	defer statusCheckTicker.Stop()

	for {
		select {
		case <-quitChan:
			return logout()
		case <-ticker.C:
			now := time.Now()

			// Check duration limit
			if !config.Forever && config.Duration > 0 {
				elapsed := now.Sub(startTime)
				if elapsed >= config.Duration {
					fmt.Println("\nAuto logout timer expired.")
					return logout()
				}
				remaining := config.Duration - elapsed
				fmt.Printf("Press q or Q to logout and stop. Time left: %d seconds\r", int(remaining.Seconds()))
			} else {
				fmt.Print("Press q or Q to logout and stop. \r")
			}

		case <-statusCheckTicker.C:
			if !checkPing() {
				fmt.Println("\nPing failed. Attempting to re-login...")
				if err := login(); err != nil {
					log.Printf("Re-login failed: %v", err)
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
		fmt.Println("No duration flag given. Defaulting to forever. Press q/Q or Ctrl+C to logout and stop...")
	} else if config.Forever {
		fmt.Println("Running forever. Press q/Q or Ctrl+C to logout and stop...")
	} else if config.Duration > 0 {
		hours := int(config.Duration.Hours())
		minutes := int(config.Duration.Minutes()) % 60
		if hours > 0 {
			fmt.Printf("Running for %d hours", hours)
			if minutes > 0 {
				fmt.Printf(" and %d minutes", minutes)
			}
		} else {
			fmt.Printf("Running for %d minutes", minutes)
		}
		fmt.Println(". Press q/Q or Ctrl+C to logout and stop...")
	}

	// Perform initial login
	if err := login(); err != nil {
		log.Fatalf("Initial login failed: %v", err)
	}

	// Run main loop
	if err := runMainLoop(config); err != nil {
		log.Printf("Error in main loop: %v", err)
		os.Exit(1)
	}
}
