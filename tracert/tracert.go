package tracert

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Trace struct {
	response_time int
	hostname      string
	addr          string
}

func Traceroute_Windows(host string) error {
	cmd := exec.Command("tracert", host)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	// printCommand(cmd)
	err := cmd.Run()
	if err != nil {
		printError(err)
		return err
	}
	output := cmdOutput.Bytes()
	printOutput(output)

	return nil
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		// fmt.Printf("==> Output: %s\n", string(outs))
		var trace Trace
		traces := parseOutput(outs, trace)
		if traces == nil {
			err := errors.New("ERROR:TIMEOUT")
			fmt.Printf("%v\n", err)
		}
		for num, traceroute := range traces {
			fmt.Printf("%d\t%-16s[%-40s]\t%-3d ms\n", num+1, traceroute.addr, traceroute.hostname, traceroute.response_time)
		}
		// fmt.Println(traces)
	}
}

func parseOutput(outs []byte, trace Trace) []Trace {
	var traces []Trace
	char_timeout_reg := regexp.MustCompile(`(\s)+\*(\s)+\*(\s)+\*`)
	match := char_timeout_reg.FindAllStringSubmatch(string(outs), -1)
	if len(match) > 2 {
		return nil
	}
	rese := regexp.MustCompile(`(\d+ +ms)(\s)*([\w/:%#\$&\?\(\)~\.=\+\-]+)?(\s)*(\[)?\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}(\])?`)
	results := rese.FindAllString(string(outs), -1)
	for _, result := range results {
		response_reg := regexp.MustCompile(`(\d+ +ms)`)
		response := response_reg.FindString(result)
		host_reg := regexp.MustCompile(`[\w/:%#\$&\?\(\)~\.=\+\-]+`)
		host := host_reg.FindAllString(result, -1)
		response_time := strings.Replace(response, " ms", "", -1)
		trace.response_time, _ = strconv.Atoi(response_time)
		ip := net.ParseIP(host[2])
		if ip == nil {
			addr_reg := regexp.MustCompile(`\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
			addr := addr_reg.FindAllString(result, -1)
			trace.addr = addr[0]
			trace.hostname = host[2]
		} else {
			trace.hostname = ""
			trace.addr = host[2]
		}
		traces = append(traces, trace)
	}

	return traces
}
