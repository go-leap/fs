package ufs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-leap/str"
)

var (
	// CreateModePerm is used by all functions in this package that create file-system directories or files, namely: `EnsureDir`, `WriteBinaryFile`, `WriteTextFile`.
	CreateModePerm = os.ModePerm

	// Del aliases `os.RemoveAll` — merely a handy short-hand during rapid iteration in non-critical code-paths that already do import `ufs` to not have to repeatedly pull in and out the extra `os` import.
	Del = os.RemoveAll
)

// AllFilePathsIn collects the full paths of all files directly or indirectly contained under `dirPath`.
func AllFilePathsIn(dirPath string, ignoreSubPath string) (allFilePaths []string) {
	if ignoreSubPath != "" && !ustr.Pref(ignoreSubPath, dirPath) {
		ignoreSubPath = filepath.Join(dirPath, ignoreSubPath)
	}
	WalkAllFiles(dirPath, func(curfilepath string) (keepwalking bool) {
		if !ustr.Pref(curfilepath, ignoreSubPath) {
			allFilePaths = append(allFilePaths, curfilepath)
		}
		return true
	})
	return
}

// ClearDir removes everything inside `dirPath`, but not `dirPath` itself and also excepting items inside `dirPath` (but not inside sub-directories) with one of the specified `keepNames`.
func ClearDir(dirPath string, keepNames ...string) (err error) {
	if IsDir(dirPath) {
		var fileInfos []os.FileInfo
		if fileInfos, err = ioutil.ReadDir(dirPath); err == nil {
			for _, fi := range fileInfos {
				if fn := fi.Name(); !ustr.In(fn, keepNames...) {
					if err = os.RemoveAll(filepath.Join(dirPath, fn)); err != nil {
						return
					}
				}
			}
		}
	}
	return
}

// CopyFile attempts an `io.Copy` from `srcFilePath` to `dstFilePath`.
func CopyFile(srcFilePath, dstFilePath string) (err error) {
	var src *os.File
	if src, err = os.Open(srcFilePath); src != nil {
		if err == nil {
			err = SaveTo(src, dstFilePath)
		}
		_ = src.Close()
	}
	return
}

// CopyAllFilesAndSubDirs copies all files and directories inside `srcDirPath` into `dstDirPath`.
// All sub-directories whose `os.FileInfo.Name` is contained in `skipDirNames` (if supplied)
// are skipped, and so are files with names ending in `skipFileSuffix` (if supplied).
func CopyAllFilesAndSubDirs(srcDirPath, dstDirPath string, skipFileSuffix string, skipDirNames ...string) (err error) {
	var fileInfos []os.FileInfo
	if fileInfos, err = ioutil.ReadDir(srcDirPath); err == nil {
		if err = EnsureDir(dstDirPath); err == nil {
			for _, fi := range fileInfos {
				fname := fi.Name()
				if srcPath, dstPath := filepath.Join(srcDirPath, fname), filepath.Join(dstDirPath, fname); fi.IsDir() {
					if !ustr.In(fname, skipDirNames...) {
						err = CopyAllFilesAndSubDirs(srcPath, dstPath, skipFileSuffix, skipDirNames...)
					}
				} else if skipFileSuffix == "" || !ustr.Suff(srcPath, skipFileSuffix) {
					err = CopyFile(srcPath, dstPath)
				}
				if err != nil {
					break
				}
			}
		}
	}
	return
}

// EnsureDir attempts to create the directory `dirPath` if it does not yet exist.
func EnsureDir(dirPath string) (err error) {
	if !IsDir(dirPath) {
		if err = EnsureDir(filepath.Dir(dirPath)); err == nil {
			err = os.Mkdir(dirPath, CreateModePerm)
		}
	}
	return
}

// IsAnyFileInDirNewerThanTheOldestOf returns whether any file directly or indirectly contained in `dirPath` is newer than the oldest of the specified `filePaths`.
func IsAnyFileInDirNewerThanTheOldestOf(dirPath string, filePaths ...string) (isAnyNewer bool) {
	var cmpfiletimeoldest int64
	if len(filePaths) == 0 {
		return true
	}
	for _, fp := range filePaths {
		if cmpfile, err := os.Stat(fp); err != nil || cmpfile == nil {
			return true
		} else if modtime := cmpfile.ModTime().UnixNano(); modtime > 0 && (cmpfiletimeoldest == 0 || modtime < cmpfiletimeoldest) {
			cmpfiletimeoldest = modtime
		}
	}
	if err := WalkAllFiles(dirPath, func(curfilepath string) (keepwalking bool) {
		if !ustr.In(curfilepath, filePaths...) {
			if curfile, errstat := os.Stat(curfilepath); errstat != nil || curfile == nil || curfile.ModTime().UnixNano() > cmpfiletimeoldest {
				isAnyNewer = true
			}
		}
		return !isAnyNewer
	}); err != nil {
		return true
	}
	return
}

