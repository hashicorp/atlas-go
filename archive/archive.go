// archive is package that helps create archives in a format that
// Harmony expects with its various upload endpoints.
package archive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveOpts are the options for defining how the archive will be built.
type ArchiveOpts struct {
	Exclude []string
	Include []string
	VCS     bool
}

// IsSet says whether any options were set.
func (o *ArchiveOpts) IsSet() bool {
	return len(o.Exclude) > 0 || len(o.Include) > 0 || o.VCS
}

// Archive takes the given path and ArchiveOpts and archives it.
//
// The archive is done async and streamed via the io.ReadCloser returned.
// The reader is blocking: data is only compressed and written as data is
// being read from the reader. Because of this, any user doesn't have to
// worry about quickly reading data to avoid memory bloat.
//
// The archive can be read with the io.ReadCloser that is returned. The error
// returned is an error that happened before archiving started, so the
// ReadCloser doesn't need to be closed (and should be nil). The error
// channel are errors that can happen while archiving is happening. When
// an error occurs on the channel, reading should stop and be closed.
func Archive(
	path string, opts *ArchiveOpts) (io.ReadCloser, <-chan error, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	// Direct file paths cannot have archive options
	if !fi.IsDir() && opts.IsSet() {
		return nil, nil, fmt.Errorf(
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
	path string, opts *ArchiveOpts) (io.ReadCloser, <-chan error, error) {
	// TODO: if file is already gzipped, then send it along
	// TODO: if file is not gzipped, then... error? or do we tar + gzip?

	return nil, nil, nil
}

func archiveDir(
	root string, opts *ArchiveOpts) (io.ReadCloser, <-chan error, error) {
	var vcsInclude []string
	if opts.VCS {
		var err error
		vcsInclude, err = vcsFiles(root)
		if err != nil {
			return nil, nil, err
		}
	}

	// We're going to write to an io.Pipe so that we can ensure the other
	// side is reading as we're writing.
	pr, pw := io.Pipe()

	// Buffer the writer so that we can keep some data moving in memory
	// while we're compressing. 4M should be good.
	bufW := bufio.NewWriterSize(pw, 4096*1024)

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

	// Create all our channels so we can send data through some tubes
	// to other goroutines.
	errCh := make(chan error, 1)
	go func() {
		werr := filepath.Walk(root, walkFn)

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

		// Close the pipe
		if err := pw.Close(); err != nil && werr == nil {
			werr = err
		}

		// Send any error we might have down the pipe if we have one
		if werr != nil {
			errCh <- werr
		}
	}()

	return pr, errCh, nil
}
