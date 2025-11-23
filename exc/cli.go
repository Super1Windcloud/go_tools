package main

import (
	"fmt"
	"github.com/gen2brain/go-unarr"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) == 2 {
		var dir = os.Args[1]
		fmt.Println("Directory to extract:", dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			//Println - 标准输出  打印到控制台
			// Fprintln - 向指定输出写入  打印到文件
			// Sprintln - 字符串连接 , 格式化字符串 不会产生任何输出，只是生成字符串
			panic(fmt.Sprintln("Directory does not exist: ", dir))
			// Go 也可以通过插入 os.Exit(0) 来替代断点调试和临时
		}

		var err = extractDir(dir, false)
		if err != nil {
			fmt.Println(err)
			var err = invoke7zip(dir, false)
			if err != nil {
				panic("7zip Error: " + err.Error())
			}
		}
	} else if len(os.Args) == 3 {
		var firstArg = os.Args[1]
		var secondArg = os.Args[2]
		if firstArg == "-r" {
			var err = extractDir(secondArg, true)
			if err != nil {
				fmt.Println(err)
			}
			err = invoke7zip(secondArg, true)
			if err != nil {
				panic("7zip Error: " + err.Error())
			}
		} else if secondArg == "-r" {
			var err = extractDir(firstArg, true)
			if err != nil {
				fmt.Println(err)
			}
			err = invoke7zip(firstArg, true)
			if err != nil {
				panic("7zip Error: " + err.Error())
			}

		} else {
			fmt.Println("Usage: dir_extract <directory> 解压目录下所有压缩包" +
				"\n Usage : dir_extract -r <directory> 解压目录下所有压缩包并删除压缩包")
		}

	} else {
		fmt.Println("Usage: dir_extract <directory> 解压目录下所有压缩包" +
			"\n Usage : dir_extract -r <directory> 解压目录下所有压缩包并删除压缩包")
	}

}

func invoke7zip(dir string, delete bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			var name = entry.Name()
			fmt.Println("Extracting file:", name)
			var path = filepath.Join(dir, name)
			var ext = filepath.Ext(name)
			var output = fmt.Sprintf("-o%s", strings.TrimSuffix(path, ext))
			var cmd = exec.Command("7z", "x", path, output, "-aoa")
			var err = cmd.Run()
			if err != nil {
				panic("7zip Error: " + err.Error())
			}
		}
	}

	if delete {
		var err = deleteAllZip(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractDir(dir string, delete bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			var name = entry.Name()
			var path  = filepath.Join(dir, name)
			var extension = filepath.Ext(path)
			var outDir = strings.TrimSuffix(path, extension)
			isCompressed, _, _ := IsCompressedByMagic(path)
			if !isCompressed {
				continue
			}
			fmt.Println("Extracting file:", name)
			var archive, err = unarr.NewArchive(filepath.Join(dir, name))
			if err != nil {
				return err
			}

			err = os.MkdirAll(outDir, 0755)
			if err != nil {
				return err
			}
			_, err = archive.Extract(outDir)

			if err != nil {
				return err
			}
			err = archive.Close() // 必须要关闭，否则会导致文件占用
			if err != nil {
				return err
			}
		}
	}
	if delete {
		var err = deleteAllZip(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteAllZip(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			var name = entry.Name()
			var path = filepath.Join(dir, name)
			isCompressed, _, _ := IsCompressedByMagic(path)
			if isCompressed {
				err = tryRemove(filepath.Join(dir, name))
				fmt.Println("Removed file:", name)
				if err != nil {
					panic("Remove Error: " + err.Error())
				}
			}
		}
	}
	return nil
}

func tryRemove(path string) error {
	for i := 0; i < 5; i++ {
		err := os.RemoveAll(path)
		if err == nil {
			return nil
		}
		fmt.Println("Retrying delete...", err)
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("failed to delete: %s", path)
}
