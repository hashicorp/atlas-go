// archive is package that helps create archives in a format that
// Atlas expects with its various upload endpoints.
package archive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveOpts are the options for defining how the archive will be built.
type ArchiveOpts struct {
	// Exclude and Include are filters of files to include/exclude in
	// the archive when creating it from a directory. These filters should
	// be relative to the packaging directory and should be basic glob
	// patterns.
	Exclude []string
	Include []string

	// Extra is a mapping of extra files to include within the archive. The
	// key should be the path within the archive and the value should be
	// an absolute path to the file to put into the archive. These extra
	// files will override any other files in the archive.
	Extra map[string]string

	// VCS, if true, will detect and use a VCS system to determien what
	// files to include the archive.
	VCS bool
}

// IsSet says whether any options were set.
func (o *ArchiveOpts) IsSet() bool {
	return len(o.Exclude) > 0 || len(o.Include) > 0 || o.VCS
}

// Archive takes the given path and ArchiveOpts and archives it.
//
// The archive will be fully completed and put into a temporary file.
// This must be done to retrieve the content length of the archive which
// is need for almost all operations involving archives with Atlas. Because
// of this, sufficient disk space will be required to buffer the archive.
func Archive(
	path string, opts *ArchiveOpts) (io.ReadCloser, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	// Direct file paths cannot have archive options
	if !fi.IsDir() && opts.IsSet() {
		return nil, 0, fmt.Errorf(
			"Options such as exclude, include, and VCS can't be set when " +
				"the path is a file.")
	}

	if fi.IsDir() {
		return archiveDir(path, opts)
	} else {
		return archiveFile(path, opts)
	}
}

func archiveFile(
	path string, opts *ArchiveOpts) (io.ReadCloser, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	if _, err := gzip.NewReader(f); err == nil {
		// Reset the read offset for future reading
		if _, err := f.Seek(0, 0); err != nil {
			f.Close()
			return nil, 0, err
		}

		// Get the file info for the size
		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, 0, err
		}

		// This is a gzip file, let it through.
		return f, fi.Size(), nil
	}

	// Close the file, no use for it anymore
	f.Close()

	// We have a single file that is not gzipped. Compress it.
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Act like we're compressing a directory, but only include this one
	// file.
	return archiveDir(filepath.Dir(path), &ArchiveOpts{
		Include: []string{filepath.Base(path)},
	})
}

