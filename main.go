package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// compressAndEncodeFiles processes files from a directory or list of paths,
// compresses them with zlib and encodes with base64
func compressAndEncodeFiles(directoryPath string, filePaths []string) (string, error) {
	filesData := make(map[string][]byte)

	// Process directory
	if directoryPath != "" {
		err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %s: %v\n", path, err)
				return nil // Continue walking despite error
			}

			if !info.IsDir() {
				relPath, err := filepath.Rel(directoryPath, path)
				if err != nil {
					fmt.Printf("Error determining relative path for %s: %v\n", path, err)
					return nil
				}

				fileContent, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", path, err)
					return nil
				}

				// Compress with zlib
				var compressedContent bytes.Buffer
				w := zlib.NewWriter(&compressedContent)
				_, err = w.Write(fileContent)
				if err != nil {
					fmt.Printf("Error compressing file %s: %v\n", path, err)
					return nil
				}
				w.Close()

				// Store the compressed content
				filesData[relPath] = compressedContent.Bytes()
			}
			return nil
		})

		if err != nil {
			return "", fmt.Errorf("error walking directory: %v", err)
		}
	}

	// Process individual files
	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("File not found: %s\n", filePath)
			continue
		}

		if fileInfo.IsDir() {
			fmt.Printf("Skipping directory: %s\n", filePath)
			continue
		}

		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		// Compress with zlib
		var compressedContent bytes.Buffer
		w := zlib.NewWriter(&compressedContent)
		_, err = w.Write(fileContent)
		if err != nil {
			fmt.Printf("Error compressing file %s: %v\n", filePath, err)
			continue
		}
		w.Close()

		// Store the compressed content using just the filename
		filesData[filepath.Base(filePath)] = compressedContent.Bytes()
	}

	// Convert binary data to base64 strings for JSON serialization
	encodedData := make(map[string]string)
	for filePath, compressedContent := range filesData {
		encodedData[filePath] = base64.StdEncoding.EncodeToString(compressedContent)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(encodedData)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Final base64 encoding
	finalEncoded := base64.StdEncoding.EncodeToString(jsonData)

	return finalEncoded, nil
}

func main() {
	directoryFlag := flag.String("d", "", "Directory containing files to process")
	outputFlag := flag.String("o", "", "Output file for the encoded data")

	// Parse flag to handle file list
	flag.Parse()

	// Non-flag arguments after the flags are treated as file paths
	files := flag.Args()

	// Check if we have required args
	if *directoryFlag == "" && len(files) == 0 {
		fmt.Println("At least one of --d (directory) or file paths must be provided")
		flag.Usage()
		os.Exit(1)
	}

	result, err := compressAndEncodeFiles(*directoryFlag, files)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Write to output file or stdout
	if *outputFlag != "" {
		err := os.WriteFile(*outputFlag, []byte(result), 0644)
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Compressed and encoded data written to %s\n", *outputFlag)
	} else {
		fmt.Println(result)
	}
}
