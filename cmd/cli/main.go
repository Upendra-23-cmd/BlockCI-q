package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  cli submit <pipeline.yaml>")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	command := os.Args[1]
	switch command {
	case "submit":
		pipelinePath := os.Args[2]

		// read yaml file
		data, err := os.ReadFile(pipelinePath)
		if err != nil {
			fmt.Println("❌ Failed to read pipeline file:", err)
			os.Exit(1)
		}

		// send to server
		resp, err := http.Post("http://localhost:8080/pipelines", "application/x-yaml", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("❌ Failed to send request:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Println("✅ Server response:", string(body))

	default:
		usage()
	}
}

