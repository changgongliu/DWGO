package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	//load函数引入方法
	"regexp"
)

type Conf struct {
	str string
}

var Rates []map[string]int = make([]map[string]int, 31)
var dur_time int

func loadData(f string) {
	//定义持续时间，以秒计数
	log.Println("begin load api log file")
	//r := regexp.MustCompile(`^{"lv":"INFO".+?"rq":"(.+?)->(.+?)".+?"io":"orig_purifier"`) //修改正则表达式
	r := regexp.MustCompile(`^{"lv":"INFO","date":"(.+?)".+?"rq":"(.+?)->(.+?)".+?"io":"orig_purifier"`)
	fs := strings.Split(f, ",")

	//log.Println("fs :  ", len(fs))//多少个文件
	//log.Println("fs :  ", fs[0])//文件名称
	//reqss := make([]map[string]int, len(fs))
	var _Rates [][31]map[string]int = make([][31]map[string]int, len(fs))

	var wg sync.WaitGroup
	for fpos := range fs {
		wg.Add(1)
		go func(fpos int) {
			defer func() {
				log.Println("end load api log file:", fs[fpos])
				wg.Done()
			}()

			f := fs[fpos]
			file, err := os.Open(f)
			if err != nil {
				log.Fatalf("load api log file: %s, error: %v", f, err)
			}
			reader := bufio.NewReader(file)
			var minuts = 0
			for {
				line, err := reader.ReadBytes('\n') //对结尾判断不正确
				//log.Println("line: ", string(line))

				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatalf("read api log file: %s, error: %v\n", f, err)
				}
				matches := r.FindSubmatch(line)
				if matches == nil { //表示未匹配的行信息
					log.Println("error on: ")
					continue
				}
				date, module, action := string(matches[1]), string(matches[2]), string(matches[3])

				//读取前dur_time时间的数据
				tmp_date := strings.Split(date, " ")
				tmp_date_second := strings.Split(tmp_date[1], ":")
				cur_minuts, err := strconv.Atoi(tmp_date_second[1])
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				// cur_second, err := strconv.Atoi(tmp_date_second[2])
				// if err != nil {
				// 	log.Fatalf("error: %v", err)
				// }
				//var cur_seconds = cur_minuts*60 + cur_second
				//fmt.Println("----cur_seconds:", cur_seconds)
				// if cur_seconds >= dur_time {
				// 	break
				// }
				log.Println("----cur_minuts:  ", cur_minuts)
				log.Println("----minuts:  ", minuts)
				if cur_minuts != minuts {
					minuts++
				}
				if minuts > 30 {
					break
				}
				ma := strings.ToLower(module + "/" + action)
				//log.Println("-------moudle-------", module)
				// log.Println("-------action-------", action)
				log.Println("-------date-------", date)
				log.Println("-------ma-------", ma)

				// 不在rates配置里就添加
				rate := _Rates[fpos][minuts]
				if rate == nil {
					rate = map[string]int{}
					_Rates[fpos][minuts] = rate //进行回塞数组
				}
				rate[ma]++
			}
		}(fpos)
	}
	wg.Wait()

	// for _, rates := range _Rates {
	// 	for k, value := range rates {
	// 		//Rates[k] += v
	// 		for m, n := range value {
	// 			log.Println("--- ",k)
	// 			Rates[k][m] += n
	// 		}
	// 	}
	// }
	for _, rates := range _Rates {
		for k, value := range rates {
			//Rates[k] += v
			for m, n := range value {
				tmp := Rates[k]
				if tmp == nil {
					tmp = map[string]int{}
					Rates[k] = tmp //进行回塞数组
				}
				Rates[k][m] += n
			}
		}
	}
	//log.Println("len of _Rates: ", len(_Rates))
	log.Println("------end load api log file-------")
}

func operateKey(f string) {
	log.Println("------begin operate log file--------")
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("load api log file: %s, error: %v", f, err)
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n') //对结尾判断不正确
		//log.Println("line: ", string(line))

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("read api log file: %s, error: %v\n", f, err)
		}
		fs := strings.Split(string(line), "key")
		for _, value := range fs {
			fmt.Printf("key%s\n", value)
		}

	}
}

func operateFormat(f string) {
	log.Println("------begin operate log file--------")
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("load api log file: %s, error: %v", f, err)
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n') //对结尾判断不正确
		//log.Println("line: ", string(line))

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("read api log file: %s, error: %v\n", f, err)
		}
		fs := strings.Split(string(line), ",")
		if len(fs) != 3 {
			continue
		}
		fmt.Printf("%-50s %-15s %-10s", fs[0], fs[1], fs[2])
	}
}

func main() {
	dur_time = 60
	dir := "/data/server/logs_tmp/logs" //日志地址
	// dir := "lhg"
	// for i := 0; i < 3; i++ {
	// 	if info, err := os.Stat(dir); err == nil && info.IsDir() {
	// 		break
	// 	}
	// 	dir = filepath.Join("..", dir)
	// }
	fmt.Println("begin path", dir)
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) (reterr error) {
			if err != nil {
				reterr = err
				return
			}
			if info.IsDir() {
				return
			}
			if filepath.Ext(path) != ".log" { //文件格式，对文件进行过滤
				return
			}
			//对文件名进行过滤
			//r := regexp.MustCompile(`20180403071_10mins`) //对文件进行过滤
			r := regexp.MustCompile(`Index-20180403071`) //对文件进行过滤
			//r := regexp.MustCompile(`test.log`) //对文件进行过滤
			matches := r.MatchString(path)
			if !matches {
				return
			}

			// fs := strings.Split(path, string(filepath.Separator))
			// fmt.Println(fs[len(fs)-1])
			fmt.Println("-----------------------------------", path, "------------------------------------------")
			//operateKey(path)
			//operateFormat(path)
			loadData(path)

			return
		},
	)
	if err != nil {
		log.Printf("conf(mongo) not right: %v\n", err)
		os.Exit(-1)
	}
	//遍历数组Rates-进行一次统计就行
	//fmt.Printf("length of Rates: %d", len(Rates))
	for key, value := range Rates {
		for k, v := range value {
			fmt.Printf("minuts: %-5d key: %-50s value: %-15d qps: %-10d\n", key, k, v, v/60)
		}
		fmt.Printf("\n")
	}
}
