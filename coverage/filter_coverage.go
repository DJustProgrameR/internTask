package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run filter_coverage.go <input_coverage_file> <output_coverage_file>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	includePatterns := []string{
		"mode:",
		"internshipPVZ/internal/domain/",
		"internshipPVZ/internal/http/",
		"internshipPVZ/internal/middleware/",
		"internshipPVZ/internal/repository/",
		"internshipPVZ/internal/usecase/",
	}

	excludePatterns := []string{
		"internshipPVZ/internal/grpc/",
		"internshipPVZ/internal/prometheus/",
		"internshipPVZ/cmd/",
		"internshipPVZ/migrations/",
		"internshipPVZ/test/",
	}

	filterCoverage(inputFile, outputFile, includePatterns, excludePatterns)
}

func filterCoverage(inputFile, outputFile string, includePatterns, excludePatterns []string) {
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer output.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(output)

	for scanner.Scan() {
		line := scanner.Text()
		if shouldInclude(line, includePatterns, excludePatterns) {
			writer.WriteString(line + "\n")
		}
	}

	writer.Flush()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func shouldInclude(line string, includePatterns, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}
	for _, pattern := range includePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
}
