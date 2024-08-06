package fshelpers

import (
	"fmt"
	"io"
	"os"
)

func MoveFile(oldpath string, newpath string) error {
	srcFile, err := os.Open(oldpath)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}

	destFile, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("could not open destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	srcFile.Close()
	if err != nil {
		return fmt.Errorf("could not write to destination file: %v", err)
	}

	// Remove original file
	if err := os.Remove(oldpath); err != nil {
		return fmt.Errorf("could not remove original file: %v", err)
	}

	return nil
}
