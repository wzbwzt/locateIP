package cmd

import (
	"container/heap"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func NewLocatecmd() *cobra.Command {
	var (
		count  int
		tofile bool
		ips    []string
	)
	cmd := &cobra.Command{
		Use:   "locate",
		Short: "locate",
		Run: func(cmd *cobra.Command, args []string) {
			filename := cmd.Flags().Lookup("file").Value.String()
			ak := cmd.Flags().Lookup("ak").Value.String()
			if filename == "" && len(ips) == 0 {
				fmt.Println("没有指定文件或者ip")
				return
			}
			if ak == "" {
				fmt.Println("没有指定key")
				return
			}

			if len(ips) > 0 {
				locateIps(ak, ips)
			}
			if filename != "" {
				err := locateFile(ak, filename, count, tofile)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			fmt.Println(" 执行完毕")
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 100, "count for sort")
	cmd.Flags().BoolVarP(&tofile, "tofile", "d", false, "write res to csv file;name:locateIp.csv")
	cmd.Flags().StringSliceVar(&ips, "ip", nil, "point ip to locate")

	return cmd
}

func locateFile(ak, file string, count int, tofile bool) (err error) {
	bar := progressbar.NewOptions64(int64(count),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(18),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Writing moshable file..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	csvfile, err := os.Open(file)
	if err != nil {
		return
	}
	defer csvfile.Close()

	csvreader := csv.NewReader(csvfile)

	var res = make(ipHeap, 0, count)

	cc := 0
	for {
		var record []string
		record, err = csvreader.Read()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
		cc++
		if cc > count {
			recordC, err := strconv.Atoi(record[1])
			if err != nil {
				return err
			}
			heap.Push(&res, ipInfo{Ip: record[0], Count: recordC})
			heap.Pop(&res)

		} else {
			recordC, err := strconv.Atoi(record[1])
			if err != nil {
				return err
			}
			res = append(res, ipInfo{Ip: record[0], Count: recordC})
			if cc == count {
				heap.Init(&res)
			}
		}
	}
	var w *csv.Writer
	if tofile {
		newfile, err := os.OpenFile("./locateIp.csv",
			os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
		if err != nil {
			return err
		}
		newfile.WriteString("\xEF\xBB\xBF")
		w = csv.NewWriter(newfile)
		w.Write([]string{"ip", "count", "status", "info", "infocode",
			"province", "city", "adcode", "rectangle", "lon", "lat"})
	}

	sort.Sort(res)
	fmt.Println("Top:", len(res))
	for i := len(res) - 1; i >= 0; i-- {
		rsp, err := reqGaode(ak, res[i].Ip)
		if err != nil {
			return err
		}

		ipdetail := newIpInfoWithLocation(&res[i], rsp)
		err = ipdetail.loadWithLonLat()
		if err != nil {
			return err
		}

		if tofile {
			w.Write([]string{
				res[i].Ip,
				strconv.Itoa(res[i].Count),
				ipdetail.Status,
				ipdetail.Info,
				ipdetail.Infocode,
				ipdetail.Province,
				ipdetail.City,
				ipdetail.Adcode,
				ipdetail.Rectangle,
				fmt.Sprintf("%f", ipdetail.Log),
				fmt.Sprintf("%f", ipdetail.Lat),
			})

			bar.Add(1)

		} else {
			fmt.Println(ipdetail.format())
		}
		time.Sleep(time.Millisecond * 50)
	}
	if w != nil {
		w.Flush()
	}
	if err := w.Error(); err != nil {
		return err
	}
	return
}

type ipInfoWithLocation struct {
	*ipInfo
	*rspInfo
	Log float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

func newIpInfoWithLocation(ip *ipInfo, rsp *rspInfo) *ipInfoWithLocation {
	return &ipInfoWithLocation{
		ipInfo:  ip,
		rspInfo: rsp,
	}
}

func (i *ipInfoWithLocation) loadWithLonLat() error {
	if i == nil || i.Rectangle == "" {
		return nil
	}

	rectangleS := make([]float64, 0, 4)
	longS := strings.Split(i.Rectangle, ";")
	if len(longS) == 2 {
		for _, v := range longS {
			shortS := strings.Split(v, ",")
			if len(shortS) == 2 {
				for _, vv := range shortS {
					la, err := strconv.ParseFloat(vv, 64)
					if err != nil {
						return err
					}
					rectangleS = append(rectangleS, la)
				}
			}
		}
	}
	lon, lat := getLatLon(rectangleS)
	i.Log = lon
	i.Lat = lat

	return nil
}

func (i *ipInfoWithLocation) format() string {
	res, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	return string(res)
}

type rspInfo struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Adcode    string `json:"adcode"`
	Rectangle string `json:"rectangle"`
}

func reqGaode(ak, ip string) (info *rspInfo, err error) {
	urlgaode := "https://restapi.amap.com/v3/ip"
	Url, err := url.Parse(urlgaode)
	if err != nil {
		return
	}
	param := url.Values{}
	param.Add("ip", ip)
	param.Add("output", "json")
	param.Add("key", ak)
	paramEncode := param.Encode()
	Url.RawQuery = paramEncode
	urlpath := Url.String()

	req, err := http.NewRequest(http.MethodGet, urlpath, nil)
	if err != nil {
		return
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	rspdata, err := io.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	rs := map[string]interface{}{}
	err = json.Unmarshal(rspdata, &rs)
	if err != nil {
		fmt.Println(string(rspdata))
		return
	}

	info = &rspInfo{
		Status:   rs["status"].(string),
		Info:     rs["info"].(string),
		Infocode: rs["infocode"].(string),
		Province: func() string {
			v, ok := rs["province"].(string)
			if ok {
				return v
			}
			return ""
		}(),
		City: func() string {
			v, ok := rs["city"].(string)
			if ok {
				return v
			}
			return ""
		}(),
		Adcode: func() string {
			v, ok := rs["adcode"].(string)
			if ok {
				return v
			}
			return ""
		}(),
		Rectangle: func() string {
			v, ok := rs["rectangle"].(string)
			if ok {
				return v
			}
			return ""
		}(),
	}

	return
}

func getLatLon(rectangle []float64) (lon, lat float64) {
	if len(rectangle) < 4 {
		return
	}
	lon = (rectangle[0] + rectangle[2]) / 2
	lat = (rectangle[1] + rectangle[3]) / 2
	return
}

func locateIps(ak string, ips []string) {
	for _, ip := range ips {
		Ip := net.ParseIP(ip)
		if Ip == nil {
			fmt.Printf("ip:%v 格式错误\n", ip)
			return
		}
		rsp, err := reqGaode(ak, ip)
		if err != nil {
			fmt.Printf("请求高德错误,ip:%v;err:%v\n", ip, err)
			return
		}
		detail := newIpInfoWithLocation(&ipInfo{Ip: ip}, rsp)
		err = detail.loadWithLonLat()
		if err != nil {
			fmt.Println("加载经纬度失败:", err)
			return
		}
		fmt.Println(detail.format())
	}
}
