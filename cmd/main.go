package main

import (
	"fmt"
	"strings"
)

func main() {
	// 1. 从环境变量中获取 key_name 的值
	// 在 shell 中，这对应于 $key_name

	// 2. 将复杂的 JSON 参数定义为字符串变量以提高可读性
	// 注意：在 Go 字符串中，我们不需要像在 shell 中那样用单引号 ' 包裹 JSON
	blockDeviceMappings := `{"DeviceName":"/dev/sda1","Ebs":{"Encrypted":false,"DeleteOnTermination":true,"Iops":3000,"SnapshotId":"snap-0bfaa845320503255","VolumeSize":8,"VolumeType":"gp3","Throughput":125}}`
	networkInterfaces := `{"AssociatePublicIpAddress":true,"DeviceIndex":0,"Groups":["sg-01ad9ea8ff4a875af"]}`
	creditSpecification := `{"CpuCredits":"standard"}`
	metadataOptions := `{"HttpEndpoint":"enabled","HttpPutResponseHopLimit":2,"HttpTokens":"required"}`

	// 3. 定义要执行的命令和它的所有参数
	// 命令是 "aws"，其余部分是参数列表
	// 每个标志和它的值都应该是独立的字符串元素
	args := []string{
		"ec2",
		"run-instances",
		"--region", "ap-east-1",
		"--image-id", "ami-050f19c6ee04f419b",
		"--instance-type", "t3.micro",
		"--block-device-mappings", blockDeviceMappings,
		"--network-interfaces", networkInterfaces,
		"--credit-specification", creditSpecification,
		"--metadata-options", metadataOptions,
		"--count", "1",
		// 如果你的 AWS CLI 配置了特定的 region 或 profile，也可以在这里添加
		// "--region", "us-east-1",
		// "--profile", "my-aws-profile",
	}

	ttm := strings.Join(args, " ")
	fmt.Println(ttm)

	// ttargs := strings.Split(ttm, " ")
	// fmt.Println(ttargs)

	// // 4. 创建一个 *exec.Cmd 对象
	// // 第一个参数是命令本身 ("aws")
	// // 第二个参数是包含所有参数的字符串切片
	// cmd := exec.Command("aws", ttargs...)

	// // 5. 将命令的标准输出和标准错误连接到当前程序的标准输出和标准错误
	// // 这样你就可以实时看到 aws cli 的所有输出了
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	// fmt.Println("\n--- 准备执行 AWS CLI 命令 ---")

	// // 6. 执行命令
	// // Run() 会启动命令并等待它完成
	// err := cmd.Run()

	// // 7. 检查执行过程中是否出错
	// if err != nil {
	// 	// 如果 aws 命令返回非零退出码，也会被视为一个错误
	// 	log.Fatalf("--- 命令执行失败: %v ---", err)
	// }

	// fmt.Println("\n--- 命令执行成功 ---")
}