// IsDir returns whether a directory (not a file) exists at the specified `fsPath`.
func IsDir(fsPath string) bool {
	if len(fsPath) == 0 {
		return false
	}
	stat, err := os.Stat(fsPath)
	return err == nil && stat.Mode().IsDir()
}

// IsFile returns whether a file (not a directory) exists at the specified `fsPath`.
func IsFile(fsPath string) bool {
	if len(fsPath) == 0 {
		return false
	}
	stat, err := os.Stat(fsPath)
	return err == nil && stat.Mode().IsRegular()
}

// ReadTextFile is a `string`-typed convenience short-hand for `ioutil.ReadFile`.
func ReadTextFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	return string(data), err
}

// ReadTextFileOr calls `ReadTextFile(filePath)` but returns `fallback` on `error`.
func ReadTextFileOr(filePath string, fallback string) string {
	src, err := ReadTextFile(filePath)
	if err != nil {
		src = fallback
	}
	return src
}

// ReadTextFileOrPanic calls `ReadTextFile(filePath)` but `panic`s on `error`.
func ReadTextFileOrPanic(filePath string) string {
	src, err := ReadTextFile(filePath)
	if err != nil {
		panic(err)
	}
	return src
}

// SaveTo attempts an `io.Copy` from `src` to `dstFilePath`.
func SaveTo(src io.Reader, dstFilePath string) (err error) {
	var file *os.File
	if file, err = os.Create(dstFilePath); file != nil {
		if err == nil {
			_, err = io.Copy(file, src)
		}
		_ = file.Close()
	}
	return
}

func walk(dirPath string, self bool, traverse bool, onDir func(string) bool, onFile func(string) bool) (keepWalking bool, err error) {
	dodirs, dofiles := onDir != nil, onFile != nil
	if keepWalking = true; self && dodirs {
		keepWalking = onDir(dirPath)
	}
	if keepWalking {
		var fileInfos []os.FileInfo
		if fileInfos, err = ioutil.ReadDir(dirPath); err == nil {
			for _, fi := range fileInfos {
				if fspath := filepath.Join(dirPath, fi.Name()); fi.Mode().IsRegular() && dofiles {
					keepWalking = onFile(fspath)
				} else if fi.Mode().IsDir() {
					if dodirs {
						keepWalking = onDir(fspath)
					}
					if keepWalking && traverse {
						keepWalking, err = walk(fspath, false, true, onDir, onFile)
					}
				}
				if err != nil || !keepWalking {
					break
				}
			}
		}
	}
	return
}

func Walk(dirPath string, self bool, traverse bool, onDir func(string) bool, onFile func(string) bool) (err error) {
	if IsDir(dirPath) {
		_, err = walk(dirPath, self, traverse, onDir, onFile)
	}
	return
}

func WalkAllFiles(dirPath string, onFile func(string) bool) error {
	return Walk(dirPath, false, true, nil, onFile)
}

func WalkDirsIn(dirPath string, onDir func(string) bool) error {
	return Walk(dirPath, false, false, onDir, nil)
}

func WalkFilesIn(dirPath string, onFile func(string) bool) error {
	return Walk(dirPath, false, false, nil, onFile)
}

// WriteBinaryFile is a convenience short-hand for `ioutil.WriteFile` that also `EnsureDir`s the destination.
func WriteBinaryFile(filePath string, contents []byte) error {
	_ = EnsureDir(filepath.Dir(filePath))
	return ioutil.WriteFile(filePath, contents, CreateModePerm)
}

// WriteTextFile is a `string`-typed convenience short-hand for `ioutil.WriteFile` that also `EnsureDir`s the destination.
func WriteTextFile(filePath, contents string) error {
	return WriteBinaryFile(filePath, []byte(contents))
}
