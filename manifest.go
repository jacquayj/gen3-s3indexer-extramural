package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
)

const chunkSize = 1024 * 1024 * 64

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, chunkSize)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}

func getKeyAtLine(path string, targetLine int) *string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	line := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if targetLine == line {
			key := scanner.Text()
			return &key
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func getKeyAtLine2(path string, targetLine, targetLine2 int) (*string, *string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var first *string
	line := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := scanner.Text()
		if targetLine == line {
			first = &key
		} else {
			if targetLine2 == line {
				last := &key
				return first, last
			}
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil, nil
}

func getManifestNumLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	return lineCounter(file)
}
