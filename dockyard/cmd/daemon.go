/*
Copyright 2016 - 2017 Huawei Technologies Co., Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/macaron.v1"

	"github.com/Huawei/containerops/common"
	"github.com/Huawei/containerops/dockyard/model"
	"github.com/Huawei/containerops/dockyard/setting"
	"github.com/Huawei/containerops/dockyard/web"
)

var addressOption string
var portOption int

// webCmd is subcommand which start/stop/monitor Dockyard's REST API daemon.
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Web subcommand start/stop/monitor Dockyard's REST API daemon.",
	Long:  ``,
}

// start Dockyard deamon subcommand
var startDaemonCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Dockyard's REST API daemon.",
	Long:  ``,
	Run:   startDeamon,
}

// stop Dockyard deamon subcommand
var stopDaemonCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop Dockyard's REST API daemon.",
	Long:  ``,
	Run:   stopDaemon,
}

// monitor Dockyard deamon subcommand
var monitorDeamonCmd = &cobra.Command{
	Use:   "monitor",
	Short: "monitor Dockyard's REST API daemon.",
	Long:  ``,
	Run:   monitorDaemon,
}

// init()
func init() {
	RootCmd.AddCommand(daemonCmd)

	// Add start subcommand
	daemonCmd.AddCommand(startDaemonCmd)
	startDaemonCmd.Flags().StringVarP(&addressOption, "address", "a", "", "http or https listen address.")
	startDaemonCmd.Flags().IntVarP(&portOption, "port", "p", 0, "the port of http.")
	startDaemonCmd.Flags().StringVarP(&configFilePath, "config", "c", "./conf/runtime.conf", "path of the config file.")

	// Add stop subcommand
	daemonCmd.AddCommand(stopDaemonCmd)
	// Add daemon subcommand
	daemonCmd.AddCommand(monitorDeamonCmd)
}

// startDeamon() start Dockyard's REST API daemon.
func startDeamon(cmd *cobra.Command, args []string) {
	if err := setting.SetConfig(configFilePath); err != nil {
		log.Fatalf("Failed to init settings: %s", err.Error())
		os.Exit(1)
	}

	model.OpenDatabase(&setting.Database)
	m := macaron.New()

	// Set Macaron Web Middleware And Routers
	web.SetDockyardMacaron(m)

	var server *http.Server
	stopChan := make(chan os.Signal)

	signal.Notify(stopChan, os.Interrupt)

	address := setting.ListenMode.Address
	if addressOption != "" {
		address = addressOption
	}
	port := setting.ListenMode.Port
	if portOption != 0 {
		port = portOption
	}
	fmt.Println("listening on", address, port)

	go func() {
		switch setting.ListenMode.Mode {
		case "http":
			listenaddr := fmt.Sprintf("%s:%d", address, port)
			server = &http.Server{Addr: listenaddr, Handler: m}
			if err := server.ListenAndServe(); err != nil {
				fmt.Printf("Start Dockyard http service error: %v\n", err.Error())
			}
			break
		case "https":
			listenaddr := fmt.Sprintf("%s:%d", address, port)
			server = &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
			if err := server.ListenAndServeTLS(setting.ListenMode.Cert, setting.ListenMode.CertKey); err != nil {
				fmt.Printf("Start Dockyard https service error: %v\n", err.Error())
			}
			break
		case "unix":
			listenaddr := fmt.Sprintf("%s", address)
			if common.IsFileExist(listenaddr) {
				os.Remove(listenaddr)
			}

			if listener, err := net.Listen("unix", listenaddr); err != nil {
				fmt.Printf("Start Dockyard unix socket error: %v\n", err.Error())
			} else {
				server = &http.Server{Handler: m}
				if err := server.Serve(listener); err != nil {
					fmt.Printf("Start Dockyard unix socket error: %v\n", err.Error())
				}
			}
			break
		default:
			fmt.Printf("Invalid listenMode: %s\n", setting.ListenMode.Mode)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	<-stopChan // wait for SIGINT
	log.Errorln("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	log.Errorln("Server gracefully stopped")
}

// stopDaemon() stop Dockyard's REST API daemon.
func stopDaemon(cmd *cobra.Command, args []string) {

}

// monitordAemon() monitor Dockyard's REST API deamon.
func monitorDaemon(cmd *cobra.Command, args []string) {

}
