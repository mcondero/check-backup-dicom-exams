package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"

	"check-backup-dicom-exams/config"
	"check-backup-dicom-exams/model"
	"check-backup-dicom-exams/processor"
)

func countProcessedLines(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count > 0 {
		count--
	}
	return count
}

func main() {
	cfg, err := config.LoadConfig("check-cfg.yaml")
	if err != nil {
		fmt.Println("Error loading check-config.yaml:", err)
		return
	}

	skipLines := countProcessedLines(cfg.OutputCSV)

	file, err := os.Open(cfg.CSVPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	_, _ = reader.Read()

	var records [][]string
	currentLine := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err == nil && len(line) >= 4 {
			if currentLine < skipLines {
				currentLine++
				continue
			}
			records = append(records, line)
		}
	}

	resultChan := make(chan model.ExamRecord, len(records))
	var wg sync.WaitGroup
	var writeWg sync.WaitGroup
	var mu sync.Mutex
	dcmCount := 0
	pdfCount := 0

	writeWg.Add(1)
	go processor.WriteCSV(cfg.OutputCSV,
		[]string{"found_path", "dcm_count", "pdf_count", "id", "exam_id", "patient_id", "study_uid"},
		resultChan, &writeWg, &dcmCount, &pdfCount, &mu)

	sem := make(chan struct{}, 8)
	for _, rec := range records {
		wg.Add(1)
		sem <- struct{}{}
		go func(r []string) {
			defer wg.Done()
			result := processor.ProcessExam(r, cfg.BaseDir)
			resultChan <- result
			<-sem
		}(rec)
	}

	wg.Wait()
	close(resultChan)
	writeWg.Wait()

	err = processor.WriteSummary(cfg.SummaryTXT, len(records), dcmCount, pdfCount)
	if err != nil {
		fmt.Println("Error saving summary:", err)
	}
}
