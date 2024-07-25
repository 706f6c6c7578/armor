package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/openpgp/armor"
)

func encode(input io.Reader, output io.Writer) error {
	w, err := armor.Encode(output, "PGP MESSAGE", nil)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, input)
	return err
}

func decode(input io.Reader, output io.Writer) error {
	dec, err := armor.Decode(input)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, dec.Body)
	return err
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-d] [file]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -d    Decode instead of encode\n")
	fmt.Fprintf(os.Stderr, "  file  Input file (optional, stdin used if not provided)\n")
	fmt.Fprintf(os.Stderr, "\nIf no file is specified, the program will read from stdin.\n")
	fmt.Fprintf(os.Stderr, "To finish input from stdin, press Ctrl+D (Unix) or Ctrl+Z (Windows).\n")
}

func main() {
	decodeFlag := flag.Bool("d", false, "Decode instead of encode")
	flag.Usage = printUsage
	flag.Parse()

	var input io.Reader
	var output io.Writer

	if flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: Too many arguments\n\n")
		printUsage()
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		// No file provided, use stdin
		input = os.Stdin
	} else {
		// Read from file
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer file.Close()
		input = file
	}

	// Use a buffer as output
	var buf bytes.Buffer
	output = &buf

	var err error
	if *decodeFlag {
		err = decode(input, output)
	} else {
		err = encode(input, output)
	}

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Write the buffer content to stdout
	fmt.Print(buf.String())

	// Add CRLF only for encoded (armored) output
	if !*decodeFlag {
		fmt.Print("\r\n")
	}
}