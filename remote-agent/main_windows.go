//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "symphony-service"
const serviceDesc = "Remote Agent Service"

type myService struct{}

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	log.Printf("Service Execute called with args: %v", args)
	s <- svc.Status{State: svc.StartPending}
	go func() {
		_ = mainLogic() // run main logic in goroutine
	}()
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			return false, 0
		default:
		}
	}
}

func isWindowsService() bool {
	log.Printf("Checking if running as a Windows service")
	ok, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("svc.IsWindowsService() error: %v", err)
	}
	return ok
}

func installService(exePath string, args []string) error {
	log.Printf("Installing service %s at %s with args: %v", serviceName, exePath, args)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}
	// Build the full command line
	binPath := "\"" + exePath + "\""
	for _, arg := range args {
		binPath += " " + arg
	}
	s, err = m.CreateService(serviceName, binPath, mgr.Config{
		DisplayName: serviceName,
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()
	if err := s.Delete(); err != nil {
		return err
	}
	_ = eventlog.Remove(serviceName)
	return nil
}

func startService() error {
	log.Printf("Starting service %s", serviceName)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()
	return s.Start()
}

func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()
	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}
	for status.State != svc.Stopped {
		status, err = s.Query()
		if err != nil {
			break
		}
	}
	return nil
}

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			// Pass all arguments after "install"
			if err := installService(execPath, os.Args[2:]); err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
			fmt.Println("Service installed. with arguments:", os.Args[2:])
			return
		case "uninstall":
			if err := removeService(); err != nil {
				log.Fatalf("Failed to remove service: %v", err)
			}
			fmt.Println("Service removed.")
			return
		case "start":
			if err := startService(); err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
			fmt.Println("Service started.")
			return
		case "stop":
			if err := stopService(); err != nil {
				log.Fatalf("Failed to stop service: %v", err)
			}
			fmt.Println("Service stopped.")
			return
		}
	}

	isService := isWindowsService()
	if isService {
		err := svc.Run(serviceName, &myService{})
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
		return
	}
	// running in console
	if err := mainLogic(); err != nil {
		log.Fatalf("mainLogic error: %v", err)
	}
}
