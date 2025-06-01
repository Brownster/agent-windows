// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

//go:generate go run github.com/tc-hib/go-winres@v0.3.3 make --product-version=git-tag --file-version=git-tag --arch=amd64,arm64

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/version"
	"github.com/Brownster/agent-windows/internal/config"
	"github.com/Brownster/agent-windows/internal/log"
	"github.com/Brownster/agent-windows/internal/log/flag"
	"github.com/Brownster/agent-windows/internal/utils"
	"github.com/Brownster/agent-windows/pkg/collector"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type PushConfig struct {
	URL      string
	Username string
	Password string
	Interval time.Duration
	AgentID  string
	JobName  string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	exitCode := run(ctx, os.Args[1:])

	stop()

	// If we are running as a service, we need to signal the service control manager that we are done.
	if !IsService {
		os.Exit(exitCode)
	}

	exitCodeCh <- exitCode

	// Wait for the service control manager to signal that we are done.
	<-serviceManagerFinishedCh
}

func run(ctx context.Context, args []string) int {
	startTime := time.Now()

	app := kingpin.New("windows_agent_collector", "A lightweight Windows metrics collector that pushes to Prometheus Push Gateway.")
	
	// Disable termination for tests
	app.Terminate(nil)
	
	// Service management commands
	installCmd := app.Command("install", "Install as Windows service")
	uninstallCmd := app.Command("uninstall", "Remove Windows service")

	var (
		// Push Gateway Configuration
		pushGatewayURL = app.Flag(
			"push.gateway-url",
			"Prometheus Push Gateway URL",
		).String()

		pushUsername = app.Flag(
			"push.username",
			"Basic auth username for push gateway",
		).String()

		pushPassword = app.Flag(
			"push.password",
			"Basic auth password for push gateway",
		).String()

		pushInterval = app.Flag(
			"push.interval",
			"Interval for pushing metrics to gateway",
		).Default("30s").Duration()

		pushJobName = app.Flag(
			"push.job-name",
			"Job name for push gateway",
		).Default("windows_agent").String()

		// Agent Configuration
		agentID = app.Flag(
			"agent-id",
			"Agent identifier for correlation with WebRTC stats",
		).String()

		enabledCollectors = app.Flag(
			"collectors.enabled",
			"Comma-separated list of collectors to use. Available: cpu,memory,net,pagefile",
		).Default("cpu,memory,net,pagefile").String()

		processPriority = app.Flag(
			"process.priority",
			"Priority of the agent process. Can be one of [\"realtime\", \"high\", \"abovenormal\", \"normal\", \"belownormal\", \"low\"]",
		).Default("normal").String()

		memoryLimit = app.Flag(
			"process.memory-limit",
			"Limit memory usage in bytes. 0 means no limit.",
		).Default("50000000").Int64()
	)

	logFile := &log.AllowedFile{}

	_ = logFile.Set("stdout")
	if IsService {
		_ = logFile.Set("eventlog")
	}

	logConfig := &log.Config{File: logFile}
	flag.AddFlags(app, logConfig)

	app.Version(version.Print("windows_agent_collector"))
	app.HelpFlag.Short('h')

	// Initialize collectors before loading and parsing CLI arguments
	collectors := collector.NewWithFlags(app)

	// Parse configuration and command line arguments
	configFilePath := config.ParseConfigFile(args)
	if configFilePath != "" {
		resolver, err := config.NewConfigFileResolver(configFilePath)
		if err != nil {
			//nolint:sloglint // we do not have a logger yet
			slog.LogAttrs(ctx, slog.LevelError, "Failed to load configuration file",
				slog.Any("err", err),
			)
			return 1
		}

		if err = resolver.Bind(app, args); err != nil {
			//nolint:sloglint // we do not have a logger yet
			slog.LogAttrs(ctx, slog.LevelError, "Failed to bind configuration",
				slog.Any("err", err),
			)
			return 1
		}
	}

	// Parse command line arguments to get the selected command
	// Handle the case where no command is specified (normal operation)
	var parsedCommand string
	var err error
	
	// Check if any service commands are specified
	hasServiceCommand := false
	for _, arg := range args {
		if arg == "install" || arg == "uninstall" || arg == "help" {
			hasServiceCommand = true
			break
		}
	}
	
	if hasServiceCommand {
		parsedCommand, err = app.Parse(args)
		if err != nil {
			//nolint:sloglint // we do not have a logger yet
			slog.LogAttrs(ctx, slog.LevelError, "Failed to parse flags",
				slog.Any("err", err),
			)
			return 1
		}
	} else {
		// No service command, parse flags only
		_, err = app.Parse(args)
		if err != nil {
			//nolint:sloglint // we do not have a logger yet
			slog.LogAttrs(ctx, slog.LevelError, "Failed to parse flags",
				slog.Any("err", err),
			)
			return 1
		}
		parsedCommand = "" // No command for normal operation
	}

	// Handle service management commands
	switch parsedCommand {
	case installCmd.FullCommand():
		return handleServiceInstall(ctx, args)
	case uninstallCmd.FullCommand():
		return handleServiceUninstall(ctx)
	case "":
		// No command specified - continue with normal operation
		break
	default:
		// Unknown command - continue with normal operation
		break
	}

	// Validate required flags for normal operation
	if *pushGatewayURL == "" {
		fmt.Println("Error: --push.gateway-url is required")
		fmt.Println("Use --help for usage information")
		return 1
	}
	if *agentID == "" {
		fmt.Println("Error: --agent-id is required")
		fmt.Println("Use --help for usage information") 
		return 1
	}

	debug.SetMemoryLimit(*memoryLimit)

	logger, err := log.New(logConfig)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create logger",
			slog.Any("err", err),
		)
		return 1
	}

	logger.LogAttrs(ctx, slog.LevelDebug, "logging has started")

	if configFilePath != "" {
		logger.LogAttrs(ctx, slog.LevelInfo, "using configuration file: "+configFilePath)
	}

	if err = setPriorityWindows(ctx, logger, os.Getpid(), *processPriority); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to set process priority",
			slog.Any("err", err),
		)
		return 1
	}

	// Create push gateway configuration
	pushConfig := PushConfig{
		URL:      *pushGatewayURL,
		Username: *pushUsername,
		Password: *pushPassword,
		Interval: *pushInterval,
		AgentID:  *agentID,
		JobName:  *pushJobName,
	}

	enabledCollectorList := expandEnabledCollectors(*enabledCollectors)
	if err := collectors.Enable(enabledCollectorList); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "couldn't enable collectors",
			slog.Any("err", err),
		)
		return 1
	}

	// Initialize collectors
	if err = collectors.Build(ctx, logger); err != nil {
		for _, err := range utils.SplitError(err) {
			logger.LogAttrs(ctx, slog.LevelError, "couldn't initialize collector",
				slog.Any("err", err),
			)
			return 1
		}
	}

	logCurrentUser(ctx, logger)

	logger.InfoContext(ctx, "Enabled collectors: "+strings.Join(enabledCollectorList, ", "))
	logger.InfoContext(ctx, fmt.Sprintf("Agent ID: %s", pushConfig.AgentID))
	logger.InfoContext(ctx, fmt.Sprintf("Push Gateway URL: %s", pushConfig.URL))
	logger.InfoContext(ctx, fmt.Sprintf("Push Interval: %s", pushConfig.Interval))

	// Create Prometheus registry
	registry := prometheus.NewRegistry()

	// Create collector wrapper that adds agent_id label
	agentCollector := &AgentCollectorWrapper{
		collectors: collectors,
		agentID:    pushConfig.AgentID,
		logger:     logger,
	}

	registry.MustRegister(agentCollector)

	logger.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("starting windows_agent_collector in %s", time.Since(startTime)),
		slog.String("version", version.Version),
		slog.String("branch", version.Branch),
		slog.String("revision", version.GetRevision()),
		slog.String("goversion", version.GoVersion),
		slog.String("builddate", version.BuildDate),
		slog.Int("maxprocs", runtime.GOMAXPROCS(0)),
	)

	// Start push gateway client
	if err := runPushGateway(ctx, logger, pushConfig, registry); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to run push gateway client",
			slog.Any("err", err),
		)
		return 1
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "windows_agent_collector has shut down")
	return 0
}

