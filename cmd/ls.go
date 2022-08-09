package cmd

import (
	"container/heap"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

func NewLscmd() *cobra.Command {
	var (
		count int
	)
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "ls",
		Run: func(cmd *cobra.Command, args []string) {
			filename := cmd.Flags().Lookup("file").Value.String()
			if filename == "" {
				fmt.Println("没有指定文件")
				return
			}
			err := readFile(filename, count)
			if err != nil {
				fmt.Println(err)
				return
			}

		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 100, "count for sort")

	return cmd
}

func readFile(file string, count int) (err error) {
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

	sort.Sort(res)
	fmt.Println("top:", len(res))
	for i := len(res) - 1; i >= 0; i-- {
		fmt.Println(res[i].format())
	}
	return
}

type ipInfo struct {
	Ip    string `json:"ip"`
	Count int    `json:"count"`
}

type ipHeap []ipInfo

func (h ipHeap) Len() int { return len(h) }

//小顶堆
func (h ipHeap) Less(i, j int) bool { return h[i].Count < h[j].Count }

func (h ipHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ipHeap) Pop() any {
	old := *h
	n := len(old)

	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *ipHeap) Push(x any) { // 绑定push方法，插入新元素
	*h = append(*h, x.(ipInfo))
}

func (i *ipInfo) format() string {
	res, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	return string(res)
}
