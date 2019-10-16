package main

import (
	"runtime"
	"net"
	"log"
	"fmt"
	//"reflect"
	"time"
	"bufio"
	"io"
	"io/ioutil"
	"sync"
	"sync/atomic"
	"os"
	"flag"
	"strings"
	"strconv"
	"gopkg.in/yaml.v2"
)

var  (
	counter int32
	// wg用来等待程序完成
	wg sync.WaitGroup
)

var version string  = "version 1.0.0. A test program that assembles msg messages."

type conf struct {
	//yaml：yaml格式 
	LoginType		string   `yaml:"login_type"` 
	User			string   `yaml:"user"` 
	Passwd			string   `yaml:"passwd"` 

	LogVerbose  	int      `yaml:"log_verbose"`              //日志

	Msghost  		string   `yaml:"msghost"`                  
	SendType 		string   `yaml:"send_type"`                //发送类型

	Path     		string   `yaml:"path_sim"`                 //路径
	Thread      	int      `yaml:"send_thread"`              //线程数
	Sleep    		int64    `yaml:"sleep_time"`               //发送间隔
	RecordSleep    	int      `yaml:"record_sleep"`             //record发送间隔
	Cnt      		int      `yaml:"send_time"`                //发送次数
	Oem        		int      `yaml:"oem"`                      //oem
	CrazyNum    	int      `yaml:"crazy_num"`                //mode of crazy `s car num 

	Local    		string   `yaml:"local_areacode"`           //归属地区域码
	Real     		string   `yaml:"real_areacode"`            //实时区域码
	Command  		string   `yaml:"command_type"`             //command（U_REPT）
	KeyType  		string   `yaml:"key_type"`                 //key类型（TYPE:0）
	SrcType  		string   `yaml:"src_type"`                 //数据来源(1:hy808; 2:hy809; 3:yy808; 4:yy809)
	Lon		  		int      `yaml:"lon"`                      //经度1
	Lat	  			int      `yaml:"lat"`                      //纬度2
	Speed	  		string   `yaml:"speed"`                    //速度3
	Dir	  			string   `yaml:"dir"`                      //方向5
	Alt	  			string	 `yaml:"altitude"`                 //海拔6
	VssSpeed		string   `yaml:"vssspeed"`                 //vss速度7
	BaseStatus		string   `yaml:"basestatus"`               //basestatus8
	AlarmCode		string   `yaml:"alarmcode"`                //alarmcode 20
	ExtendStatus 	string   `yaml:"extendstatus"`             //extendstatus500
	Carid	  		string   `yaml:"carid"`                    //车牌号
}