func runPushGateway(ctx context.Context, logger *slog.Logger, config PushConfig, registry *prometheus.Registry) error {
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Initial push
	if err := pushMetrics(ctx, logger, config, registry); err != nil {
		logger.LogAttrs(ctx, slog.LevelWarn, "Initial metrics push failed",
			slog.Any("err", err),
		)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-stopCh:
			return nil
		case <-ticker.C:
			if err := pushMetrics(ctx, logger, config, registry); err != nil {
				logger.LogAttrs(ctx, slog.LevelWarn, "Metrics push failed",
					slog.Any("err", err),
				)
			}
		}
	}
}

func pushMetrics(ctx context.Context, logger *slog.Logger, config PushConfig, registry *prometheus.Registry) error {
	pusher := push.New(config.URL, config.JobName).
		Gatherer(registry).
		Grouping("agent_id", config.AgentID)

	if config.Username != "" && config.Password != "" {
		pusher = pusher.BasicAuth(config.Username, config.Password)
	}

	start := time.Now()
	err := pusher.Push()
	duration := time.Since(start)

	if err != nil {
		if logger != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to push metrics",
				slog.Any("err", err),
				slog.Duration("duration", duration),
			)
		}
		return err
	}

	if logger != nil {
		logger.LogAttrs(ctx, slog.LevelDebug, "Successfully pushed metrics",
			slog.Duration("duration", duration),
		)
	}

	return nil
}

// AgentCollectorWrapper wraps the collector and adds agent_id label to all metrics
type AgentCollectorWrapper struct {
	collectors collector.Collection
	agentID    string
	logger     *slog.Logger
}

func (a *AgentCollectorWrapper) Describe(ch chan<- *prometheus.Desc) {
	a.collectors.Describe(ch)
}

func (a *AgentCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	originalCh := make(chan prometheus.Metric, 1000)
	go func() {
		defer close(originalCh)
		a.collectors.Collect(originalCh, a.logger, 30*time.Second)
	}()

	for metric := range originalCh {
		ch <- metric
	}
}

