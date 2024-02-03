package storage

import (
	"fmt"
	"io"
	"os"

	"github.com/colinmarc/hdfs/v2"
	log "github.com/sirupsen/logrus"
)

// NoSQL: FileStorage struct encapsulating HDFS client
type FileStorage struct {
	client *hdfs.Client
}

func New() (*FileStorage, error) {
	// Local instance
	hdfsUri := os.Getenv("HDFS_URI")

	client, err := hdfs.New(hdfsUri)
	if err != nil {
		log.Panic(fmt.Sprintf("[acco-hdfs]acfs#1 Failed to create client: %v", err))
		return nil, err
	}

	// Return storage handler with logger and HDFS client
	return &FileStorage{
		client: client,
	}, nil
}

func (fs *FileStorage) Close() {
	// Close all underlying connections to the HDFS server
	fs.client.Close()
}

func (fs *FileStorage) CreateDirectories() error {
	// Default permissions
	// 0644 Only the owner can read and write. Everyone else can only read. No one can execute the file.
	err := fs.client.MkdirAll(hdfsCopyDir, 0644)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#2 Failed to create copy directories: %v", err))
		return err
	}

	// NoSQL TODO: What is the difference between MkdirAll and Mkdir?
	err = fs.client.Mkdir(hdfsWriteDir, 0644)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#3 Failed to create write directory: %v", err))
		return err
	}

	return nil
}

func (fs *FileStorage) WalkDirectories() []string {
	// Walk all files in HDFS root directory and all subdirectories
	var paths []string
	callbackFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			log.Info(fmt.Sprintf("[acco-hdfs]acfs#4 Directory: %s\n", path))
			path = fmt.Sprintf("Directory: %s\n", path)
			paths = append(paths, path)
		} else {
			log.Info(fmt.Sprintf("[acco-hdfs]acfs#5 File: %s\n", path))
			path = fmt.Sprintf("File: %s\n", path)
			paths = append(paths, path)
		}
		return nil
	}
	fs.client.Walk(hdfsRoot, callbackFunc)
	return paths
}

func (fs *FileStorage) CopyLocalFile(localFilePath, fileName string) error {
	// Create local file
	file, err := os.Create(localFilePath)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#6 Failed to create local file: %v", err))
		return err
	}
	fileContent := "Hello World!"
	_, err = file.WriteString(fileContent)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#7 Failed to write to local file: %v", err))
		return err
	}
	file.Close()

	// Copy file to HDFS
	_ = fs.client.CopyToRemote(localFilePath, hdfsCopyDir+fileName)
	return nil
}

func (fs *FileStorage) WriteFile(fileContent string, fileName string) error {
	filePath := hdfsWriteDir + fileName

	// Create file on HDFS with default replication and block size
	file, err := fs.client.Create(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#8 Failed to create file on HDFS: %v", err))
		return err
	}

	// Write content
	// Create byte array from string file content
	fileContentByteArray := []byte(fileContent)

	// IMPORTANT: writes content to local buffer, content is pushed to HDFS only when Close is called!
	_, err = file.Write(fileContentByteArray)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#9 Failed to write to file on HDFS: %v", err))
		return err
	}

	// CLOSE FILE WHEN ALL WRITING IS DONE
	// Ensuring all changes are flushed to HDFS
	// defer file.Close() can be used at the begining of the method to ensure closing is not forgotten
	_ = file.Close()
	return nil
}

func (fs *FileStorage) ReadFile(fileName string, isCopied bool) (string, error) {
	var filePath string
	if isCopied {
		filePath = hdfsCopyDir + fileName
	} else {
		filePath = hdfsWriteDir + fileName
	}

	// Open file for reading
	file, err := fs.client.Open(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#10 Failed to open file for reading on HDFS: %v", err))
		return "", err
	}

	// Read file content
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#11 Failed to read file on HDFS: %v", err))
		return "", err
	}

	// Convert number of read bytes into string
	fileContent := string(buffer[:n])
	return fileContent, nil
}

// Writing bytes to a file. Used for saving images in HDFS.
// Returns error if it fails.
func (fs *FileStorage) WriteFileBytes(imageData []byte, fileName string) error {
	filePath := hdfsWriteDir + fileName

	// Create file on HDFS with default replication and block size
	file, err := fs.client.Create(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#12 Failed to create byte file on HDFS: %v", err))
		return err
	}

	// Write image data
	_, err = file.Write(imageData)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#13 Failed to write to byte file on HDFS: %v", err))
		return err
	}

	// Close file to ensure all changes are flushed to HDFS
	err = file.Close()
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#14 Failed to close byte file on HDFS: %v", err))
		return err
	}

	return nil
}

// TODO NoSQL: add method that returns file content as byte array when content is not human readable (images, video,...)

// Reading bytes from a file. Used for getting images in HDFS.
// Returns ([]byte, nil) if successful, otherwise (nil, error).
func (fs *FileStorage) ReadFileBytes(fileName string, isCopied bool) ([]byte, error) {
	var filePath string
	if isCopied {
		filePath = hdfsCopyDir + fileName
	} else {
		filePath = hdfsWriteDir + fileName
	}

	// Open file for reading
	file, err := fs.client.Open(filePath)
	if err != nil {
		log.Warning(fmt.Sprintf("[acco-hdfs]acfs#15 Failed to open byte file for reading on HDFS: %v", err))
		return nil, err
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#16 Failed to read byte file on HDFS: %v", err))
		return nil, err
	}

	return fileContent, nil
}

func (fs *FileStorage) DeleteFile(fileName string, isCopied bool) error {
	var filePath string
	if isCopied {
		filePath = hdfsCopyDir + fileName
	} else {
		filePath = hdfsWriteDir + fileName
	}

	err := fs.client.Remove(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-hdfs]acfs#17 Failed to delete file on HDFS: %v", err))
		return err
	}

	return nil
}
