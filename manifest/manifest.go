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

type BatchRun struct {
	StartKey *string `json:"start_key"`
	EndKey   *string `json:"end_key"`
}

type Jobs struct {
	BatchRuns []BatchRun `json:"jobs"`
}

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

var numTotalObjs = -1
var objsPerNode = -1

func calculateStartEndKeys(batchSize, batchIndex int) BatchRun {
	if numTotalObjs == -1 {
		if numTotalObj, err := getManifestNumLines(MANIFEST_FILE); err == nil {
			numTotalObjs = numTotalObj
			objsPerNode = int(math.Ceil(float64(numTotalObjs) / float64(batchSize)))
		} else {
			log.Fatal(err)
		}
	}

	startLine := batchIndex * objsPerNode
	endLine := (batchIndex + 1) * objsPerNode

	if batchIndex == 0 {
		return BatchRun{
			nil,
			getKeyAtLine(MANIFEST_FILE, endLine),
		}
	}

	if (batchIndex + 1) == batchSize {
		return BatchRun{
			getKeyAtLine(MANIFEST_FILE, startLine),
			nil,
		}
	}

	keys := getKeysAtLines(MANIFEST_FILE, []int{startLine, endLine})
	return BatchRun{
		keys[0],
		keys[1],
	}
}
