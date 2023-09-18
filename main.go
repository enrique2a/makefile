package main

import (
	"flag"
	"fmt"
	"math/rand"

	"os/user"

	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	isRoot, err := isRoot()
	if err != nil {
		fmt.Println("Error to get the operating user information:", err)
		return
	}

	if isRoot {
		fmt.Println("This program should not be run as root.")
		return
	}

	//parse the arguments
	var (
		overwriteFile bool
		outputFile    string
		fileSize      string
		verbose       bool
	)

	flag.BoolVar(&overwriteFile, "o", false, "overwrite the file if exist")
	flag.BoolVar(&verbose, "verbose", true, "overwrite the file if exist")
	flag.StringVar(&outputFile, "of", "", "output file name")
	flag.StringVar(&fileSize, "s", "", "size of the file to generate. Use K kilobytes, M megabytes, G gigabytes")
	flag.Parse()

	if outputFile == "" {
		fmt.Println("Error: output file name not specified, use -of option to specify it")
		return
	}

	if fileSize == "" {
		fmt.Println("Error: file size not specified, use -s option to specify it")
		return
	}

	//blockSizeArg := args[2]
	size, err := parseSize(fileSize)
	if err != nil {
		fmt.Println("Error parsing file size:", err)

		return
	}

	blockSizeArg := "4K" //default block size for many filesystems

	blockSize, err := parseSize(blockSizeArg)
	if err != nil {
		fmt.Println("Error parsing block size:", err)
		return
	}

	//check if the file already exists
	exist, _, err := fileExists(outputFile)
	if err != nil {
		fmt.Println("Error to check if file", outputFile, "exists:", err)
		return
	}

	if exist && !overwriteFile {
		fmt.Println("the file already exists, use -o (overwrite) to overwrite")
		return
	}

	//create the file in current directory

	outputFile = filepath.Base(outputFile)

	if err := generateFileWithBlocks(outputFile, size, blockSize, verbose); err != nil {
		fmt.Println("Error generating file:", err)
		return
	}

	fmt.Printf("File \"%s\" successfully generated with a size of %d bytes.\n", outputFile, size)
}

func generateFileWithBlocks(fileName string, fileSize int, blockSize int, verbose bool) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Initializes the processed bytes counter
	bytesProcessed := int64(0)

	for remainingSize := fileSize; remainingSize > 0; {

		// Calculate the size for the current block
		blockSizeToWrite := blockSize
		if blockSizeToWrite > remainingSize {
			blockSizeToWrite = remainingSize
		}

		// Generate a block of random characters
		block := make([]byte, blockSizeToWrite)
		for i := 0; i < blockSizeToWrite; i++ {
			block[i] = alphabet[rand.Intn(len(alphabet))]
		}

		// Write the block to the file
		_, err := file.Write(block)
		if err != nil {
			return err
		}

		if verbose {

			progress := float64(bytesProcessed) / float64(fileSize) * 100

			isWhole := int(progress*100) == int(progress)*100

			if isWhole {
				fmt.Printf("Progress: %.2f%%\r", progress)
			}

			bytesProcessed += int64(blockSizeToWrite)
		}

		remainingSize -= blockSizeToWrite

	}

	return nil
}

func parseSize(sizeArg string) (int, error) {
	sizeArg = strings.ToUpper(sizeArg)

	multipliers := map[string]int{
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
	}

	if len(sizeArg) < 2 {
		return 0, fmt.Errorf("invalid size format")
	}

	suffix := sizeArg[len(sizeArg)-1:]
	if _, ok := multipliers[suffix]; !ok {
		return 0, fmt.Errorf("invalid size suffix: %s", suffix)
	}

	sizeStr := sizeArg[:len(sizeArg)-1]
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size, expected integer number. Provided: %s", sizeStr)
	}

	return size * multipliers[suffix], nil
}

func help() {

	binaryPath := os.Args[0]
	binaryName := filepath.Base(binaryPath)

	fmt.Println("Usage:", binaryName, "[OPTION]... -f filename -s size")
	fmt.Println("Size format, K for Kilobytes, M for Megabytes, G for Gigabytes")
	fmt.Println("Example, generate 1 Gb file and overwrite the output file:", binaryName, "-o myfile1Gb.txt 1G")

	fmt.Println("")
	fmt.Println("OPTIONS:")
	flag.Usage()

}

func isRoot() (bool, error) {
	currentUser, err := user.Current()
	if err != nil {
		return true, err
	}

	if currentUser.Username == "root" || currentUser.Uid == "0" {
		return true, nil
	}
	return false, nil
}

func dirExists(path string) (exist bool, notExist bool, err error) {
	// delete the space and normalized the path of the file
	path = strings.TrimSpace(path)
	path = filepath.Clean(path)

	//extract the direcotry from path
	path = filepath.Dir(path)

	// check if the file already exists
	dirInfo, err := os.Stat(path)

	// really does not exist
	if os.IsNotExist(err) {
		return false, true, nil
	} else if err != nil {
		// it was an error to check if dir exists
		return false, false, err
	} else {
		if dirInfo.IsDir() {
			return true, false, nil
		} else {
			return false, false, nil
		}
	}
}

func fileExists(filePath string) (exist bool, notExist bool, err error) {
	// delete the space and normalized the path of the file
	filePath = strings.TrimSpace(filePath)
	filePath = filepath.Clean(filePath)

	// check if the file already exists

	_, err = os.Stat(filePath)

	// really does not exist
	if os.IsNotExist(err) {
		return false, true, nil
	} else if err != nil {
		// it was an error to check if the file exists
		return false, false, err
	} else {
		return true, false, nil
	}
}
