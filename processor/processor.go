package processor

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"check-backup-dicom-exams/model"
)

func ProcessExam(record []string, baseDir string) model.ExamRecord {
	id := strings.TrimSpace(record[0])
	examID := strings.TrimSpace(record[1])
	patientID := strings.TrimSpace(record[2])
	studyUID := strings.TrimSpace(record[3])

	examDir := ""
	_ = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && filepath.Base(path) == examID {
			examDir = path
			return filepath.SkipDir
		}
		return nil
	})

	foundPath := "not_found"
	dcmCount := 0
	pdfCount := 0

	if examDir != "" {
		foundPath = examDir
		_ = filepath.Walk(examDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				name := strings.ToLower(info.Name())
				if strings.HasSuffix(name, ".dcm") {
					dcmCount++
				}
				if strings.HasSuffix(name, ".pdf") {
					pdfCount++
				}
			}
			return nil
		})
	}

	return model.ExamRecord{
		FoundPath: foundPath,
		DCMCount:  dcmCount,
		PDFCount:  pdfCount,
		ID:        id,
		ExamID:    examID,
		PatientID: patientID,
		StudyUID:  studyUID,
	}
}

func WriteSummary(path string, total, dcmCount, pdfCount int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "Verification Report\n")
	fmt.Fprintf(file, "-----------------------\n")
	fmt.Fprintf(file, "Total exams processed in this run: %d\n", total)
	fmt.Fprintf(file, "Exams with at least one .dcm file: %d\n", dcmCount)
	fmt.Fprintf(file, "Exams with at least one .pdf file: %d\n", pdfCount)

	return nil
}

func WriteCSV(outputPath string, headers []string, results <-chan model.ExamRecord, wg *sync.WaitGroup, dcmCounter *int, pdfCounter *int, mu *sync.Mutex) error {
	defer wg.Done()

	fileExists := false
	if _, err := os.Stat(outputPath); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		writer.Write(headers)
	}

	for result := range results {
		writer.Write([]string{
			result.FoundPath,
			fmt.Sprintf("%d", result.DCMCount),
			fmt.Sprintf("%d", result.PDFCount),
			result.ID, result.ExamID, result.PatientID, result.StudyUID,
		})

		mu.Lock()
		if result.DCMCount > 0 {
			*dcmCounter++
		}
		if result.PDFCount > 0 {
			*pdfCounter++
		}
		mu.Unlock()
	}

	return nil
}