func (c *conf) getConf(parameter string) *conf {
	yamlFile, err := ioutil.ReadFile(parameter)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

/** 
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) (bool) {
	var exist = true;
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false;
	}
	return exist;
}

// 读取配置
func readConfig(path string) []string {
	var arr []string

	f, err := os.Open(path)
	if err != nil {
		return arr
	}
	
	defer f.Close()

	//var i int
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#")   {
			continue;
		}

		arr = append(arr, line)
	}
	log.Println("load sim", len(arr))

	return arr
}

func checkParameter(c *conf) {
	if c.LoginType == "" {
		c.LoginType = "SEND"
	}

	if c.User == "" {
		c.User = "test"
	}

	if c.Passwd == "" {
		c.Passwd = "test"
	}

	if c.Command == "" {
		c.Command = "U_REPT"
	}

	if c.Dir == "" {
		c.Dir = "60"
	}

	if c.KeyType == ""  {
		c.KeyType = "0"
	}

	if c.Lat == 0 {
		c.Lat = 68874135
	}

	if c.Lon == 0 {
		c.Lon = 15422113
	}

	if c.Speed == "" {
		c.Speed = "400"
	}

	if c.Carid == "" {
		c.Carid = "2_京"
	}

}

func help() {
	fmt.Printf("Usage:\n")
	fmt.Printf("-v: version info\n")
	fmt.Printf("-h: help info\n")
	fmt.Printf("-c: conf (the conf file sim.yaml follows)\n")
}

func init() {

	// 解析参数
	args := os.Args
	if len(args) < 2 {
		help()
		os.Exit(0)
	}
	operate := args[1]
	switch operate {
		case "-v": {
			fmt.Printf("%s\n", version)
			os.Exit(0)
		}
		case "-h": {
			help()
			os.Exit(0)
		}
	}
}

func main() {

	// 解析参数
	var parameter string
	//var boolhelp bool
	flag.StringVar(&parameter, "c", "", "conf file")
	//flag.BoolVar(&boolhelp, "h", false, "conf file")
	flag.Parse()
	fmt.Printf("conf=%s\n", parameter)

	// 检查参数
	if len(os.Args) < 2 {
		fmt.Println("***************** msgtest *****************")
		flag.Usage()
		return
	}

	// 解析yaml
	var c conf
	c.getConf(parameter)

	// 检查参数
	checkParameter(&c)

	// 检查文件是否存在
	if c.SendType != "crazy" {
		if !checkFileIsExist(c.Path) {
			fmt.Println("当前目录缺少文件: sim.txt")
			return
		}
	}

	arr := readConfig(c.Path)
	if len(arr) <= 0 {
		fmt.Println("nono sim")
	}

	// 分配4个逻辑处理器给调度器使用
	runtime.GOMAXPROCS(4)

	//定义主机名
	addr := c.Msghost 
	//拨号操作，需要指定协议。
	conn, err := net.Dial("tcp",addr) 
	if err != nil {
		log.Fatal(err)
	}
	/*获取“conn”中的公网地址。注意：最好是加上后面的String方法，因为他们的那些是不一样的哟·当然你打印的时候
	可以不加输出结果是一样的，但是你的内心是不一样的哟！*/
	log.Println("访问公网IP地址是：",conn.RemoteAddr().String()) 
	//获取到本地的访问地址和端口。
	log.Printf("客户端链接的地址及端口是：%v\n",conn.LocalAddr()) 
	//log.Println("“conn”所对应的数据类型是：",reflect.TypeOf(conn))
	//log.Println("“conn.LocalAddr()”所对应的数据类型是：",reflect.TypeOf(conn.LocalAddr()))
	//log.Println("“conn.RemoteAddr().String()”所对应的数据类型是：",reflect.TypeOf(conn.RemoteAddr().String()))

	login_str := "LOGI " + c.LoginType + " " + c.User + " " + c.Passwd + " " + "\r\n"
	n, err := conn.Write([]byte(login_str))	
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SEND", conn.RemoteAddr().String(), login_str)

	buf := make([]byte,1024)
	n, err = conn.Read(buf)
	if err == io.EOF {
		conn.Close()
	}
	if !strings.HasPrefix(string(buf), "LACK 0 0 0")   {
		log.Printf("Login msg failed: %s", string(buf))
		return;
	}
	log.Printf("RECV %s len:%d, %s", conn.RemoteAddr().String(), n, string(buf))

	log.Println("Sendto", conn.RemoteAddr().String(), "test data by mode:", c.SendType, "start!")
	wg.Add(c.Thread)
	
	for index := 0; index < c.Thread; index++ {
		go worker(conn, arr, c, index)
	}

	atomic.AddInt32(&counter, int32(c.Thread))
	for {
		if atomic.LoadInt32(&counter) == 0  {
			break
		}

		n, err := conn.Write([]byte("NOOP \r\n"))	
		if err != nil {
			log.Fatal(err)
		}
		log.Println("SEND", conn.RemoteAddr().String(), "NOOP")

		buf := make([]byte,1024)
		n, err = conn.Read(buf)
		if err == io.EOF {
			conn.Close()
		}
		log.Printf("RECV %s len:%d, %s", conn.RemoteAddr().String(), n, string(buf))

		time.Sleep(time.Duration(10)*time.Second)
	}

	fmt.Println("Waiting To Finish")
	wg.Wait()
	fmt.Println("Terminating Program")

	//断开TCP链接。
	conn.Close()  
}

