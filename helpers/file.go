package helpers

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadImage(imageUrl string, retuenLocalPath bool) (string, error) {
	filenameExt := getFileExtension(imageUrl)
	randomId, err := generateRandomID(16)
	if err != nil {
		return "", err
	}
	// create folder if not exists
	if _, err := os.Stat("./public/uploads"); os.IsNotExist(err) {
		err := os.MkdirAll("./public/uploads", os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	filename := randomId + filenameExt
	filepath := filepath.Join("./public/uploads", filename)

	// Send HTTP GET request
	response, err := http.Get(imageUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return "./" + filepath, nil
}

func generateRandomID(length int) (string, error) {
	// Generate 16 random bytes
	randomBytes := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, randomBytes)
	if err != nil {
		return "", err
	}

	// Format as UUID (8-4-4-4-12)
	uuid := fmt.Sprintf("%x", randomBytes[10:])

	return uuid, nil
}

func getFileExtension(url string) string {
	params := strings.Split(url, "?")
	// Split the URL by the last "/"
	parts := strings.Split(params[0], "/")
	// Get the last part which contains the filename
	filename := parts[len(parts)-1]
	// Get the file extension
	extension := filepath.Ext(filename)
	// Remove the leading "." from the extension
	// extension = strings.TrimPrefix(extension, ".")
	return extension
}
