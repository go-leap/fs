package ufs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-leap/str"
)

var (
	// CreateModePerm (rwx for user,group,other) is used by all functions in this package that create file-system directories or files, namely: `EnsureDir`, `WriteBinaryFile`, `WriteTextFile`.
	CreateModePerm = os.ModePerm

	// Del aliases `os.RemoveAll` — merely a handy short-hand during rapid iteration in non-critical code-paths that already do import `ufs` to not have to repeatedly pull in and out the extra `os` import.
	Del = os.RemoveAll

	WalkReadDirFunc       = ioutil.ReadDir
	WalkIgnoreReadDirErrs bool
)

// AllFilePathsIn collects the full paths of all files directly or indirectly contained under `dirPath`.
func AllFilePathsIn(dirPath string, ignoreSubPath string, fileName ustr.Pat) (allFilePaths []string) {
	if ignoreSubPath != "" && !ustr.Pref(ignoreSubPath, dirPath) {
		ignoreSubPath = filepath.Join(dirPath, ignoreSubPath)
	}
	ok1, ok2 := ignoreSubPath == "", fileName == ""
	WalkAllFiles(dirPath, func(curfilepath string, _ os.FileInfo) (keepwalking bool) {
		if (ok1 || !ustr.Pref(curfilepath, ignoreSubPath)) && (ok2 || fileName.Match(filepath.Base(curfilepath))) {
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
		if fileInfos, err = Dir(dirPath); err == nil {
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
	if fileInfos, err = Dir(srcDirPath); err == nil {
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

// Dir is like ioutil.ReadDir without the sorting
func Dir(dirPath string) (contents []os.FileInfo, err error) {
	var f *os.File
	if f, err = os.Open(dirPath); err == nil {
		contents, err = f.Readdir(-1)
		_ = f.Close()
	}
	return
}

// EnsureDir attempts to create the directory `dirPath` if it does not yet exist.
func EnsureDir(dirPath string) (err error) {
	err = os.MkdirAll(dirPath, CreateModePerm)
	// if !IsDir(dirPath) {
	// if err = EnsureDir(filepath.Dir(dirPath)); err == nil {
	// 	err = os.Mkdir(dirPath, CreateModePerm)
	// }
	// }
	return
}

// Locate finds the `filePath` with the given `fileName` that is nearest to `curPath`.
func Locate(curPath string, fileName string) (filePath string) {
	if fspath := filepath.Join(curPath, fileName); IsFile(fspath) {
		filePath = fspath
	} else if fspath = filepath.Dir(curPath); fspath != curPath {
		filePath = Locate(fspath, fileName)
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
	if err := WalkAllFiles(dirPath, func(curfilepath string, curfile os.FileInfo) (keepwalking bool) {
		if !ustr.In(curfilepath, filePaths...) {
			if curfile == nil || curfile.ModTime().UnixNano() > cmpfiletimeoldest {
				isAnyNewer = true
			}
		}
		return !isAnyNewer
	}); err != nil {
		return true
	}
	return
}

func HasFilesWithSuffix(dirPath string, suff string) (has bool) {
	_ = WalkFilesIn(dirPath, func(fullpath string, fileinfo os.FileInfo) (keepWalking bool) {
		has = has || ustr.Suff(fullpath, suff)
		return !has
	})
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

// IsNewerThanTime returns whether the specified `filePath` was last modified later than the specified `unixNanoTime`.
func IsNewerThanTime(filePath string, unixNanoTime int64) (newer bool, err error) {
	if newer = true; unixNanoTime > 0 {
		var fileinfo os.FileInfo
		if fileinfo, err = os.Stat(filePath); err == nil && fileinfo != nil {
			newer = fileinfo.ModTime().UnixNano() > unixNanoTime
		}
	}
	return
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

func walk(dirPath string, self bool, traverse bool, onDir func(string, os.FileInfo) bool, onFile func(string, os.FileInfo) bool) (keepWalking bool, err error) {
	dodirs, dofiles := onDir != nil, onFile != nil
	if keepWalking = true; self && dodirs {
		fi, _ := os.Stat(dirPath)
		keepWalking = onDir(dirPath, fi)
	}
	if keepWalking {
		var fileInfos []os.FileInfo
		if fileInfos, err = WalkReadDirFunc(dirPath); err == nil || WalkIgnoreReadDirErrs {
			err = nil
			for _, fi := range fileInfos {
				fname, fmode := fi.Name(), fi.Mode()
				if fspath := filepath.Join(dirPath, fname); fmode.IsRegular() && dofiles {
					keepWalking = onFile(fspath, fi)
				} else if fmode.IsDir() {
					if dodirs {
						keepWalking = onDir(fspath, fi)
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

func Walk(dirPath string, callOnDirOnSelf bool, traverse bool, onDir func(string, os.FileInfo) bool, onFile func(string, os.FileInfo) bool) (err error) {
	if IsDir(dirPath) {
		_, err = walk(dirPath, callOnDirOnSelf, traverse, onDir, onFile)
	}
	return
}

func WalkAllFiles(dirPath string, onFile func(string, os.FileInfo) bool) error {
	return Walk(dirPath, false, true, nil, onFile)
}

func WalkDirsIn(dirPath string, onDir func(string, os.FileInfo) bool) error {
	return Walk(dirPath, false, false, onDir, nil)
}

func WalkFilesIn(dirPath string, onFile func(string, os.FileInfo) bool) error {
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

// ModificationsWatcher returns a func that mustn't be called concurrently without manual protection.
func ModificationsWatcher(dirPathsRecursive []string, dirPathsOther func() []string, restrictFilesToSuffix string, dirOk func(string, string) bool, delayIfAnyModsLaterThanThisAgo time.Duration, onModTime func(map[string]os.FileInfo, int64)) func() int {
	type gather struct {
		os.FileInfo
		modTime int64
	}
	var gathers map[string]gather
	var modnewest int64
	checkmodtime := func(fullpath string, fileinfo os.FileInfo) {
		if fileinfo == nil {
			fileinfo, _ = os.Stat(fullpath)
		}
		if fileinfo != nil {
			modtime := fileinfo.ModTime().UnixNano()
			gathers[fullpath] = gather{fileinfo, modtime}
			if modtime > modnewest {
				modnewest = modtime
			}
		}
	}
	var ondirorfile func(string, os.FileInfo) bool
	ondirorfile = func(fullpath string, fileinfo os.FileInfo) bool {
		if isdir := fileinfo.IsDir(); (isdir && (dirOk == nil || dirOk(fullpath, fileinfo.Name()))) ||
			((!isdir) && (len(restrictFilesToSuffix) == 0 || ustr.Suff(fullpath, restrictFilesToSuffix))) {
			checkmodtime(fullpath, fileinfo)
			if isdir {
				if dircontents, err := WalkReadDirFunc(fullpath); err == nil {
					for _, fi := range dircontents {
						ondirorfile(filepath.Join(fullpath, fi.Name()), fi)
					}
				}
			}
		}
		return true
	}

	var raisings map[string]os.FileInfo
	firstrun, gatherscap, holdoff, timeslastraised :=
		true, 64, int64(delayIfAnyModsLaterThanThisAgo), make(map[string]int64, 128)
	return func() (numraised int) {
		tstart := time.Now().UnixNano()
		modnewest, gathers = 0, make(map[string]gather, gatherscap)
		for i := range dirPathsRecursive {
			_, _ = walk(dirPathsRecursive[i], true, false, ondirorfile, nil)
		}
		for _, fullpath := range dirPathsOther() {
			_, _ = walk(fullpath, false, false, nil, ondirorfile)
			checkmodtime(fullpath, nil)
		}
		gatherscap = len(gathers)
		if firstrun || holdoff <= 0 || (tstart-modnewest) > holdoff {
			for fullpath, gather := range gathers {
				if tlr, _ := timeslastraised[fullpath]; tlr == 0 || gather.modTime == 0 || tlr <= gather.modTime {
					if timeslastraised[fullpath] = tstart; raisings == nil {
						raisings = make(map[string]os.FileInfo, 4)
					}
					raisings[fullpath] = gather.FileInfo
				}
			}
		}
		onModTime(raisings, tstart)
		numraised, raisings, firstrun = len(raisings), nil, false
		return
	}
}
