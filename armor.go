package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/openpgp/armor"
)

const chunkSize = 4096 // 4KB chunks

func encode(input io.Reader, output io.Writer) error {
	bufOutput := bufio.NewWriter(output)
	
	w, err := armor.Encode(bufOutput, "PGP MESSAGE", nil)
	if err != nil {
		return err
	}

	buf := make([]byte, chunkSize)
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		_, err = w.Write(buf[:n])
		if err != nil {
			return err
		}

		if err == io.EOF {
			break
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	if _, err := bufOutput.Write([]byte("\r\n")); err != nil {
		return err
	}

	return bufOutput.Flush()
}

func decode(input io.Reader, output io.Writer) error {
	dec, err := armor.Decode(input)
	if err != nil {
		return err
	}

	buf := make([]byte, chunkSize)
	for {
		n, err := dec.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		_, err = output.Write(buf[:n])
		if err != nil {
			return err
		}

		if err == io.EOF {
			break
		}
	}
	return nil
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
