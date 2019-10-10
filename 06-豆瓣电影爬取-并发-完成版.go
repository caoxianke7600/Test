package main

import (
	"fmt"
	"strconv"
	"net/http"
	"io"
	"regexp"
	"os"
	"sync"
)

// 获取一网页数据
func myHttpGet(url string) string {
	// 调用http.GET 方法,提取网页数据 --resp
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Get err:", err)
		return ""
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)

	// 定义变量保存网页全部数据
	var result string

	// 使用 resp.Body 读取服务器回发的网页信息
	for {
		n, err := resp.Body.Read(buf)
		if n == 0 { // 判断结束 n == 0
			break
		}
		if err != nil && err != io.EOF {
			fmt.Println("Read err:", err)
			return ""
		}
		// 组织成 string
		result += string(buf[:n])
	}
	//  返回
	return result
}

// 封装函数,爬取一个页面的有效数据,保存程一个文件.
func SpiderPage(i int, group *sync.WaitGroup)  {
	url := "https://movie.douban.com/top250?start=" + strconv.Itoa((i-1)*25) + "&filter="

	// 提取网页的所有数据.
	result := myHttpGet(url)

	// 编译解析正则表达式,
	ret1 := regexp.MustCompile(`<img width="100" alt="(?s:(.*?))"`) // (.*?) 也可以提取电影名
	// 匹配电影名称
	filmNames := ret1.FindAllStringSubmatch(result, -1)

	// 编译解析正则表达式,
	ret2 := regexp.MustCompile(`<span class="rating_num" property="v:average">(?s:(.*?))</span>`) // \d\.\d
	// 匹配 评分
	filmScores := ret2.FindAllStringSubmatch(result, -1)

	// 编译解析正则表达式,
	ret3 := regexp.MustCompile(`<span>(\d+)人评价</span>`)
	// 匹配 评价人数
	peoNums := ret3.FindAllStringSubmatch(result, -1)

	// 调用函数, 保存提取的数据
	Save2File(i, filmNames, filmScores, peoNums)

	group.Done()
}

func main() {

	// 提示用户,指定爬取的起始/终止页
	fmt.Print("请输入爬取的起始页start(>=1):")
	var start int
	fmt.Scan(&start)

	fmt.Print("请输入爬取的结束页end(>=start):")
	var end int
	fmt.Scan(&end)

	// 创建 WaitGroup 对象-- wg
	var wg sync.WaitGroup
	// 在主go程中,添加等待的zigo程个数
	wg.Add(end-start+1)

	// 按用户指定的 页面, 组织对应的url +25
	for i := start; i <= end; i++ {
		// 创建并发爬取页面
		go SpiderPage(i, &wg)	// 传引用
	}
	// 在主go程中,等待子go程结束.
	wg.Wait()
}

// 封装函数,将电影名/分数/评价人数, 写出保存成一个文件
func Save2File(i int, filmNames, filmScores, peoNums [][]string) {
	// 创建文件名
	filePath := "/home/itcast/go5/" + "第" + strconv.Itoa(i) + "页.txt"
	//filePath := "第" + strconv.Itoa(i) + "页.txt"

	// 创建文件  fp
	fp, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Create err:", err)
		return
	}
	defer fp.Close()

	// 写入文件 抬头
	fp.WriteString("电影名称\t\t评价分数\t\t评价人数\n")

	// 循环 [][]string 依次提取对应的 名称/分数/人数 写入文件
	for i := 0; i < len(filmNames); i++ {
		fp.WriteString(filmNames[i][1] + "\t\t" + filmScores[i][1] +
			"\t\t" + peoNums[i][1] + "\n")
	}
	fmt.Println("第", i, "个页面爬取完毕!")
}
