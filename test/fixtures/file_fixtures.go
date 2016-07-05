package fixtures

import (
	"github.com/docker/docker/pkg/ioutils"
	"path"
	"os"
)

// FileFixtures writes files to a temporary location for use in testing.
type FileFixtures struct {
	tempDir string
	// If an error has occurred setting up fixtures
	Error   error
}

// Set up a new FileFixtures object by passing an interlaced list of file names
// and file contents. The file names will be interpreted as relative to some
// temporary root directory that is fixed when allocate() is called on the
// FileFixtures struct.
func NewFileFixtures() *FileFixtures {
	dir, err := ioutils.TempDir("", "FileFixtures")
	return &FileFixtures{
		tempDir: dir,
		Error: err,
	}
}

// Add a file relative to the FileFixtures tempDir using name for the relative
// part of the path.
func (ffs *FileFixtures) AddFile(name, content string) string {
	if ffs.Error != nil {
		return ""
	}
	filePath := path.Join(ffs.tempDir, name)
	ffs.AddDir(path.Dir(name))
	if (ffs.Error == nil) {
		ffs.Error = createWriteClose(filePath, content)
	}
	return filePath
}

// Ensure that the directory relative to the FileFixtures tempDir exists using
// name for the relative part of the path.
func (ffs *FileFixtures) AddDir(name string) string {
	if ffs.Error != nil {
		return ""
	}
	filePath := path.Join(ffs.tempDir, name)
	ffs.Error = os.MkdirAll(filePath, 0777)
	return filePath
}


// Cleans up the the temporary files (with fire)
func (ffs *FileFixtures) RemoveAll() {
	if err := os.RemoveAll(ffs.tempDir); err != nil {
		// Since we expect to be called from being deferred in a test it's
		// better if we panic here so that the caller finds
		panic(err)
	}
}

// Create a text file at filename with contents content
func createWriteClose(filename, content string) error {
	// We'll create any parent dirs, with permissive permissions
	err := os.MkdirAll(path.Dir(filename), 0777)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	f.WriteString(content)
	defer f.Close()
	return nil
}