func worker(conn net.Conn, arr []string, c conf, index int) {
	defer wg.Done()
	var code = []string{"A","B","C","D","E","F","G","H","I","J","K","L","M","N","O","P","Q","R","S","T","U","V","W","X","Y","Z"} 
	var last int64 = 0

	for loop := 0; loop < c.Cnt; loop++ {
		//str := "CAITS 0_0 70104_15000010023 140105 U_REPT {TYPE:0,RET:0,1:68874135,2:15422113,3:400,4:20181206/153853,5:0,6:500,8:528387,20:0,999:1544081933,1000:360782,9:10240} \r\n"
		var str string 
		var carnum int
		if c.SendType == "crazy" {
			carnum = c.CrazyNum
		} else {
			carnum = len(arr)
		}

		now := time.Now().UnixNano() / 1e6
		for now - last < c.Sleep {
			time.Sleep(time.Duration(50)*time.Millisecond)
			now = time.Now().UnixNano() / 1e6
		}
		last = now;

		for i := 0; i < carnum; i++ {
			if c.SendType == "record" {
				str = arr[i] 
				str += "\r\n"
				if c.RecordSleep != 0 {
					time.Sleep(time.Duration(c.RecordSleep)*time.Millisecond)
				}
			} else {
				gpstime := time.Now().Format("20060102/150405")
				nowtime := strconv.FormatInt(time.Now().Unix(), 10)
				carcolor := string([]rune(c.Carid)[:1])
				carid := string([]rune(c.Carid)[2:]) + code[index] + strconv.Itoa(i+10000)
				carid998 := c.Carid + code[index] + strconv.Itoa(i+10000) 
				
				var macid string
				var sim string
				var lon string
				var lat string
				if c.SendType == "sim" {
					sim = arr[i]
					macid = strconv.Itoa(index + c.Oem) + "_" + sim
					lon = strconv.Itoa(loop*10000 + c.Lon)
					lat = strconv.Itoa(loop*10000 + c.Lat)

				} else if c.SendType == "crazy" {
					sim = strconv.Itoa(13100010000 + i)
					macid = strconv.Itoa(index + c.Oem) + "_" + sim
					lon = strconv.Itoa(c.Lon)
					lat = strconv.Itoa(c.Lat)
				}

				str = "CAITS 0_0 " 
				str += macid + " "
				str += c.Local + " " 
				str += c.Command + " {TYPE:"
				str += c.KeyType + ","
				str += "RET:0,"
				str += "1:" + lon + ","
				str += "2:" + lat + ","
				str += "3:" + c.Speed + ","
				str += "4:" + gpstime + ","
				str += "5:" + c.Dir + ","
				str += "6:" + c.Alt + ","
				str += "7:" + c.VssSpeed + ","
				str += "8:" + c.BaseStatus + ","
				str += "9:10240,"
				str += "20:" + c.AlarmCode + ","
				str += "104:" + carid + ","
				str += "202:" + carcolor + ","
				str += "500:" + c.ExtendStatus + ","
				str += "990:" + c.SrcType + ","
				str += "991:" + sim + ","
				str += "998:" + carid998 + ","
				str += "999:" + nowtime + ","
				str += "1000:" + c.Real
				str += "} \r\n"
			}

			n, err := conn.Write([]byte(str))
			if err != nil {
				log.Fatal(err)
			}
			
			if c.LogVerbose != 0{
				log.Printf("Worker %d sendto %s len:%d, data:%s", index, conn.RemoteAddr().String(), n, str)
			}
		}

		log.Printf("Worker %d send %d records, sleep: %dms", index, carnum, c.Sleep)
		//time.Sleep(time.Duration(c.Sleep)*time.Second)
		//time.Sleep(time.Duration(c.Sleep)*time.Millisecond)
	}
	atomic.AddInt32(&counter, -1)

	log.Printf("Worker %d Send finish", index, c.Sleep)
}


