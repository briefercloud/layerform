package pathutils

import (
	"os"
	"path/filepath"
	"strings"
)

func FindCommonParentPath(filePaths []string) string {
	// Split the first file path into directories.
	parentDirs := strings.Split(filepath.Dir(filePaths[0]), string(os.PathSeparator))

	// Iterate through the remaining file paths and find the common parent directory.
	for _, filePath := range filePaths[1:] {
		// Split the current file path into directories.
		currDirs := strings.Split(filepath.Dir(filePath), string(os.PathSeparator))

		// Update the parentDirs to contain only the common directories so far.
		for i := 0; i < len(parentDirs) && i < len(currDirs); i++ {
			if parentDirs[i] != currDirs[i] {
				parentDirs = parentDirs[:i]
				break
			}
		}
	}

	// Join the common parent directories back into a path.
	commonParent := strings.Join(parentDirs, string(os.PathSeparator))

	return commonParent
}
