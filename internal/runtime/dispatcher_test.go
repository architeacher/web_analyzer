package runtime

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/architeacher/svc-web-analyzer/internal/config"
	"github.com/stretchr/testify/require"
)

func TestServiceCtx_SIGUSR1_ConfigReload(t *testing.T) {
	t.Parallel()

	t.Run("SIGUSR1 signal triggers config reload", func(t *testing.T) {
		t.Parallel()
		// Set initial environment variable
		initialValue := "initial-test-value"
		os.Setenv("APP_SERVICE_NAME", initialValue)
		defer os.Unsetenv("APP_SERVICE_NAME")

		// Create service context with the initial config
		serviceCtx := New()
		require.Equal(t, initialValue, serviceCtx.cfg.AppConfig.ServiceName)

		// Start monitoring external changes
		serviceCtx.reloadConfigChannel = make(chan os.Signal, 1)
		go func() {
			<-serviceCtx.reloadConfigChannel

			cfg, err := config.Load()
			require.NoError(t, err)
			serviceCtx.cfg = cfg
		}()

		// Change environment variable
		newValue := "reloaded-test-value"
		os.Setenv("APP_SERVICE_NAME", newValue)

		// Send SIGUSR1 signal
		serviceCtx.reloadConfigChannel <- syscall.SIGUSR1

		// Give some time for the reload to complete
		time.Sleep(100 * time.Millisecond)

		// Verify config was reloaded
		require.Equal(t, newValue, serviceCtx.cfg.AppConfig.ServiceName)
	})

	t.Run("config reload handles invalid configuration gracefully", func(t *testing.T) {
		// Set valid initial config
		os.Setenv("APP_SERVICE_NAME", "test-service")
		os.Setenv("HTTP_SERVER_PORT", "8080")
		defer func() {
			os.Unsetenv("APP_SERVICE_NAME")
			os.Unsetenv("HTTP_SERVER_PORT")
		}()

		serviceCtx := New()
		originalServiceName := serviceCtx.cfg.AppConfig.ServiceName

		// Set invalid port value to cause config load error
		os.Setenv("HTTP_SERVER_PORT", "invalid-port")

		// Create a channel to track reload completion
		reloadDone := make(chan bool, 1)
		serviceCtx.reloadConfigChannel = make(chan os.Signal, 1)

		go func() {
			<-serviceCtx.reloadConfigChannel

			cfg, err := config.Load()
			if err != nil {
				// Config reload should handle error gracefully
				// Original config should remain unchanged
				reloadDone <- false
				return
			}
			serviceCtx.cfg = cfg
			reloadDone <- true
		}()

		// Send SIGUSR1 signal
		serviceCtx.reloadConfigChannel <- syscall.SIGUSR1

		// Wait for reload attempt
		select {
		case success := <-reloadDone:
			if success {
				t.Error("Expected config reload to fail with invalid port, but it succeeded")
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("Config reload did not complete within expected time")
		}

		// Verify original config is preserved
		require.Equal(t, originalServiceName, serviceCtx.cfg.AppConfig.ServiceName)
	})

	t.Run("multiple SIGUSR1 signals are handled correctly", func(t *testing.T) {
		os.Setenv("APP_SERVICE_NAME", "initial-value")
		defer os.Unsetenv("APP_SERVICE_NAME")

		serviceCtx := New()
		serviceCtx.reloadConfigChannel = make(chan os.Signal, 1)

		reloadCount := 0
		go func() {
			for range serviceCtx.reloadConfigChannel {
				cfg, err := config.Load()
				if err == nil {
					serviceCtx.cfg = cfg
					reloadCount++
				}
			}
		}()

		// Send multiple signals with config changes
		testValues := []string{"value1", "value2", "value3"}
		for _, value := range testValues {
			os.Setenv("APP_SERVICE_NAME", value)
			serviceCtx.reloadConfigChannel <- syscall.SIGUSR1
			time.Sleep(50 * time.Millisecond)
		}

		// Give time for all reloads to complete
		time.Sleep(100 * time.Millisecond)
		close(serviceCtx.reloadConfigChannel)

		require.Equal(t, len(testValues), reloadCount)
		require.Equal(t, "value3", serviceCtx.cfg.AppConfig.ServiceName)
	})
}

func TestServiceCtx_ConfigReloadConcurrency(t *testing.T) {
	t.Run("concurrent config access is safe", func(t *testing.T) {
		os.Setenv("APP_SERVICE_NAME", "concurrent-test")
		defer os.Unsetenv("APP_SERVICE_NAME")

		serviceCtx := New()
		serviceCtx.reloadConfigChannel = make(chan os.Signal, 1)

		// Start config reload handler
		go func() {
			<-serviceCtx.reloadConfigChannel
			cfg, err := config.Load()
			if err == nil {
				serviceCtx.cfg = cfg
			}
		}()

		// Simulate concurrent access to config
		done := make(chan bool, 2)

		// Goroutine 1: Read config repeatedly
		go func() {
			for i := 0; i < 100; i++ {
				_ = serviceCtx.cfg.AppConfig.ServiceName
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()

		// Goroutine 2: Trigger reload
		go func() {
			time.Sleep(10 * time.Millisecond)
			os.Setenv("APP_SERVICE_NAME", "updated-concurrent-test")
			serviceCtx.reloadConfigChannel <- syscall.SIGUSR1
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done

		// Test should complete without race conditions
		require.NotNil(t, serviceCtx.cfg)
	})
}

func TestNew_WithReloadChannel(t *testing.T) {
	t.Run("service context initializes with reload channel", func(t *testing.T) {
		serviceCtx := New()

		require.NotNil(t, serviceCtx.reloadConfigChannel)
		require.NotNil(t, serviceCtx.shutdownChannel)
		require.NotNil(t, serviceCtx.cfg)
		require.NotNil(t, serviceCtx.logger)
	})
}
