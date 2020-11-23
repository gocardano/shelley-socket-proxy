package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/gocardano/go-cardano-client/multiplex"

	log "github.com/sirupsen/logrus"
)

var version string = "-"

const (
	borderRequest  = "\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> R E Q U E S T   # %5d >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n"
	borderResponse = "\n<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< R E S P O N S E   # %5d <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n"

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
			log.WithError(err).Error("Error accepting the unix socket connection")
			os.Exit(1)
		}

		// Connect to destination socket
		destinationConn, err := connectDestinationSocket(*destinationSocketFile)
		if err != nil {
			log.WithError(err).Error("Unable to connect to destination socket")
			os.Exit(1)
		}

		go func(sourceConn, destinationConn net.Conn) {
			go proxy(sourceConn, destinationConn, "src->dst", borderRequest)
			go proxy(destinationConn, sourceConn, "dst->src", borderResponse)
		}(sourceConn, destinationConn)
	}
}

func proxy(source, dest net.Conn, mode, border string) {

	data := []byte{}

	for {
		buf := make([]byte, receivePacketSize)
		read, err1 := source.Read(buf)
		if err1 != nil {
			log.WithError(err1).Error("error reading ", mode)
			break
		}

		_, err2 := dest.Write(buf[:read])
		if err2 != nil {
			log.WithError(err2).Error("error writing ", mode)
		}

		data = append(data, buf[:read]...)
	}

	sdus, err := multiplex.ParseServiceDataUnits(data)
	if err != nil {
		log.WithError(err).Error("Error parsing SDUs")
		return
	}

	for id, sdu := range sdus {
		fmt.Printf(border, id)
		fmt.Println(sdu.Debug())
	}

	log.Debug("Retiring thread for ", mode)
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
