package x

import (
	"embed"
	"net"
	"strings"
	"time"
)

func WaitForTCP(network string, ip string, port uint16, timeout time.Duration) bool {
	deadLine := time.Now().Add(timeout)

	for time.Now().Before(deadLine) {
		ColorPrintf("blue", "Waiting for %s %s:%d", network, ip, port)
		conn, err := net.DialTimeout(network, Sprintf("%s:%d", ip, port), time.Second*3)
		if err != nil {
			ColorPrintf("blue", " ...\n")
			time.Sleep(time.Second)
			continue
		} else {
			ColorPrintf("green", " OK\n")
			conn.Close()
			return true
		}
	}

	return false
}

func Ternary[T any](cond bool, trueValue, falseValue T) T {
	if cond {
		return trueValue
	} else {
		return falseValue
	}
}

func fnListJsonTasks(efs embed.FS, dir string, subDir string) ([]string, error) {
	var result []string

	if entries, err := efs.ReadDir(dir + "/" + subDir); err != nil {
		return nil, err
	} else {
		for _, entry := range entries {
			subPath := Ternary(subDir == "", entry.Name(), subDir+"/"+entry.Name())
			if entry.IsDir() {
				if subFiles, err := fnListJsonTasks(efs, dir, subPath); err != nil {
					return nil, err
				} else {
					result = append(result, subFiles...)
				}
			} else if strings.HasSuffix(entry.Name(), ".json") {
				result = append(result, strings.TrimSuffix(subPath, ".json"))
			}
		}

		return result, nil
	}
}

func ListJsonTasks(efs embed.FS, dir string) ([]string, error) {
	return fnListJsonTasks(efs, dir, "")
}
