// patcher.go
// 修复从手机端导出的缓存优酷MP4文件只能在优酷播放器播放的问题
// 优酷对MP4源文件进行了简单的加密处理，导致只能在优酷播放器里播放
// 修复后的MP4文件可以在任意播放器里播放
// fix youku mp4 file.
// https://github.com/Hell0wor1d
package main

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const (
	FILE_EXT    = ".mp4"
	FILE_BUFFER = 1024
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) <= 0 {
		fmt.Println("Please enter a directory or file path.\r\nroot@kev7n.com")
		return
	}
	target, err := os.Stat(argsWithoutProg[0])
	if os.IsNotExist(err) {
		fmt.Printf("No such file or directory: %s", argsWithoutProg[0])
		return
	}

	//process directory
	if target.IsDir() {
		files, _ := ioutil.ReadDir(argsWithoutProg[0])
		for _, file := range files {
			if file.IsDir() {
				continue
			} else {
				filePath := path.Join(argsWithoutProg[0], file.Name())
				if path.Ext(filePath) == FILE_EXT {
					PatchFile(filePath)
				}
			}
		}
	} else {
		if path.Ext(argsWithoutProg[0]) == FILE_EXT {
			PatchFile(argsWithoutProg[0])
		}
	}
}

func PatchFile(fName string) {
	srcFile, err := os.Open(fName) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	// close srcFile on exit and check for its returned error
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	stat, err := srcFile.Stat()
	if err != nil {
		log.Fatal(err)
	}
	srcFileSize := stat.Size()
	srcFile.Seek(srcFileSize-8, 0)
	data := make([]byte, 8)
	count, err := srcFile.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	flat := string(data[:count])
	if strings.Contains(flat, "ftyp") {
		fNameInfo := strings.Split(fName, ".")
		newFileName := fNameInfo[0] + "_patched." + fNameInfo[1]

		// equivalent to Python's `if os.path.exists(filename)`
		if _, err := os.Stat(newFileName); err == nil {
			//log.Println("Patched file exists.", newFileName)
			return
		}
		newFile, err := os.Create(newFileName)
		if err != nil {
			log.Fatal(err)
		}

		// close newFile on exit and check for its returned error
		defer func() {
			if err := newFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		newFile.Write(data)
		// make a buffer to keep chunks that are read
		buf := make([]byte, FILE_BUFFER)
		srcFile.Seek(8, 0)
		bar := pb.StartNew(int(srcFileSize) - 8)
		bar.ShowCounters = false
		var readLength int64 = 0
		for {
			n, err := srcFile.Read(buf)
			readLength += int64(n)

			if err != nil && err != io.EOF {
				log.Fatal(err)
			}

			if n == 0 {
				break
			}

			if readLength >= (srcFileSize - 8) {
				n = n - 8
			}
			bar.Add(n)
			// write a chunk
			if _, err := newFile.Write(buf[:n]); err != nil {
				log.Fatal(err)
			}
		}
		bar.FinishPrint("The srcFile " + fName + " has been patched successfully.")
	} else {
		log.Println("The file dose not need to patch.", fName)
	}
}
