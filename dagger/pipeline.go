package main

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Writer()))
	if err != nil {
		log.Fatalf("Failed to connect to Dagger engine: %v", err)
	}
	defer client.Close()

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"dagger/**", "Dockerfile", "build/**"},
	})

	goContainer := client.Container().
		From("golang:1.24").
		WithMountedDirectory("/app", src).
		WithWorkdir("/app").
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{"go", "test", "./..."}).
		WithExec([]string{"go", "build", "-o", "app"})

	finalImage := client.Container().
		From("gcr.io/distroless/static").
		WithFile("/app", goContainer.File("/app/app")).
		WithEntrypoint([]string{"/app"}).
		WithExposedPort(8080).
		WithRegistryAuth(
			"docker.io",
			client.Host().EnvVariable("DOCKERHUB_USERNAME"),
			client.Host().EnvVariable("DOCKERHUB_PASSWORD"),
		)

	tag := "modock93/go-server:latest"
	fmt.Println("📦 Publishing image to Docker Hub:", tag)

	_, err = finalImage.Publish(ctx, "docker.io/"+tag)
	if err != nil {
		log.Fatalf("❌ Failed to push Docker image: %v", err)
	}

	fmt.Println("✅ Image pushed to Docker Hub:
