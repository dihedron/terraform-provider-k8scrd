package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Result struct {
	CommandLine string `json:"commandline"`
	InputData   string `json:"input"`
	OutputData  string `json:"output"`
}

func main() {
	result := &Result{
		CommandLine: strings.Join(os.Args, " "),
	}
	// fmt.Printf("COMMAND LINE  : %s\n", strings.Join(os.Args, " "))
	// fmt.Println("STANDARD INPUT:")
	var buffer bytes.Buffer
	hasher := sha256.New()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
		buffer.Write([]byte("\n"))
		hasher.Write(scanner.Bytes())
		hasher.Write([]byte("\n"))
	}
	result.InputData = buffer.String()
	result.OutputData = hex.EncodeToString(hasher.Sum(nil))

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Println(err)
	}

	red := color.New(color.FgRed)
	red.Printf("%s\n", string(output))
}
