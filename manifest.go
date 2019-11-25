package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math"
	"os"
	"sort"
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

func getKeysAtLines(path string, targetLines []int) []*string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	sort.Ints(targetLines)

	keys := make([]*string, 0, 100)

	line := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := scanner.Text()
		if targetLines[0] == line {
			keys = append(keys, &key)
			if len(targetLines) == 1 {
				return keys
			}
			targetLines = targetLines[1:]
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return keys
}

func getManifestNumLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	return lineCounter(file)
}

type BatchRun struct {
	StartKey *string
	EndKey   *string
}

func calculateStartEndKeys() BatchRun {
	numTotalObjs, err := getManifestNumLines("/manifest.txt")
	if err != nil {
		log.Fatal(err)
	}
	objsPerNode := int(math.Ceil(float64(numTotalObjs) / float64(batchSize)))

	startLine := batchIndex * objsPerNode
	endLine := (batchIndex + 1) * objsPerNode

	if batchIndex == 0 {
		return BatchRun{
			nil,
			getKeyAtLine("/manifest.txt", endLine),
		}
	}

	if (batchIndex + 1) == batchSize {
		return BatchRun{
			getKeyAtLine("/manifest.txt", startLine),
			nil,
		}
	}

	keys := getKeysAtLines("/manifest.txt", []int{startLine, endLine})
	return BatchRun{
		keys[0],
		keys[1],
	}
}
