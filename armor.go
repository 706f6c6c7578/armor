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

const chunkSize = 4096 // 4KB chunks

// crlfWriter wraps an io.Writer and ensures CRLF line endings
type crlfWriter struct {
	w io.Writer
}

func (cw *crlfWriter) Write(p []byte) (n int, err error) {
	// Replace LF with CRLF
	p = bytes.ReplaceAll(p, []byte{'\n'}, []byte{'\r', '\n'})
	return cw.w.Write(p)
}

func encode(input io.Reader, output io.Writer) error {
	// Create a buffer to capture the output
	var buf bytes.Buffer
	
	// Wrap the buffer with our crlfWriter
	crlfOutput := &crlfWriter{w: &buf}

	w, err := armor.Encode(crlfOutput, "PGP MESSAGE", nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, input)
	if err != nil {
		return err
	}

	// Close the armor writer
	if err := w.Close(); err != nil {
		return err
	}

	// Trim any trailing newlines and add a single CRLF
	trimmedOutput := bytes.TrimRight(buf.Bytes(), "\r\n")
	trimmedOutput = append(trimmedOutput, '\r', '\n')

	// Write the final output
	_, err = output.Write(trimmedOutput)
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
	var output io.Writer = os.Stdout

	if flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: Too many arguments\n\n")
		printUsage()
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		input = os.Stdin
	} else {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer file.Close()
		input = file
	}

	var err error
	if *decodeFlag {
		err = decode(input, output)
	} else {
		err = encode(input, output)
	}

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}