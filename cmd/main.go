package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/urfave/cli/v2"
)

type MovieDetail struct {
	ID   int
	Name string
}

func getData(wg *sync.WaitGroup, val string, out chan<- MovieDetail) {
	defer wg.Done()
	url := "https://api.themoviedb.org/3/movie/" + val + "?api_key=3e6d17bf52794586b274fcbd78187e0a"
	var body []byte
	err := retry.Do(
		func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {

				return err
			}

			return nil
		},
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retrying request after error: %v", err)
		}), retry.Delay(3*time.Second),
	)
	if err != nil {
		log.Fatalf("error when create request")

		panic(err)

	}
	resp := map[string]interface{}{}
	// var resp string
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("err when marshal", err)
	}
	intTest, _ := strconv.Atoi(val)
	movieName := fmt.Sprintf("%v", resp["original_title"])
	if resp["original_title"] == nil {
		movieName = "Not Found"
	}

	out <- MovieDetail{
		ID:   intTest,
		Name: movieName,
	}

}
func main() {
	(&cli.App{}).Run(os.Args)
	result := map[int]string{}

	movies := []string{}
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "list",
				Value:   "1",
				Aliases: []string{"l"},
				Usage:   "List of movies what to find with format ex: 1,2,3,4,5",
			},
		},
		Action: func(cCtx *cli.Context) error {
			movies = strings.Split(cCtx.String("list"), ",")

			return nil
		},
	}
	var wg sync.WaitGroup
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	chanRes := make(chan MovieDetail, 1)
	// for _, val := range movies {

	go func() {
		for {
			res := <-chanRes
			result[res.ID] = res.Name
		}
	}()
	// for i := 1; i <= 100; i++ {
	// val := strconv.Itoa(i)
	for _, i := range movies {
		val := i
		intI, _ := strconv.Atoi(i)
		wg.Add(1)
		// because has limitation on how many connection can be open and used, so add more 10%
		if len(movies) > 5000 && intI%5000 == 0 {
			time.Sleep(time.Second * 10)
		}

		go getData(&wg, val, chanRes)
	}
	wg.Wait()
	for {
		// fmt.Println(len(result), "total result before running indexes")
		if len(result) == len(movies) {
			break
		}
	}
	indexes := []int{}
	for idx := range result {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	for _, val := range indexes {
		fmt.Println("" + fmt.Sprintf(`Movie: %d -> %s`, val, result[val]))

	}

}
