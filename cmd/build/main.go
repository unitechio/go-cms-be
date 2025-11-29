package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	os.MkdirAll("tmp", 0755)

	bin := "tmp/main"
	if runtime.GOOS == "windows" {
		bin = "tmp/main.exe"
	}

	fmt.Println("Building for:", runtime.GOOS)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Build failed:", err)
		os.Exit(1)
	}

	fmt.Println("Build success:", bin)
}
