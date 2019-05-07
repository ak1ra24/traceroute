package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	// "./traceroute"
	// "./tracert"
	"github.com/ak1ra24/traceroute/traceroute"
	"github.com/ak1ra24/traceroute/tracert"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage: ./trace <host or ipaddress>")
		os.Exit(1)
	}
	host := os.Args[1]
	switch runtime.GOOS {
	case "darwin":
	case "linux":
		traceroute.Traceroute(host)
	default:
		fmt.Println("EXEC tracert. So it takes long long time.")
		fmt.Println("Please wait")
		isExist := isCommandExist("tracert")
		if isExist {
			fmt.Println("EXIST COMMAND: tracert")
		} else {
			os.Exit(1)
		}

		err := tracert.Traceroute_Windows(host)
		if err != nil {
			log.Fatalln(err.Error())
			os.Exit(1)
		}
	}
}

func isCommandExist(cmdname string) bool {
	cmd := exec.Command("cmd", "tracert")
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