func logCurrentUser(ctx context.Context, logger *slog.Logger) {
	u, err := user.Current()
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelWarn, "Unable to determine which user is running this agent",
			slog.Any("err", err),
		)
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "Running as "+u.Username)

	if strings.Contains(u.Username, "ContainerAdministrator") || strings.Contains(u.Username, "ContainerUser") {
		logger.LogAttrs(ctx, slog.LevelWarn, "Running as a preconfigured Windows Container user. Some functionality may not work as expected.")
	}
}

// setPriorityWindows sets the priority of the current process to the specified value.
func setPriorityWindows(ctx context.Context, logger *slog.Logger, pid int, priority string) error {
	// Mapping of priority names to uint32 values required by windows.SetPriorityClass.
	priorityStringToInt := map[string]uint32{
		"realtime":    windows.REALTIME_PRIORITY_CLASS,
		"high":        windows.HIGH_PRIORITY_CLASS,
		"abovenormal": windows.ABOVE_NORMAL_PRIORITY_CLASS,
		"normal":      windows.NORMAL_PRIORITY_CLASS,
		"belownormal": windows.BELOW_NORMAL_PRIORITY_CLASS,
		"low":         windows.IDLE_PRIORITY_CLASS,
	}

	winPriority, ok := priorityStringToInt[priority]

	// Only set process priority if a non-default and valid value has been set
	if !ok || winPriority == windows.NORMAL_PRIORITY_CLASS {
		return nil
	}

	logger.LogAttrs(ctx, slog.LevelDebug, "setting process priority to "+priority)

	// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
	handle, err := windows.OpenProcess(
		windows.STANDARD_RIGHTS_REQUIRED|windows.SYNCHRONIZE|windows.SPECIFIC_RIGHTS_ALL,
		false, uint32(pid),
	)
	if err != nil {
		return fmt.Errorf("failed to open own process: %w", err)
	}

	if err = windows.SetPriorityClass(handle, winPriority); err != nil {
		return fmt.Errorf("failed to set priority class: %w", err)
	}

	if err = windows.CloseHandle(handle); err != nil {
		return fmt.Errorf("failed to close handle: %w", err)
	}

	return nil
}

func expandEnabledCollectors(enabled string) []string {
	// For our lightweight agent, we only support specific collectors
	supportedCollectors := []string{"cpu", "memory", "net", "pagefile"}

	// Handle empty input
	if enabled == "" {
		return []string{}
	}

	expanded := strings.ReplaceAll(enabled, "[defaults]", "cpu,memory,net,pagefile")
	requested := slices.Compact(strings.Split(expanded, ","))

	// Filter to only supported collectors
	var filtered []string
	for _, collector := range requested {
		if slices.Contains(supportedCollectors, collector) {
			filtered = append(filtered, collector)
		}
	}

	return filtered
}

// handleServiceInstall installs the Windows Agent Collector as a Windows service
func handleServiceInstall(ctx context.Context, args []string) int {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return 1
	}

	// Build service command with all arguments except "install"
	var serviceArgs []string
	for _, arg := range args {
		if arg != "install" {
			serviceArgs = append(serviceArgs, arg)
		}
	}

	// Install the service
	err = installService(execPath, serviceArgs)
	if err != nil {
		fmt.Printf("Error installing service: %v\n", err)
		return 1
	}

	fmt.Println("Windows Agent Collector service installed successfully")
	fmt.Println("To start the service, run: sc start windows_agent_collector")
	return 0
}

// handleServiceUninstall removes the Windows Agent Collector service
func handleServiceUninstall(ctx context.Context) int {
	err := uninstallService()
	if err != nil {
		fmt.Printf("Error uninstalling service: %v\n", err)
		return 1
	}

	fmt.Println("Windows Agent Collector service uninstalled successfully")
	return 0
}

// installService installs the Windows service
func installService(execPath string, args []string) error {
	const serviceName = "windows_agent_collector"
	const serviceDisplayName = "Windows Agent Collector"
	const serviceDescription = "Lightweight Windows metrics collector for WebRTC troubleshooting"

	// Build command line with arguments
	cmdLine := fmt.Sprintf(`"%s"`, execPath)
	if len(args) > 0 {
		cmdLine += " " + strings.Join(args, " ")
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Check if service already exists
	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// Create the service
	s, err = m.CreateService(serviceName, cmdLine, mgr.Config{
		DisplayName:      serviceDisplayName,
		Description:      serviceDescription,
		StartType:        mgr.StartAutomatic,
		ServiceStartName: "",
	})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	return nil
}

// uninstallService removes the Windows service
func uninstallService() error {
	const serviceName = "windows_agent_collector"

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer s.Close()

	// Stop the service if it's running
	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("failed to query service status: %w", err)
	}

	if status.State == svc.Running {
		_, err = s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		// Wait for service to stop
		for {
			status, err = s.Query()
			if err != nil {
				return fmt.Errorf("failed to query service status: %w", err)
			}
			if status.State == svc.Stopped {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Delete the service
	err = s.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}
