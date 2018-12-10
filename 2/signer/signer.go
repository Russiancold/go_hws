package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type multiHashData struct {
	index int
	data  string
}

func main() {
}

func SingleHash(in, res chan interface{}) {
	t := time.Now()
	wg := &sync.WaitGroup{}
	log.Println("SingleHash call")
	mu := &sync.Mutex{}
	opened := true
	md5in := []string{}
	md5out := map[string]string{}
	go func() {
		for opened {
			if len(md5in) != 0 {
				mu.Lock()
				md5out[md5in[len(md5in)-1]] = DataSignerMd5(md5in[len(md5in)-1])
				md5in = md5in[:len(md5in)-1]
				mu.Unlock()
			}
		}
	}()
	for input := range in {
		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			mu.Lock()
			md5in = append(md5in, data)
			mu.Unlock()
			log.Println("SingleHash for " + data)
			var crc32Data string
			out := make(chan string, 10)
			wg.Add(1)
			go func(data string) {
				defer wg.Done()
				crc32Data = DataSignerCrc32(data)
			}(data)
			wg.Add(1)
			go func(out chan string) {
				defer wg.Done()
				pulled := false
				for !pulled {
					mu.Lock()
					data, exist := md5out[data]
					mu.Unlock()
					if exist {
						out <- DataSignerCrc32(data)
						delete(md5out, data)
						pulled = true
					}
				}
			}(out)
			crcMdData := <-out
			log.Println("SingleHash for " + data + " done")
			res <- crc32Data + "~" + crcMdData
		}(fmt.Sprint(input))
	}
	wg.Wait()
	opened = false
	log.Println("!!!!!!!!!!!!!", time.Now().Sub(t))
}

func MultiHash(in, res chan interface{}) {
	wg := &sync.WaitGroup{}
	log.Println("MultiHash call")
	for input := range in {
		wg.Add(1)
		go func(data string) {
			log.Println("MultiHash for " + data)
			out := make(chan multiHashData, 6)
			innerWg := &sync.WaitGroup{}
			for i := 0; i < 6; i++ {
				innerWg.Add(1)
				go func(i int, data string, out chan multiHashData) {
					defer innerWg.Done()
					out <- multiHashData{i, DataSignerCrc32(strconv.Itoa(i) + data)}
				}(i, data, out)
			}
			innerWg.Wait()
			results := make([]string, 6)
			for i := 0; i < 6; i++ {
				a := <-out
				results[a.index] = a.data
			}
			res <- strings.Join(results, "")
			log.Println("MultiHash for " + data + " done")
			wg.Done()
		}(input.(string))
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for input := range in {
		results = append(results, fmt.Sprint(input))
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}

func executeJob(worker job, in, out chan interface{}) {
	worker(in, out)
	close(out)
	log.Println("out is closed")
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 10)
	out := make(chan interface{}, 10)
	for id, job := range jobs {
		if id == len(jobs)-1 {
			job(in, out)
		} else {
			go executeJob(job, in, out)
		}
		in = out
		out = make(chan interface{}, 10)
	}
}
