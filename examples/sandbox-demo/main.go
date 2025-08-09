package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/justinlyon12/ancli/internal/sandbox"
	_ "github.com/justinlyon12/ancli/internal/sandbox/podman" // Register driver
)

func main() {
	fmt.Println("🔧 AnCLI Sandbox Demo - Secure Command Execution")
	fmt.Println("================================================")

	// Check available drivers
	drivers := sandbox.Available()
	fmt.Printf("📦 Available sandbox drivers: %v\n", drivers)

	// Create a Podman driver instance
	driver, err := sandbox.Get("podman")
	if err != nil {
		log.Fatalf("❌ Failed to get podman driver: %v", err)
	}

	fmt.Printf("✅ Created driver: %s\n", driver.Name())

	// Create a secure execution config
	config := sandbox.NewExecutionConfig().
		WithImage("alpine:latest").
		WithCommand("echo", "Hello from secure AnCLI sandbox!").
		WithCorrelationID("demo-001")

	fmt.Println("\n🛡️  Security Configuration:")
	fmt.Printf("   • Image: %s\n", config.Image)
	fmt.Printf("   • Network enabled: %v\n", config.NetworkEnabled)
	fmt.Printf("   • Read-only filesystem: %v\n", config.ReadOnlyRootFS)
	fmt.Printf("   • Capabilities: %v (empty = all dropped)\n", config.Capabilities)
	fmt.Printf("   • Timeout: %v\n", config.Timeout)

	// Execute the command
	fmt.Println("\n🚀 Executing command in secure container...")
	ctx := context.Background()

	start := time.Now()
	result, err := driver.Run(ctx, config)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("❌ Command execution failed: %v", err)
	}

	// Display results
	fmt.Println("\n📊 Execution Results:")
	fmt.Printf("   • Exit code: %d\n", result.ExitCode)
	fmt.Printf("   • Success: %v\n", result.Success)
	fmt.Printf("   • Duration: %v\n", result.Duration)
	fmt.Printf("   • Total time (including setup): %v\n", duration)
	fmt.Printf("   • Container ID: %s\n", result.ContainerID)
	fmt.Printf("   • Stdout: %q\n", result.Stdout)
	fmt.Printf("   • Stderr: %q\n", result.Stderr)

	// Demonstrate session reuse with another command
	fmt.Println("\n🔄 Testing session reuse (should be faster)...")

	config2 := config.
		WithCommand("sh", "-c", "echo 'Second command' && pwd && ls -la /tmp").
		WithCorrelationID("demo-002")

	start = time.Now()
	result2, err := driver.Run(ctx, config2)
	duration2 := time.Since(start)

	if err != nil {
		log.Fatalf("❌ Second command failed: %v", err)
	}

	fmt.Printf("   • Duration: %v (should be ~10ms vs ~200ms for first)\n", result2.Duration)
	fmt.Printf("   • Total time: %v\n", duration2)
	fmt.Printf("   • Same container: %v\n", result.ContainerID == result2.ContainerID)
	fmt.Printf("   • Output:\n%s", result2.Stdout)

	// Test security - attempt to write to read-only filesystem
	fmt.Println("\n🔒 Testing security hardening...")

	config3 := config.
		WithCommand("sh", "-c", "echo 'Testing write to /' && echo test > /test.txt 2>&1 || echo 'Blocked: Read-only filesystem'").
		WithCorrelationID("demo-security")

	result3, err := driver.Run(ctx, config3)
	if err != nil && result3 == nil {
		log.Fatalf("❌ Security test failed: %v", err)
	}

	fmt.Printf("   • Security test output: %q\n", result3.Stdout)

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	if err := driver.Cleanup(ctx); err != nil {
		log.Printf("⚠️  Cleanup warning: %v", err)
	} else {
		fmt.Println("✅ Cleanup completed")
	}

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("   The sandbox provides secure, isolated command execution")
	fmt.Println("   with performance optimization through session reuse.")
}
