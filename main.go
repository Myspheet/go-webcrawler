package main

import (
	"fmt"
)

func main() {
	c, err := NewCrawler("https://en.wikipedia.org/wiki/Daredevil:_Born_Again_season_2")
	if err != nil {
		fmt.Println(err)
		return
	}

	wg.Add(1)
	go c.Crawl(c.StartingUrl)

	go func() {
		for link := range jobs {
			wg.Add(1)
			mu.Lock()
			visited[link] = true
			mu.Unlock()
			go c.Crawl(link)
		}
	}()

	wg.Wait()

	fmt.Println("----------------------------")
	fmt.Println("Number of unique links: ", len(visited))
	fmt.Println("----------------------------")
}
