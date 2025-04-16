package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Writer()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to Dagger engine: %v", err)
	}
	defer client.Close()

	// Get Docker credentials from environment
	username := os.Getenv("DOCKERHUB_USERNAME")
	password := os.Getenv("DOCKERHUB_PASSWORD")

	if username == "" || password == "" {
		log.Fatal("DOCKERHUB_USERNAME or DOCKERHUB_PASSWORD not set")
	}

	// Mount project dir
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"dagger/**", "Dockerfile", "build/**"},
	})

	// Build Go app
	goContainer := client.Container().
		From("golang:1.24").
		WithMountedDirectory("/app", src).
		WithWorkdir("/app").
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{"go", "test", "./..."}).
		WithExec([]string{"go", "build", "-o", "app"})

	// Build final image
	finalImage := client.Container().
		From("gcr.io/distroless/static").
		WithFile("/app", goContainer.File("/app/app")).
		WithEntrypoint([]string{"/app"}).
		WithExposedPort(8080).
		WithRegistryAuth("docker.io", username, client.SetSecret("dockerhub-password", password))

	tag := "modock93/go-server:latest"
	fmt.Println("📦 Publishing to Docker Hub as:", tag)

	_, err = finalImage.Publish(ctx, "docker.io/"+tag)
	if err != nil {
		log.Fatalf("❌ Failed to push Docker image: %v", err)
	}

	fmt.Println("✅ Successfully pushed to Docker Hub: docker.io/" + tag)
}
