package enc

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestZipDatabase(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "test_db_dir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy file inside the temporary directory
	dummyFilePath := tempDir + "/dummy.txt"
	if err := ioutil.WriteFile(dummyFilePath, []byte("test database content"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	zipPath := "./test_db.zip"
	defer os.Remove(zipPath)

	// Run the function
	if err := ZipDatabase(tempDir, zipPath); err != nil {
		t.Fatalf("Error zipping database: %v", err)
	}

	// Check if the ZIP file was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Fatalf("ZIP file was not created")
	}
}

func TestEncryptFile(t *testing.T) {
	inputPath := "./test_encrypt_input.zip"
	outputPath := "./encrypted_db.zip"
	key := []byte("longg-key-for-AES-256-encryption")

	// Create a dummy input file
	if err := ioutil.WriteFile(inputPath, []byte("dummy zip content"), 0644); err != nil {
		t.Fatalf("Failed to create dummy input file: %v", err)
	}
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	// Run the function
	if err := EncryptFile(inputPath, outputPath, key); err != nil {
		t.Fatalf("Error encrypting file: %v", err)
	}

	// Check if the encrypted file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Encrypted file was not created")
	}
}

func TestAddFilesToZip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some temporary files
	file1Path := tempDir + "/file1.txt"
	file2Path := tempDir + "/file2.txt"

	// Write some content to the files
	content := []byte("Hello, world!")
	if err := ioutil.WriteFile(file1Path, content, 0644); err != nil {
		t.Fatalf("Failed to write content to file1: %v", err)
	}
	if err := ioutil.WriteFile(file2Path, content, 0644); err != nil {
		t.Fatalf("Failed to write content to file2: %v", err)
	}

	// Create a ZIP file and add the temporary files
	zipPath := "./test_db.zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create ZIP file: %v", err)
	}
	defer os.Remove(zipPath)

	zipWriter := zip.NewWriter(zipFile)
	err = AddFilesToZip(zipWriter, tempDir, "")
	
	// Close writer and file explicitly before reading
	zipWriter.Close()
	zipFile.Close()

	if err != nil {
		t.Fatalf("Error adding files to ZIP: %v", err)
	}

	// Read the contents of the ZIP file
	zipContent, err := ioutil.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("Failed to read ZIP file: %v", err)
	}

	// Check if the ZIP file contains the expected content
	expectedContent := map[string][]byte{
		"/file1.txt": content,
		"/file2.txt": content,
	}

	reader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		t.Fatalf("Failed to create ZIP reader: %v", err)
	}

	for _, file := range reader.File {
		expected, ok := expectedContent[file.Name]
		if !ok {
			t.Fatalf("Unexpected file found in ZIP: %s", file.Name)
		}

		zipFile, err := file.Open()
		if err != nil {
			t.Fatalf("Failed to open file in ZIP: %v", err)
		}

		zipContent, err := ioutil.ReadAll(zipFile)
		if err != nil {
			t.Fatalf("Failed to read file in ZIP: %v", err)
		}

		if !bytes.Equal(zipContent, expected) {
			t.Fatalf("ZIP file content does not match expected")
		}
	}
}
