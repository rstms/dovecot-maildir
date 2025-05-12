/*
Copyright © 2025 Matt Krueger <mkrueger@rstms.net>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package cmd

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func IsCompressed(pathName string) (bool, error) {
	file, err := os.Open(pathName)
	if err != nil {
		return false, err
	}
	defer file.Close()
	cmpType, err := DetectCompressedFile(file)
	if err != nil {
		return false, fmt.Errorf("DetectCompressedFile: %v", err)
	}
	if cmpType == nil {
		return false, nil
	}
	return true, nil
}

func UncompressFile(pathName string) error {

	verbose := viper.GetBool("verbose")
	debug := viper.GetBool("debug")

	stat, err := os.Stat(pathName)
	if err != nil {
		return fmt.Errorf("failed stat on compressed file: %v", err)
	}

	var compressionType *string
	decoded, err := func() ([]byte, error) {
		var decoded []byte
		file, err := os.Open(pathName)
		if err != nil {
			return decoded, fmt.Errorf("failed opening compressed file: %v", err)
		}
		defer file.Close()

		compressionType, err = DetectCompressedFile(file)
		if err != nil {
			return decoded, fmt.Errorf("DetectCompressedFile: %v", err)
		}

		switch *compressionType {
		case "zstd":
			decoded, err = decompressZstd(file)
		case "gzip":
			decoded, err = decompressGzip(file)
		case "bzip2":
			decoded, err = decompressBzip2(file)
		default:
			err = fmt.Errorf("unknown compression type: %s", pathName)
		}
		if err != nil {
			return decoded, err
		}
		return decoded, nil

	}()
	if err != nil {
		return err
	}

	var lineCount int64
	scanner := bufio.NewScanner(bytes.NewBuffer(decoded))
	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			fmt.Printf("%s\n", line)
		}
		lineCount += 1
	}
	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("failed reading compressed lines: %v", err)
	}

	name, flags, found := strings.Cut(pathName, ":")
	if !found {
		return fmt.Errorf("missing ':' in filename: %s", pathName)
	}
	if !strings.HasPrefix(flags, "2,") {
		return fmt.Errorf("missing '2,' in filename: %s", pathName)
	}
	parts := strings.Split(name, ",")

	var nameSize int64
	var nameSizeW int64
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "S=") {
			_, numStr, _ := strings.Cut(part, "=")
			numVal, _ := strconv.Atoi(numStr)
			nameSize = int64(numVal)
		}
		if strings.HasPrefix(part, "W=") {
			_, numStr, _ := strings.Cut(part, "=")
			numVal, _ := strconv.Atoi(numStr)
			nameSizeW = int64(numVal)
		}
	}

	size := int64(len(decoded))
	sizeW := int64(size + lineCount)

	if verbose {
		log.Printf("inFile=%s\n", pathName)
		log.Printf("type=%s\n", *compressionType)
		log.Printf("flags=%s\n", flags)
		log.Printf("size=%v\n", size)
		log.Printf("sizeW=%v\n", sizeW)
		log.Printf("nameSize=%v\n", nameSize)
		log.Printf("nameSizeW=%v\n", nameSizeW)
	}

	if nameSize > 0 {
		if nameSize != size {
			return fmt.Errorf("uncompressed S=%d mismatches filename S=value: %s", size, pathName)
		}
	}

	if nameSizeW > 0 {
		if nameSizeW != sizeW {
			return fmt.Errorf("uncompressed W=%d mismatches filename W=value: %s", sizeW, pathName)
		}
	}

	err = os.WriteFile(pathName, decoded, 0600)
	if err != nil {
		return fmt.Errorf("failed writing decoded data to %s: %v", pathName, err)
	}

	err = SetStat(pathName, stat)
	if err != nil {
		return err
	}

	return nil
}

func decompressZstd(file io.Reader) ([]byte, error) {
	var decoded []byte
	decoder, err := zstd.NewReader(file)
	if err != nil {
		return decoded, fmt.Errorf("failed creating zstandard decoder: %v", err)
	}
	decoded, err = io.ReadAll(decoder.IOReadCloser())
	if err != nil {
		return decoded, fmt.Errorf("failed writing zstandard uncompressed data: %v", err)
	}
	return decoded, nil
}

func decompressGzip(file io.Reader) ([]byte, error) {
	var decoded []byte
	decoder, err := gzip.NewReader(file)
	if err != nil {
		return decoded, fmt.Errorf("failed creating gzip decoder: %v", err)
	}
	defer decoder.Close()
	decoded, err = ioutil.ReadAll(decoder)
	if err != nil {
		return decoded, fmt.Errorf("failed writing gzip uncompressed data: %v", err)
	}
	return decoded, nil
}

func decompressBzip2(file io.Reader) ([]byte, error) {
	var decoded []byte
	decoder := bzip2.NewReader(file)
	decoded, err := ioutil.ReadAll(decoder)
	if err != nil {
		return decoded, fmt.Errorf("failed writing bzip2 uncompressed data: %v", err)
	}
	return decoded, nil
}

func IsMaildir(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("Stat failed: %v", err)
	}
	if !stat.IsDir() {
		return false, fmt.Errorf("not a directory: %s", dir)
	}
	stat, err = os.Stat(filepath.Join(dir, "cur"))
	return err == nil && stat.IsDir(), nil
}

func ListMaildirs(dir string) (*[]string, error) {
	maildir, err := IsMaildir(dir)
	if err != nil {
		return nil, err
	}
	if !maildir {
		return nil, fmt.Errorf("not a maildir: %s", dir)
	}
	if !viper.GetBool("recurse") {
		return &[]string{dir}, nil
	}
	mailDirs := []string{}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			maildir, err := IsMaildir(path)
			if err != nil {
				return err
			}
			if maildir {
				mailDirs = append(mailDirs, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir failed: %v", err)
	}
	return &mailDirs, nil
}

func ListMaildirFiles(dir string) (*[]string, error) {

	listUncompressed := viper.GetBool("uncompressed")
	listAll := viper.GetBool("all")
	debug := viper.GetBool("debug")

	stat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("Stat failed: %v", err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	path := filepath.Join(dir, "cur")
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("ReadDir failed: %v", err)
	}
	filenames := []string{}
	count := 0
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		pathName := filepath.Join(path, entry.Name())
		isCompressed, err := IsCompressed(pathName)
		if err != nil {
			return nil, err
		}
		if !listAll {
			if listUncompressed {
				if isCompressed {
					continue
				}
			} else {
				if !isCompressed {
					continue
				}
			}
		}
		filenames = append(filenames, pathName)
		count += 1
		if debug {
			fmt.Printf("%d %v %s\n", count, isCompressed, pathName)
		}
	}
	return &filenames, nil
}

func SetStat(path string, info fs.FileInfo) error {

	// replicate access mode bits
	err := os.Chmod(path, info.Mode())
	if err != nil {
		return fmt.Errorf("mode change failed on '%s': %v", path, err)
	}

	// replicate modification time
	err = os.Chtimes(path, time.Now(), info.ModTime())
	if err != nil {
		return fmt.Errorf("mod time change failed on '%s': %v", path, err)
	}

	// replicate ownership
	uid := info.Sys().(*syscall.Stat_t).Uid
	gid := info.Sys().(*syscall.Stat_t).Gid
	err = os.Chown(path, int(uid), int(gid))
	if err != nil {
		return fmt.Errorf("ownership change failed on '%s': %v", path, err)
	}

	return nil
}
