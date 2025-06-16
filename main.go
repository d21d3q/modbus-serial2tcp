package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tbrandon/mbserver"
	"go.bug.st/serial"
)

var (
	// Serial port settings flags
	serialPort   string
	serialSpeed  int
	serialParity string
	timeout      int
	logLevel     int
	tcpServer    string
	unitID       int
)

// validateParity validates that the parity value is one of the allowed values
func validateParity(parity string) error {
	validParities := []string{"N", "E", "O"}

	for _, valid := range validParities {
		if valid == parity {
			return nil
		}
	}

	return fmt.Errorf("invalid parity '%s'. Valid values are: N (none), E (even), O (odd)", parity)
}

// validateTCPServer validates that the TCP server address is in the format ip:port
func validateTCPServer(tcpServer string) error {
	if tcpServer == "" {
		return nil // Empty is allowed (optional flag)
	}

	parts := strings.Split(tcpServer, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid TCP server format '%s'. Expected format: ip:port (e.g., 0.0.0.0:502)", tcpServer)
	}

	ip := parts[0]
	portStr := parts[1]

	// Validate IP address
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address '%s' in TCP server '%s'", ip, tcpServer)
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port '%s' in TCP server '%s'. Port must be a number", portStr, tcpServer)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d in TCP server '%s'. Port must be between 1 and 65535", port, tcpServer)
	}

	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "modbus-serial2tcp",
	Short: "A Modbus Serial to TCP gateway",
	Long: `A Modbus Serial to TCP gateway that allows converting Modbus RTU (serial) 
communication to Modbus TCP for integration with modern systems.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// set log level
		switch logLevel {
		case 1:
			log.SetLevel(log.InfoLevel)
		case 2:
			log.SetLevel(log.DebugLevel)
		case 3:
			log.SetLevel(log.TraceLevel)
		default:
			log.SetLevel(log.ErrorLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Validate parity setting
		if err := validateParity(serialParity); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Validate TCP server setting
		if err := validateTCPServer(tcpServer); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Validate Unit ID
		if unitID < 1 || unitID > 247 {
			fmt.Printf("Error: Unit ID must be between 1 and 247, got %d\n", unitID)
			os.Exit(1)
		}

		fmt.Printf("Modbus Serial to TCP Gateway\n")
		fmt.Printf("Serial Port: %s\n", serialPort)
		fmt.Printf("Speed: %d\n", serialSpeed)
		fmt.Printf("Parity: %s\n", serialParity)
		fmt.Printf("Timeout: %d\n", timeout)
		fmt.Printf("Unit ID: %d\n", unitID)

		// TODO: Implement the actual gateway logic here
		fmt.Println("Gateway would start here...")

		mode := &serial.Mode{
			BaudRate: serialSpeed,
			DataBits: 8,
			StopBits: serial.OneStopBit,
			Parity:   serial.NoParity,
		}

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start the handler in a goroutine
		go handler(mode, serialPort, tcpServer, unitID)

		// Wait for signal
		sig := <-sigChan
		log.Infof("Received signal: %v", sig)
		fmt.Println("\nShutting down gracefully...")
	},
}

func init() {
	rootCmd.PersistentFlags().CountVarP(&logLevel, "verbose", "v", "verbosity level")
	// Define persistent flags for serial port settings
	rootCmd.PersistentFlags().StringVarP(&serialPort, "port", "p", "/dev/ttyUSB0", "Serial port path (e.g., /dev/ttyUSB0, COM1)")
	rootCmd.PersistentFlags().IntVarP(&serialSpeed, "speed", "s", 9600, "Serial port speed/baud rate")
	rootCmd.PersistentFlags().StringVar(&serialParity, "parity", "N", "Serial port parity (N for none, E for even, O for odd)")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 1, "Timeout in seconds")
	rootCmd.PersistentFlags().StringVarP(&tcpServer, "tcp-server", "S", "", "TCP server address (e.g. 0.0.0.0:502)")
	rootCmd.PersistentFlags().IntVarP(&unitID, "unitid", "u", 1, "Modbus Unit ID to filter frames (1-247)")

	// Mark the port flag as required
	rootCmd.MarkPersistentFlagRequired("port")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// parseFrame attempts to parse the accumulated frame buffer and check unit ID
func parseFrame(frameBuffer []byte, expectedUnitID int) {
	if len(frameBuffer) == 0 {
		return
	}

	log.Debugf("Attempting to parse %d bytes: %x", len(frameBuffer), frameBuffer)

	frame, parseErr := mbserver.NewRTUFrame(frameBuffer)
	if parseErr != nil {
		log.Printf("bad serial frame error %v\n", parseErr)
		log.Printf("Keep the RTU server running!!\n")
		return
	}

	// Check if the frame is for our unit ID
	frameUnitID := int(frame.Address)
	if frameUnitID != expectedUnitID {
		log.Debugf("Frame unit ID %d does not match expected unit ID %d, ignoring frame", frameUnitID, expectedUnitID)
		return
	}

	log.Debugf("Frame unit ID %d matches expected unit ID, processing frame", frameUnitID)
	log.Debugf("Successfully parsed RTU frame with %d bytes for unit ID %d", len(frameBuffer), frameUnitID)
	_ = frame // Use the frame for further processing
	// request := &mbserver.Request{port, frame}
}

func handler(mode *serial.Mode, serialPort string, tcpServer string, unitID int) {
	var ser serial.Port = nil
	var err error

	// Frame buffer for accumulating received bytes from potentially fragmented frames
	var frameBuffer []byte

	// Calculate character timeout (time for one character transmission)
	// Character time = (start bit + data bits + parity bit + stop bit) / baud rate
	// For 8N1: 1 + 8 + 0 + 1 = 10 bits per character
	// For 8E1 or 8O1: 1 + 8 + 1 + 1 = 11 bits per character
	bitsPerChar := 10 // Default for 8N1
	if mode.Parity != serial.NoParity {
		bitsPerChar = 11 // 8E1 or 8O1
	}
	charTime := time.Duration(float64(bitsPerChar) / float64(mode.BaudRate) * float64(time.Second))
	frameTimeout := charTime * 4 // 4 times character time

	log.Infof("Character time: %v, Frame timeout: %v", charTime, frameTimeout)

	// Create a timer for frame timeout detection
	frameTimer := time.NewTimer(frameTimeout)
	frameTimer.Stop() // Stop it initially
	log.Infof("Read timeout: 10ms, Frame timeout: %v", frameTimeout)

	for {
		// Try to open the serial port if it's not open
		if ser == nil {
			log.Infof("Attempting to open serial port: %s", serialPort)
			ser, err = serial.Open(serialPort, mode)
			if err != nil {
				log.Errorf("Failed to open serial port %s: %v", serialPort, err)
				log.Infof("Retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
			// Set a very short read timeout to allow frequent frame timeout checking
			ser.SetReadTimeout(10 * time.Millisecond)
			log.Infof("Successfully opened serial port: %s", serialPort)

			// Reset buffer when port is reopened
			frameBuffer = nil
			frameTimer.Stop()
		}

		select {
		case <-frameTimer.C:
			// Frame timeout occurred - parse accumulated data
			if len(frameBuffer) > 0 {
				log.Debugf("Frame timer expired, attempting to parse %d bytes: %x", len(frameBuffer), frameBuffer)
				parseFrame(frameBuffer, unitID)
				frameBuffer = nil
			}

		default:
			// Try to read from the serial port
			readBuffer := make([]byte, 512)
			bytesRead, err := ser.Read(readBuffer)

			// Handle read timeout (not an error for our buffering strategy)
			if err != nil {
				if err.Error() == "timeout" || strings.Contains(err.Error(), "timeout") {
					// Continue to next iteration for timeout
					continue
				}

				// Handle other read errors
				log.Errorf("Error reading from serial port: %v", err)
				// Close the port and set to nil to trigger reconnection
				if ser != nil {
					ser.Close()
					ser = nil
				}
				log.Infof("Closed serial port, will attempt to reconnect...")
				time.Sleep(2 * time.Second)
				continue
			}

			if bytesRead > 0 {
				log.Debugf("Received %d bytes from serial port: %x", bytesRead, readBuffer[:bytesRead])

				// Append new data to frame buffer (frames may arrive in parts)
				frameBuffer = append(frameBuffer, readBuffer[:bytesRead]...)

				// Reset the frame timer on new data
				frameTimer.Stop()
				frameTimer.Reset(frameTimeout)

				log.Debugf("Frame buffer now contains %d bytes: %x", len(frameBuffer), frameBuffer)
			}
		}

		// Small delay to prevent busy waiting
		time.Sleep(1 * time.Millisecond)
	}
}
