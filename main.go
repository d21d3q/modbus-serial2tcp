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
	slaveID      int
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

	// Validate IP address or hostname
	if net.ParseIP(ip) == nil && ip != "localhost" {
		return fmt.Errorf("invalid IP address or hostname '%s' in TCP server '%s'", ip, tcpServer)
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
	Use:     "modbus-serial2tcp",
	Version: "0.2.0",
	Short:   "A Modbus Serial to TCP gateway",
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

		// Validate Slave ID
		if slaveID < 1 || slaveID > 247 {
			fmt.Printf("Error: Slave ID must be between 1 and 247, got %d\n", slaveID)
			os.Exit(1)
		}

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
		go handler(mode, serialPort, tcpServer, slaveID, timeout)

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
	rootCmd.PersistentFlags().StringVarP(&serialParity, "parity", "P", "N", "Serial port parity (N for none, E for even, O for odd)")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 1, "Timeout in seconds")
	rootCmd.PersistentFlags().StringVarP(&tcpServer, "tcp-server", "S", "", "TCP server address (e.g. 0.0.0.0:502)")
	rootCmd.PersistentFlags().IntVarP(&slaveID, "slaveid", "i", 1, "Modbus Slave ID to filter frames (1-247)")

	// Mark the port flag as required
	rootCmd.MarkPersistentFlagRequired("port")
	rootCmd.MarkPersistentFlagRequired("speed")
	rootCmd.MarkPersistentFlagRequired("tcp-server")
	rootCmd.MarkPersistentFlagRequired("slaveid")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// parseFrame attempts to parse the accumulated frame buffer and check slave ID
func parseFrame(frameBuffer []byte, expectedSlaveID int) (*mbserver.RTUFrame, error) {
	if len(frameBuffer) == 0 {
		return nil, fmt.Errorf("empty frame buffer")
	}

	log.Debugf("Attempting to parse %d bytes: %x", len(frameBuffer), frameBuffer)

	frame, parseErr := mbserver.NewRTUFrame(frameBuffer)
	if parseErr != nil {
		log.Printf("bad serial frame error %v\n", parseErr)
		log.Printf("Keep the RTU server running!!\n")
		return nil, parseErr
	}

	// Check if the frame is for our slave ID
	frameSlaveID := int(frame.Address)
	if frameSlaveID != expectedSlaveID {
		log.Debugf("Frame slave ID %d does not match expected slave ID %d, ignoring frame", frameSlaveID, expectedSlaveID)
		return nil, fmt.Errorf("frame slave ID %d does not match expected slave ID %d", frameSlaveID, expectedSlaveID)
	}

	log.Debugf("Frame slave ID %d matches expected slave ID, processing frame", frameSlaveID)
	log.Debugf("Successfully parsed RTU frame with %d bytes for slave ID %d", len(frameBuffer), frameSlaveID)
	return frame, nil
}

// makeTCPRequest sends a Modbus RTU frame to a TCP server and returns the response
func makeTCPRequest(tcpServer string, frame *mbserver.RTUFrame, timeoutSeconds int) (*mbserver.RTUFrame, *mbserver.Exception) {
	log.Debugf("Connecting to TCP server: %s", tcpServer)

	// Connect to TCP server
	timeoutDuration := time.Duration(timeoutSeconds) * time.Second
	conn, err := net.DialTimeout("tcp", tcpServer, timeoutDuration)
	if err != nil {
		return nil, &mbserver.GatewayPathUnavailable
	}
	defer conn.Close()

	// create the TCP transport
	transport := newTCPTransport(conn, timeoutDuration)
	defer transport.Close()

	// Convert RTU frame to PDU for TCP transport
	pdu := &pdu{
		unitId:       frame.Address,
		functionCode: frame.Function,
		payload:      frame.Data,
	}

	log.Debugf("Sending Modbus request: Unit=%d, Function=%d, Data=%x", pdu.unitId, pdu.functionCode, pdu.payload)

	// Execute request through TCP transport
	response, err := transport.ExecuteRequest(pdu)
	if err != nil {
		return nil, &mbserver.GatewayTargetDeviceFailedtoRespond
	}

	log.Debugf("Received Modbus response: Unit=%d, Function=%d, Data=%x", response.unitId, response.functionCode, response.payload)

	// For now, let's return the parsed frame as-is since we need a CRC function
	// TODO: Calculate and append CRC for RTU frame
	responseFrame := &mbserver.RTUFrame{
		Address:  response.unitId,
		Function: response.functionCode,
		Data:     response.payload,
	}

	return responseFrame, &mbserver.Success
}

func handler(mode *serial.Mode, serialPort string, tcpServer string, slaveID int, timeoutSeconds int) {
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
	frameTimeout := charTime * 8 // 4 // 4 times character time

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
				frame, err := parseFrame(frameBuffer, slaveID)
				if err == nil && frame != nil {
					// Send frame to TCP server
					tcpRet, exc := makeTCPRequest(tcpServer, frame, timeoutSeconds)
					log.Debug("TCP response:", tcpRet, "Exception:", exc)
					if exc != &mbserver.Success {
						log.Errorf("Failed to send frame to TCP server: %v", err)
						frame.SetException(exc)
						ser.Write(frame.Bytes())
						continue
					}
					// write to serial port
					ret := tcpRet.Bytes()
					n, err := ser.Write(ret)
					if err != nil {
						log.Errorf("Failed to write frame to serial port: %v", err)
					} else {
						log.Debugf("Wrote %d bytes to serial port: %x", n, ret)
					}
				}
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
