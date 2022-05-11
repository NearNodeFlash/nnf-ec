/* -----------------------------------------------------------------
 * main.go -
 *
 * Provides an entrypoint and server for developing and testing rfsf-openapi.
 *
 * Author: Caleb Carlson
 *
 * © Copyright 2021 Hewlett Packard Enterprise Development LP
 *
 * ----------------------------------------------------------------- */

package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	sw "github.com/nearnodeflash/nnf-ec/pkg/rfsf/pkg/routermux"
)

func isEmulationEnv() bool { return os.Getenv("EMU") == "true" }
func isTestingEnv() bool   { return os.Getenv("TESTING") == "true" }

func main() {

	serverPort := "8080"
	serverHost := ""

	log.Infof("Redfish-Swordfish OpenAPI Server started on port %s", serverPort)

	switch true {
	case isEmulationEnv():
		log.Info("Running in emulation mode")
	case isTestingEnv():
		log.Info("Running in testing mode")
	}

	router := sw.NewRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", serverHost, serverPort), router))
}
