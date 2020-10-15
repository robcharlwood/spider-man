package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type Crawler struct {
	Domain   *url.URL
	Depth    int
	Parallel int
	Wait     time.Duration
}

type location struct {
	URL    *url.URL
	Parent *url.URL
	Depth  int
}

var crawlCmd = &cobra.Command{
	Use:   "crawl https://monzo.com",
	Short: "Crawl a domain",
	Long:  `Crawl a single domain`,
	RunE:  crawl,
	Args:  cobra.ExactArgs(1),
}

func init() {
	includeCrawlFlags(crawlCmd)
}

func includeCrawlFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("depth", "d", 0, "How deep to crawl")
	cmd.Flags().IntP("parallel", "p", 5, "How many parallel requests are made to the domain - default 5")
	cmd.Flags().DurationP("wait", "w", time.Second, "How long to wait between requests - default 1 second")
}

func crawl(ccmd *cobra.Command, args []string) error {
	depth, err := ccmd.Flags().GetInt("depth")
	if err != nil {
		return errors.New("The depth flag provided is not valid!")
	}
	parallel, err := ccmd.Flags().GetInt("parallel")
	if err != nil {
		return errors.New("The parallel flag provided is not valid!")
	}
	wait, err := ccmd.Flags().GetDuration("wait")
	if err != nil {
		return errors.New("The wait flag provided is not valid!")
	}
	domain, err := url.ParseRequestURI(args[0])
	if err != nil {
		return errors.New("The domain name provided is not valid!")
	}
	if len(domain.Path) != 0 {
		return errors.New("Please only provide the root domain name!")
	}

	err = Crawler{
		Domain:   domain,
		Depth:    depth,
		Parallel: parallel,
		Wait:     wait,
	}.Run()

	if err != nil {
		return errors.New("Failed to crawl " + domain.String())
	}

	return nil
}

// run the crawler
func (c Crawler) Run() error {
	queue, locations, pending := buildQueues()
	pending <- 1
	queue <- location{
		URL:    c.Domain,
		Parent: nil,
		Depth:  c.Depth,
	}

	var wg sync.WaitGroup
	for i := 0; i < c.Parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.fetch(locations, queue, pending)
		}()
	}

	wg.Wait()
	return nil
}

func buildQueues() (chan<- location, <-chan location, chan<- int) {
	queueCount := 0
	pending := make(chan int)
	locations := make(chan location)
	queue := make(chan location)
	visited := map[string]struct{}{}

	// watches the pending queue and closes the queue channel when
	// there is nothing else left to process
	go func() {
		for delta := range pending {
			queueCount += delta
			if queueCount == 0 {
				close(queue)
			}
		}
	}()

	// takes discovered links from the queue, checks to see if they have already
	// been visited and if not adds it to the locations channel for crawling.
	// when we run out of links in the queue, we close the locations and pending channels
	go func() {
		for loc := range queue {
			u := loc.URL.String()
			if _, v := visited[u]; !v {
				visited[u] = struct{}{}
				locations <- loc
			} else {
				pending <- -1
			}
		}
		close(locations)
		close(pending)
	}()

	return queue, locations, pending
}

// loops over the locations to crawl, crawls them and adds discovered urls to the queue
func (c Crawler) fetch(locations <-chan location, queue chan<- location, pending chan<- int) {
	for loc := range locations {
		var warnings []string

		// crawl the current url and grab any links found on the page.
		links, err := crawlDomain(loc)
		if err != nil {
			parent := ""
			if loc.Parent != nil {
				parent = fmt.Sprintf("on page %v", loc.Parent)
			}
			warnings = append(warnings, fmt.Sprintf("%v %s", err, parent))
		}

		// convert the discovered links to url.URLs and raise warnings for any invalid urls that will be ignored.
		urls, err := convertToURLs(links, loc.URL.Parse)
		if err != nil {
			warnings = append(
				warnings,
				fmt.Sprintf(
					"Warning: These URLs are invalid and will be ignored: %v",
					err,
				),
			)
		}

		// render table
		printTable(loc.URL, urls)
		if len(warnings) > 0 {
			fmt.Println(strings.Join(warnings, "\n"))
			fmt.Println("")
		}

		// update the pending count with the number of discovered urls - minus the url we just crawled
		pending <- len(urls) - 1

		// add links to our queue - this is non-blocking on the fetch workers
		go queueURLs(queue, urls, loc.URL, loc.Depth-1)

		// be a good neighbour and wait for the specified amount of time before running another request.
		time.Sleep(c.Wait)
	}
}

func crawlDomain(loc location) ([]string, error) {
	resp, err := http.Get(loc.URL.String())
	if err != nil {
		return nil, fmt.Errorf("Warning: Failed to get url %v: %v", loc.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"Warning: %v raised a status code %d",
			loc.URL,
			resp.StatusCode,
		)
	}

	if loc.Depth == 1 {
		return nil, err
	}

	// grab links on current page if they dont contain disallowed prefixes and are not external
	links := []string{}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"Warning: Failed to parse html body for url %s: %v",
			loc.URL.String(),
			err,
		)
	}
	document.Find("a").Each(func(i int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists {
			if !containsDisallowedPrefix(href) && !isExternal(href, loc.URL) {
				links = append(links, href)
			}
		}
	})
	return links, err
}

// adds a list of discovered urls to the queue for processing
func queueURLs(queue chan<- location, urls []*url.URL, parent *url.URL, depth int) {
	for _, u := range urls {
		queue <- location{
			URL:    u,
			Parent: parent,
			Depth:  depth,
		}
	}
}

func convertToURLs(links []string, parse func(string) (*url.URL, error)) (urls []*url.URL, err error) {
	// converts string urls into golang url.URLs so that we can make use of the URL APIs
	var invalids []string
	for _, stringURL := range links {
		loc, err := parse(stringURL)
		if err != nil {
			invalids = append(invalids, fmt.Sprintf("'%s'", stringURL))
			continue
		}
		urls = append(urls, loc)
	}
	if len(invalids) > 0 {
		err = fmt.Errorf("%v", invalids)
	}
	return urls, err
}

func containsDisallowedPrefix(url string) bool {
	// checks that our url doesn't begin with certain prefixes - these are urls
	// that we don't want to attempt to crawl e.g mailto. :)
	prefixes := []string{
		"mailto:",
		"monzo://",
		"tel:",
		"?",
		"#",
		"ftp://",
		"file://",
		"telnet://",
		"gopher://",
		"javascript:",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}

func isExternal(url string, domain *url.URL) bool {
	// if the url does not start with http(s), then its relative
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// if the url starts with the current domain (the domain we are crawling)
	// then we allow it as its not an external url or separate sub domain
	if strings.HasPrefix(url, domain.Scheme+"://"+domain.Host) {
		return false
	}

	// otherwise its external
	return true
}

// renders an ascii table to the terminal of results of the crawl
func printTable(location *url.URL, links []*url.URL) {
	tableString := &strings.Builder{}
	stringLinks := []string{}
	for _, loc := range links {
		stringLinks = append(stringLinks, loc.String())
	}
	if len(stringLinks) == 0 {
		stringLinks = append(stringLinks, "None")
	}
	data := [][]string{
		[]string{location.String(), strings.Join(stringLinks, "\n")},
	}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Page", "Discovered URLs"})
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
	fmt.Println(tableString.String())
}
