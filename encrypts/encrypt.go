package enc

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

func ZipDatabase(dirPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return AddFilesToZip(zipWriter, dirPath, "")
}

func AddFilesToZip(zipWriter *zip.Writer, basePath, baseInZip string) error {
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := basePath + "/" + file.Name()

		if file.IsDir() {
			AddFilesToZip(zipWriter, filePath, baseInZip+"/"+file.Name())
		} else {
			fileToAdd, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer fileToAdd.Close()

			writer, err := zipWriter.Create(baseInZip + "/" + file.Name())
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, fileToAdd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func EncryptFile(inputPath, outputPath string, key []byte) error {
	plaintext, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer plaintext.Close()

	ciphertext, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer ciphertext.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return err
	}

	writer := &cipher.StreamWriter{S: cipher.NewOFB(block, iv), W: ciphertext}

	if _, err := io.Copy(writer, plaintext); err != nil {
		return err
	}

	return nil
}

func main() {
	dirToZip := "./test_db"
	zipPath := "./test_db.zip"
	encryptedZipPath := "./encrypted_db.zip"
	key := []byte("longg-key-for-AES-256-encryption")

	err := ZipDatabase(dirToZip, zipPath)
	if err != nil {
		fmt.Println("Error zipping Database:", err)
		return
	}

	err = EncryptFile(zipPath, encryptedZipPath, key)
	if err != nil {
		fmt.Println("Error encrypting zip file:", err)
		return
	}

	fmt.Println("Database encrypted successfully!")
}
