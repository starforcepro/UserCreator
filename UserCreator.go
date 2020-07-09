package main

import (
	"bytes"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"net/http"
	"sort"
	"sync"
	"time"
)

func main() {
	doRequests(1000)
	buildGraph()

	fmt.Println(responseTimeList)
}

var responseTimeList = make([]int, 0)

func doRequests(count int) {
	httpClient := http.Client{}
	requestsCount := count
	requestsDataList := make([]*http.Request, requestsCount)
	for i := 0; i < requestsCount; i++ {
		requestsDataList = append(requestsDataList, createRequestData(i))
	}
	wg := new(sync.WaitGroup)
	sendRequestsSemaphore := make(chan int, 4)
	for i := 0; i < requestsCount; i++ {
		requestsData := requestsDataList[len(requestsDataList)-1]
		requestsDataList = requestsDataList[:len(requestsDataList)-1]

		wg.Add(1)
		go measureRequestTime(httpClient, requestsData, sendRequestsSemaphore, wg)
	}
	wg.Wait()
}

func buildGraph() {
	plot, err := plot.New()
	if err != nil {
		panic(err)
	}

	averageRequestTime := getAverageRequestTime()
	maxRequestTime := getMaxRequestTime()
	minRequestTime := getMinRequestTime()

	plot.Title.Text = fmt.Sprintf("AverageRequestTime = %vms\nMaxRequestTime = %vms\nMinRequestTime = %vms", averageRequestTime, maxRequestTime, minRequestTime)
	plot.X.Label.Text = "Requests"
	plot.Y.Label.Text = "Request Time Millis"

	err = plotutil.AddLinePoints(plot,
		"Response Time", getPoints())
	if err != nil {
		panic(err)
	}

	if err := plot.Save(170*vg.Millimeter, 170*vg.Millimeter, "graph.png"); err != nil {
		panic(err)
	}
}

func getMinRequestTime() int {
	responseTimeListForSort := copyResponseTimeList()
	sort.Ints(responseTimeListForSort)
	return responseTimeList[0]
}

func getMaxRequestTime() int {
	responseTimeListForSort := copyResponseTimeList()
	sort.Ints(responseTimeListForSort)
	return responseTimeList[len(responseTimeList)-1]
}

func copyResponseTimeList() []int {
	responseTimeListForSort := make([]int, len(responseTimeList))
	copy(responseTimeListForSort, responseTimeList)
	return responseTimeListForSort
}

func getAverageRequestTime() int {
	responseTimeSum := 0
	for i := 0; i < len(responseTimeList); i++ {
		responseTimeSum += responseTimeList[len(responseTimeList)-(1+i)]
	}

	return responseTimeSum / len(responseTimeList)
}

func getPoints() plotter.XYs {
	pts := make(plotter.XYs, len(responseTimeList))
	for i := range pts {
		pts[i].Y = float64(responseTimeList[len(responseTimeList)-(1+i)])
		pts[i].X = float64(i + 1)
	}

	return pts
}

func measureRequestTime(httpClient http.Client, request *http.Request, semaphore chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	semaphore <- 0
	start := time.Now()
	_, err := httpClient.Do(request)
	if err != nil {
		panic(err)
	}
	elapsedTimeMillis := int(time.Since(start).Milliseconds())
	responseTimeList = append(responseTimeList, elapsedTimeMillis)
	<-semaphore
}

func createRequestData(iteration int) *http.Request {
	message := fmt.Sprintf(`{"context": {"__name":"example%v"}}`, iteration)
	url := "your url"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(message)))
	if err != nil {
		panic(err)
	}
	token := "your token"
	request.Header.Set("X-Token", token)

	return request
}