func archiveDir(
	root string, opts *ArchiveOpts) (io.ReadCloser, int64, error) {
	var vcsInclude []string
	if opts.VCS {
		var err error
		vcsInclude, err = vcsFiles(root)
		if err != nil {
			return nil, 0, err
		}
	}

	// Create the temporary file that we'll send the archive data to.
	archiveF, err := ioutil.TempFile("", "atlas-archive")
	if err != nil {
		return nil, 0, err
	}

	// Create the wrapper for the result which will automatically
	// remove the temporary file on close.
	archiveWrapper := &readCloseRemover{F: archiveF}

	// Buffer the writer so that we can push as much data to disk at
	// a time as possible. 4M should be good.
	bufW := bufio.NewWriterSize(archiveF, 4096*1024)

	// Gzip compress all the output data
	gzipW := gzip.NewWriter(bufW)

	// Tar the file contents
	tarW := tar.NewWriter(gzipW)

	// Build the function that'll do all the compression
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the path since it contains the root
		// plus the path.
		subpath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if subpath == "." {
			return nil
		}

		// If we have a list of VCS files, check that first
		skip := false
		if len(vcsInclude) > 0 {
			skip = true
			for _, f := range vcsInclude {
				if f == subpath {
					skip = false
					break
				}

				if info.IsDir() && strings.HasPrefix(f, subpath+"/") {
					skip = false
					break
				}
			}
		}

		// If include is present, we only include what is listed
		if len(opts.Include) > 0 {
			skip = true
			for _, include := range opts.Include {
				match, err := filepath.Match(include, subpath)
				if err != nil {
					return err
				}
				if match {
					skip = false
					break
				}
			}
		}

		// If exclude, it is one last gate to excluding files
		for _, exclude := range opts.Exclude {
			match, err := filepath.Match(exclude, subpath)
			if err != nil {
				return err
			}
			if match {
				skip = true
				break
			}
		}

		// If we have to skip this file, then skip it, properly skipping
		// children if we're a directory.
		if skip {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Read the symlink target. We don't track the error because
		// it doesn't matter if there is an error.
		target, _ := os.Readlink(path)

		// Build the file header for the tar entry
		header, err := tar.FileInfoHeader(info, target)
		if err != nil {
			return fmt.Errorf(
				"Failed creating archive header: %s", path)
		}

		// Modify the header to properly be the full subpath
		header.Name = subpath
		if info.IsDir() {
			header.Name += "/"
		}

		// Write the header first to the archive.
		if err := tarW.WriteHeader(header); err != nil {
			return fmt.Errorf(
				"Failed writing archive header: %s", path)
		}

		// If it is a directory, then we're done (no body to write)
		if info.IsDir() {
			return nil
		}

		// Open the target file to write the data
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf(
				"Failed opening file '%s' to write compressed archive.", path)
		}
		defer f.Close()

		if _, err = io.Copy(tarW, f); err != nil {
			return fmt.Errorf(
				"Failed copying file to archive: %s", path)
		}

		return nil
	}

	// First, walk the path and do the normal files
	werr := filepath.Walk(root, walkFn)
	if werr == nil {
		// If that succeeded, handle the extra files
		werr = copyExtras(tarW, opts.Extra)
	}

	// Attempt to close all the things. If we get an error on the way
	// and we haven't had an error yet, then record that as the critical
	// error. But we still try to close everything.

	// Close the tar writer
	if err := tarW.Close(); err != nil && werr == nil {
		werr = err
	}

	// Close the gzip writer
	if err := gzipW.Close(); err != nil && werr == nil {
		werr = err
	}

	// Flush the buffer
	if err := bufW.Flush(); err != nil && werr == nil {
		werr = err
	}

	// If we had an error, then close the file (removing it) and
	// return the error.
	if werr != nil {
		archiveWrapper.Close()
		return nil, 0, werr
	}

	// Seek to the beginning
	if _, err := archiveWrapper.F.Seek(0, 0); err != nil {
		archiveWrapper.Close()
		return nil, 0, err
	}

	// Get the file information so we can get the size
	fi, err := archiveWrapper.F.Stat()
	if err != nil {
		archiveWrapper.Close()
		return nil, 0, err
	}

	return archiveWrapper, fi.Size(), nil
}

func copyExtras(w *tar.Writer, extra map[string]string) error {
	for entry, path := range extra {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// Read the symlink target. We don't track the error because
		// it doesn't matter if there is an error.
		target, _ := os.Readlink(path)

		// Build the file header for the tar entry
		header, err := tar.FileInfoHeader(info, target)
		if err != nil {
			return fmt.Errorf(
				"Failed creating archive header: %s", path)
		}

		// Modify the header to properly be the full subpath
		header.Name = entry

		// Write the header first to the archive.
		if err := w.WriteHeader(header); err != nil {
			return fmt.Errorf(
				"Failed writing archive header: %s", path)
		}

		// If it is a directory, then we're done (no body to write)
		if info.IsDir() {
			return nil
		}

		// Open the target file to write the data
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf(
				"Failed opening file '%s' to write compressed archive.", path)
		}

		_, err = io.Copy(w, f)
		f.Close()
		if err != nil {
			return fmt.Errorf(
				"Failed copying file to archive: %s", path)
		}
	}

	return nil
}

// readCloseRemover is an io.ReadCloser implementation that will remove
// the file on Close(). We use this to clean up our temporary file for
// the archive.
type readCloseRemover struct {
	F *os.File
}

func (r *readCloseRemover) Read(p []byte) (int, error) {
	return r.F.Read(p)
}

func (r *readCloseRemover) Close() error {
	// First close the file
	err := r.F.Close()

	// Next make sure to remove it, or at least try, regardless of error
	// above.
	os.Remove(r.F.Name())

	return err
}
