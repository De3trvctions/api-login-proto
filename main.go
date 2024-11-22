package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	_ "google.golang.org/protobuf/reflect/protoreflect"
	_ "google.golang.org/protobuf/runtime/protoimpl"
)

func main() {
	// 新增服务需要在下面添加一行，用于生成proto代码（参数为文件夹名字）
	gen(
		"login",
		"common",
		"account",
	)
}

func gen(serviceName ...string) {
	VerifyProtoName(serviceName)
	wg := sync.WaitGroup{}
	for _, name := range serviceName {
		wg.Add(1)
		go func(currentName string) {
			defer wg.Done()
			//protoc -Icommon -Ilogin -I../ --go_out=login --go_opt=paths=source_relative --go-grpc_out=login --go-grpc_opt=paths=source_relative login/*.proto
			//fmt.Println(fmt.Sprintf("%s   -> ", currentName), exec.Command("protoc", "-Icommon", fmt.Sprintf("-I%s", currentName), "-I../", fmt.Sprintf("--go_out=%s", currentName), "--go_opt=paths=source_relative", fmt.Sprintf("--go-grpc_out=%s", currentName), "--go-grpc_opt=paths=source_relative", fmt.Sprintf("%s/*.proto", currentName)).Run())
			CleanFiles(currentName)
			fileList := GetFiles(currentName)
			for _, oneName := range fileList {
				cmd := exec.Command("protoc", "-Icommon", fmt.Sprintf("-I%s", currentName), "-I../", fmt.Sprintf("--go_out=%s", currentName), "--go_opt=paths=source_relative", fmt.Sprintf("--go-grpc_out=%s", currentName), "--go-grpc_opt=paths=source_relative", fmt.Sprintf("%s/%s.proto", currentName, oneName))
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					fmt.Println(fmt.Sprintf("%s/%s => 失败", currentName, oneName), fmt.Sprint(err)+": "+stderr.String())
					return
				} else {
					fmt.Println(fmt.Sprintf("%s/%s => 成功", currentName, oneName))
				}
			}
		}(name)
	}
	wg.Wait()
}

// 检查.proto是否有重名
func VerifyProtoName(serviceNameArray []string) {
	// 把检查结果放进map，格式如下
	// { "player.proto"  : ["player", "gameapi"]
	//   "activity.proto": ["game"]	}
	resultMap := make(map[string][]string)
	for _, serviceName := range serviceNameArray {
		fileNameList := GetFiles(serviceName)
		for _, fileName := range fileNameList {
			_, exists := resultMap[fileName]
			if !exists {
				resultMap[fileName] = []string{serviceName}
			} else {
				resultMap[fileName] = append(resultMap[fileName], serviceName)
			}
		}
	}
	// 移除没有重名的entry
	for key, value := range resultMap {
		if len(value) == 1 {
			delete(resultMap, key)
		}
	}
	// 把检查结果print去console，如果有重名就不要让流水线过
	if len(resultMap) > 0 {
		for key, value := range resultMap {
			fmt.Println(fmt.Sprintf("重名'%s.proto'出现在 => %s", key, strings.Join(value, ", ")))
		}
		fmt.Println("因发现重名proto文件, 不执行后续逻辑。")
		os.Exit(1)
	}
}

// 只获取proto文件
func GetFiles(folder string) (fileList []string) {
	files, _ := os.ReadDir(folder)
	for _, file := range files {
		fileName := file.Name()
		if fileName[len(fileName)-6:] == ".proto" {
			fileList = append(fileList, fileName[:len(fileName)-6])
		}
	}

	return
}

// 只获取proto文件
func CleanFiles(folder string) {
	files, _ := os.ReadDir(folder)
	for _, file := range files {
		fileName := file.Name()
		if fileName[len(fileName)-6:] == ".pb.go" {
			err := os.Remove(folder + "/" + fileName)
			if err != nil {
				fmt.Println(fmt.Sprintf("%s/%s => 删除失败,%s", folder, fileName, err.Error()))
			} else {
				fmt.Println(fmt.Sprintf("%s/%s => 删除成功", folder, fileName))
			}

		}
	}
	return
}
