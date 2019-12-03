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
	BatchRuns    []BatchRun    `json:"jobs"`
	RawBatchRuns []BatchRunRaw `json:"-"`
	Opts         ManifestOpts  `json:"opts"`
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

type BatchRunRaw struct {
	StartKeyLine *int
	EndKeyLine   *int
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func resolveBatchRuns(resp *Jobs) {
	// Get list of lines to fetch keys from, with no duplicates
	lines := make([]int, 0, len(resp.RawBatchRuns)*2) // Upto 2 lines per batch run
	for _, rbr := range resp.RawBatchRuns {
		if rbr.StartKeyLine != nil && !contains(lines, *rbr.StartKeyLine) {
			lines = append(lines, *rbr.StartKeyLine)
		}
		if rbr.EndKeyLine != nil && !contains(lines, *rbr.EndKeyLine) {
			lines = append(lines, *rbr.EndKeyLine)
		}
	}

	// Create map for line/key lookups
	lineMap := make(map[int]*string)
	keys := getKeysAtLines(MANIFEST_FILE, lines)
	if len(keys) != len(lines) {
		panic("Num keys doesn't match number of lines")
	}
	for i := 0; i < len(keys); i++ {
		lineMap[lines[i]] = keys[i]
	}

	// Set the start and end keys
	resp.BatchRuns = make([]BatchRun, len(resp.RawBatchRuns))
	for i, rbr := range resp.RawBatchRuns {
		br := BatchRun{}

		if rbr.StartKeyLine != nil {
			br.StartKey = lineMap[*rbr.StartKeyLine]
		}
		if rbr.EndKeyLine != nil {
			br.EndKey = lineMap[*rbr.EndKeyLine]
		}

		resp.BatchRuns[i] = br
	}

}

func calculateStartEndKeys(batchSize, batchIndex int) BatchRunRaw {
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
		return BatchRunRaw{
			nil,
			&endLine,
		}
	}

	if (batchIndex + 1) == batchSize {
		return BatchRunRaw{
			&startLine,
			nil,
		}
	}

	return BatchRunRaw{
		&startLine,
		&endLine,
	}
}
