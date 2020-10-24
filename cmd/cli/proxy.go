package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/gocardano/go-cardano-client/multiplex"

	log "github.com/sirupsen/logrus"
)

var version string = "-"

const (
	borderRequest  = ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> R E Q U E S T >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
	borderResponse = "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< R E S P O N S E <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"

	networkUnix       = "unix"
	receivePacketSize = 8192
)

func main() {

	sourceSocketFile := flag.String("source-socket-file", "", "Source socket filename to read from")
	destinationSocketFile := flag.String("destination-socket-file", "", "Socket filename to write to")
	debug := flag.Bool("debug", false, "Enable debug/trace level logging")
	showVersion := flag.Bool("version", false, "Display version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("shelley-socket-proxy version: %s\n", version)
		os.Exit(0)
	}

	if *sourceSocketFile == "" || *destinationSocketFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.SetLevel(log.ErrorLevel)
	if *debug {
		log.SetLevel(log.TraceLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(false)

	log.Infof("Starting application version: %s", version)

	// Delete source socket file if it exists
	if _, err := os.Stat(*sourceSocketFile); err == nil {
		if err = os.Remove(*sourceSocketFile); err != nil {
			log.WithError(err).WithField("sourceSocketFile", *sourceSocketFile).Error("Error deleting read socket file")
		}
	}

	// Create read socket file
	listener, err := net.ListenUnix(networkUnix, &net.UnixAddr{Name: *sourceSocketFile, Net: networkUnix})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"sourceSocketFile": *sourceSocketFile,
		}).Error("Error creating read socket")
		os.Exit(1)
	}

	for {
		// Accept the unix connection
		sourceConn, err := listener.AcceptUnix()
		if err != nil {
			log.WithError(err).Error("Error accepting the unix connection")
			os.Exit(1)
		}

		// Connect to destination socket
		destinationConn, err := connectDestinationSocket(*destinationSocketFile)
		if err != nil {
			log.Error("Unable to connect to write socket")
			os.Exit(1)
		}

		go func(sourceConn, destinationConn net.Conn) {
			for {
				// send request to cardano-node
				err = relayMessage(sourceConn, destinationConn, borderRequest)
				if err != nil {
					if err != io.EOF {
						log.WithError(err).Error("Error relaying message from source to destination")
					}
					break
				}

				// send response from cardano-node
				err = relayMessage(destinationConn, sourceConn, borderResponse)
				if err != nil {
					if err != io.EOF {
						log.WithError(err).Error("Error relaying message from source to destination")
					}
					break
				}
			}
		}(sourceConn, destinationConn)
	}
}

// connectDestinationSocket returns a connection to the destination unix socket, after doing some checks
func connectDestinationSocket(filename string) (net.Conn, error) {

	// Check if the write socket file exists
	writeSocket, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		log.WithError(err).WithField("filename", filename).Error("File does not exists")
	} else if err != nil {
		log.WithError(err).WithField("filename", filename).Error("Unknown error")
	} else if writeSocket.IsDir() {
		log.WithField("filename", filename).Error("Socket is a directory, expecting a unix file socket to cardano-node", filename)
	}

	return net.Dial(networkUnix, filename)
}

// debugShelleyContainer prints out the shelley container
func debugShelleyContainer(data []byte, mode string) error {
	container, err := multiplex.ParseContainer(data)
	if err != nil {
		return err
	}

	fmt.Println("\n\n" + mode)
	fmt.Println(container.Debug())
	return nil
}

// relayMessage sends the message from the source to destination
func relayMessage(source, destination net.Conn, mode string) error {

	buf := make([]byte, receivePacketSize)

	// Read from socket
	read, err := source.Read(buf[:])
	if err != nil {
		if err != io.EOF {
			log.WithError(err).Error("Error reading from source socket")
		}
		return err
	}
	log.Debugf("Successfuly read [%d] bytes from the source", read)

	// Decode and print out
	if err = debugShelleyContainer(buf[:read], mode); err != nil {
		log.WithError(err).Error("Error parsing shelly container")
		return err
	}

	// Write to socket
	wrote, err := destination.Write(buf[:read])
	if err != nil {
		log.WithError(err).Error("Error writing to destination socket")
		return err
	}
	log.Debugf("Successfuly sent [%d] bytes to the destination", wrote)
	return nil
}
