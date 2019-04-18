# ufs
--
    import "github.com/go-leap/fs"


## Usage

```go
var (
	// CreateModePerm (rwx for user,group,other) is used by all functions in this package that create file-system directories or files, namely: `EnsureDir`, `WriteBinaryFile`, `WriteTextFile`.
	CreateModePerm = os.ModePerm

	// Del aliases `os.RemoveAll` — merely a handy short-hand during rapid iteration in non-critical code-paths that already do import `ufs` to not have to repeatedly pull in and out the extra `os` import.
	Del = os.RemoveAll

	WalkReadDirFunc       = ioutil.ReadDir
	WalkIgnoreReadDirErrs bool
)
```

#### func  AllFilePathsIn

```go
func AllFilePathsIn(dirPath string, ignoreSubPath string, fileName ustr.Pat) (allFilePaths []string)
```
AllFilePathsIn collects the full paths of all files directly or indirectly
contained under `dirPath`.

#### func  ClearDir

```go
func ClearDir(dirPath string, keepNames ...string) (err error)
```
ClearDir removes everything inside `dirPath`, but not `dirPath` itself and also
excepting items inside `dirPath` (but not inside sub-directories) with one of
the specified `keepNames`.

#### func  CopyAllFilesAndSubDirs

```go
func CopyAllFilesAndSubDirs(srcDirPath, dstDirPath string, skipFileSuffix string, skipDirNames ...string) (err error)
```
CopyAllFilesAndSubDirs copies all files and directories inside `srcDirPath` into
`dstDirPath`. All sub-directories whose `os.FileInfo.Name` is contained in
`skipDirNames` (if supplied) are skipped, and so are files with names ending in
`skipFileSuffix` (if supplied).

#### func  CopyFile

```go
func CopyFile(srcFilePath, dstFilePath string) (err error)
```
CopyFile attempts an `io.Copy` from `srcFilePath` to `dstFilePath`.

#### func  Dir

```go
func Dir(dirPath string) (contents []os.FileInfo, err error)
```
Dir is like ioutil.ReadDir without the sorting

#### func  EnsureDir

```go
func EnsureDir(dirPath string) (err error)
```
EnsureDir attempts to create the directory `dirPath` if it does not yet exist.

#### func  IsAnyFileInDirNewerThanTheOldestOf

```go
func IsAnyFileInDirNewerThanTheOldestOf(dirPath string, filePaths ...string) (isAnyNewer bool)
```
IsAnyFileInDirNewerThanTheOldestOf returns whether any file directly or
indirectly contained in `dirPath` is newer than the oldest of the specified
`filePaths`.

#### func  IsDir

```go
func IsDir(fsPath string) bool
```
IsDir returns whether a directory (not a file) exists at the specified `fsPath`.

#### func  IsFile

```go
func IsFile(fsPath string) bool
```
IsFile returns whether a file (not a directory) exists at the specified
`fsPath`.

#### func  IsNewerThanTime

```go
func IsNewerThanTime(filePath string, unixNanoTime int64) (newer bool, err error)
```
IsNewerThanTime returns whether the specified `filePath` was last modified later
than the specified `unixNanoTime`.

#### func  Locate

```go
func Locate(curPath string, fileName string) (filePath string)
```
Locate finds the `filePath` with the given `fileName` that is nearest to
`curPath`.

#### func  ModificationsWatcher

```go
func ModificationsWatcher(delayIfAnyModsLaterThanThisAgo time.Duration, dirPathsRecursive []string, dirPathsOther []string, restrictFilesToSuffix string, onModTime func(map[string]os.FileInfo, int64)) func()
```

#### func  ReadTextFile

```go
func ReadTextFile(filePath string) (string, error)
```
ReadTextFile is a `string`-typed convenience short-hand for `ioutil.ReadFile`.

#### func  ReadTextFileOr

```go
func ReadTextFileOr(filePath string, fallback string) string
```
ReadTextFileOr calls `ReadTextFile(filePath)` but returns `fallback` on `error`.

#### func  ReadTextFileOrPanic

```go
func ReadTextFileOrPanic(filePath string) string
```
ReadTextFileOrPanic calls `ReadTextFile(filePath)` but `panic`s on `error`.

#### func  SaveTo

```go
func SaveTo(src io.Reader, dstFilePath string) (err error)
```
SaveTo attempts an `io.Copy` from `src` to `dstFilePath`.

#### func  Walk

```go
func Walk(dirPath string, callOnDirOnSelf bool, traverse bool, onDir func(string, os.FileInfo) bool, onFile func(string, os.FileInfo) bool) (err error)
```

#### func  WalkAllFiles

```go
func WalkAllFiles(dirPath string, onFile func(string, os.FileInfo) bool) error
```

#### func  WalkDirsIn

```go
func WalkDirsIn(dirPath string, onDir func(string, os.FileInfo) bool) error
```

#### func  WalkFilesIn

```go
func WalkFilesIn(dirPath string, onFile func(string, os.FileInfo) bool) error
```

#### func  WriteBinaryFile

```go
func WriteBinaryFile(filePath string, contents []byte) error
```
WriteBinaryFile is a convenience short-hand for `ioutil.WriteFile` that also
`EnsureDir`s the destination.

#### func  WriteTextFile

```go
func WriteTextFile(filePath, contents string) error
```
WriteTextFile is a `string`-typed convenience short-hand for `ioutil.WriteFile`
that also `EnsureDir`s the destination.
