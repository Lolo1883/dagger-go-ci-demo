package main

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	// Connect to Dagger Engine
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Writer()))
	if err != nil {
		log.Fatalf("Failed to connect to Dagger engine: %v", err)
	}
	defer client.Close()

	// Source code directory
	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"dagger/**", "Dockerfile", "build/**"},
	})

	// Go base image container
	goContainer := client.Container().
		From("golang:1.24").
		WithMountedDirectory("/app", src).
		WithWorkdir("/app").
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{"go", "test", "./..."}).
		WithExec([]string{"go", "build", "-o", "app"})

	// Final image based on distroless
	finalImage := client.Container().
		From("gcr.io/distroless/static").
		WithFile("/app", goContainer.File("/app/app")).
		WithEntrypoint([]string{"/app"}).
		WithExposedPort(8080)

	// Push to Docker Hub
	tag := "modock93/go-server:latest" // you can change this
	fmt.Println("📦 Publishing image to Docker Hub:", tag)

	_, err = finalImage.Publish(ctx, "docker.io/"+tag, dagger.ContainerPublishOpts{
		PlatformVariants: nil,
	})
	if err != nil {
		log.Fatalf("❌ Failed to push Docker image: %v", err)
	}

	fmt.Println("✅ Image pushed to Docker Hub: docker.io/" + tag)
}
