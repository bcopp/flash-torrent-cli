package main

import (
	// Standard libraries
	"bufio"
	List "container/list"
	"fmt"
	"math/rand"
	URL "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"encoding/json"

	// Third party libraries
	"github.com/PuerkitoBio/goquery" // Jquery like lib for Go
	log "github.com/Sirupsen/logrus" // Logger
	"github.com/jessevdk/go-flags"   // CMD Parser
)

// =====================
// ==== CMD PARSING ====
// =====================
type Opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	Meme []bool `short:"m" long:"meme" description:"Start program with a different name each time"`

	Search string `short:"s" long:"search" description:"The search you would like to be made on \"The Pirate Bay\"" value-name:"SEARCH"`

	DownloadFolder string `short:"f" long:"folder" description:"Folder to download torrents to. DEFAULTS to $HOME/Downloads" value-name:"FOLDER" required:"false"`

	NoVlc []bool `short:"n" long:"no-vlc" description:"Do NOT run VLC when the file has finished downloading"`

	JsonFE []bool `short:"j" long:"json" description:"Print the results to the screen as JSON"`
}

func (opts Opts) String() string {
	return fmt.Sprintf("Verbosity: %[1]v\nSearch: %[2]v", opts.Verbose, opts.Search)
}

func setup() Opts {
	var opts Opts
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic(err)
	}

	// Setup Verbosity
	log.SetOutput(os.Stdout)
	switch len(opts.Verbose) {
	case 1:
		log.SetLevel(log.InfoLevel)
	case 2:
		log.SetLevel(log.DebugLevel)
	case 3:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.FatalLevel)
	}

	// Setup Download Folder if not specified
	if opts.DownloadFolder == "" {
		opts.DownloadFolder = "/home/creator-76/Downloads/test"
	}


	return opts
	/*
	   ----------Logger Examples---------
	   log.Trace("Something very low level.")
	   log.Debug("Useful debugging information.")
	   log.Info("Something noteworthy happened!")
	   log.Warn("You should probably take a look at this.")
	   log.Error("Something failed but I'm not quitting.")
	   // Calls os.Exit(1) after logging
	   log.Fatal("Bye.")
	   // Calls panic() after logging
	   log.Panic("I'm bailing.")
	*/
}

// If 'transmission-daemon' is not running, start it
func initTorrentDaemon() {
	out, err := exec.Command("bash", "-c", "ps -ef | grep transmission | awk '{print $8}'").Output()
	if err != nil {
		log.Fatal("ERR reading io for daemon initialization: %s", err)
		panic("QUITTING")
	}
	if !strings.Contains(string(out), "transmission") {
		log.Info("Starting Torrent Daemon")
		cmd := exec.Command("transmission-daemon")
		cmd.Start()
		time.Sleep(100 * time.Millisecond)
	}
}

// =====================
// ===== Scraping ======
// =====================

// TODO:
// Create folder for download, Therefore vlc can search for compatible file under namespace

// Fix transmission-remote seeding ratio
// Set up with peer blocker

// Program Design
// Parse cmd, Init services
// Wire up program

func main() {
	opts := setup()
	initTorrentDaemon()

	// Run from cmd opts firstime
	var fe FrontEnd
	if len(opts.JsonFE) > 0{
		fe = JsonFE{}
	} else{
		fe = TerminalFE{}
	}
	fe.run(opts)
}

// ==========================
// ======= FRONT END ========
// ==========================

type FrontEnd interface {
	run(Opts)
	Display
	Input
}
type Display interface {
	renderSplash(ops Opts)
	renderNoResults()
	renderSearchPrompt()
	renderSelectPrompt()
	renderTorrentInfos([]TorrentInfo)
}
type Input interface {
	userSelectTorrent() string
	getUserInput() string
}

type TerminalFE struct {
	FrontEnd
}
type JsonFE struct {
	FrontEnd
}

// =====================
// ====== DISPLAY ======
// =====================
func (fe JsonFE) run(opts Opts){
	backend := TorrentCache{infos: make(map[string][]TorrentInfo)}
	infos := backend.search(opts.Search)
	fe.renderTorrentInfos(infos)
}
func (fe JsonFE) renderSplash(opts Opts) {}
func (fe JsonFE) renderNoResults() {}
func (fe JsonFE) renderSearchPrompt() {}
func (fe JsonFE) renderSelectPrompt() {}
func (fe JsonFE) renderTorrentInfos(infos []TorrentInfo) {
	jsonInfos, _ := json.Marshal(infos)
	os.Stdout.Write(jsonInfos)
}
func (fe JsonFE) userSelectTorrent() string {return ""}
func (fe JsonFE) getUserInput() string {return ""}


func (fe TerminalFE) renderSplash(opts Opts) {
	r := 0
	if len(opts.Meme) > 0 {
		r = rand.Intn(9)
	}
	lolz := [9]string{"FLASH TORRENT", "BETTER NETLFIX", "LIFE OF PI-RATE", "YARRRG", "Super Smash Brother Melee, the 2001 classic for the Nintendo gamecube", "PIGS PIGS PIGS", "VIM", "LINUX RULEZ", "1337"}
	fmt.Printf("\n-----------------------------------------------------------------")
	fmt.Printf("\n--------------------------%s--------------------------", lolz[r])
	fmt.Printf("\n-------------------------By Brendan Copp-------------------------")
	fmt.Printf("\n-----------------------------------------------------------------\n")
}
func (fe TerminalFE) renderSearchPrompt() {
	fmt.Printf("Type name of something you would like to search or type 'exit'\n")
}
func (fe TerminalFE) renderNoResults() {
	fmt.Printf("There were no results found for your search, please try again.\n")
}
func (fe TerminalFE) renderSelectPrompt() {
	fmt.Printf("-------------------------------------------------------------\n")
	fmt.Printf("Type the ID of the torrent you would like to start:\n")
	fmt.Printf("Type 'exit' to quit, 'search' to search again:\n")
}
func (fe TerminalFE) renderTorrentInfos(infos []TorrentInfo) {
	fmt.Printf("\n")
	for i, info := range infos {
		fmt.Printf("ID: %d", i+1)
		fmt.Print(info.toStringAbbr(), "\n")
	}
}

func (fe TerminalFE) getUserInput() string {
	scanner := bufio.NewReader(os.Stdin)
	raw, _ := scanner.ReadString('\n')
	s_raw := strings.Split(raw, "\n")
	input := strings.TrimSpace(s_raw[0])
	return input
}

// =====================
// ====== INPUT ======
// =====================
func (fe TerminalFE) run(opts Opts) {
	backend := TorrentCache{infos: make(map[string][]TorrentInfo)}
	selectAgain := true
	var infos []TorrentInfo

	fe.renderSplash(opts)

	if opts.Search != "" {
		infos = backend.search(opts.Search)
	} else {
		fe.renderSearchPrompt()
		res := fe.getUserInput()
		if res == "exit" {
			selectAgain = false
		} else {
			selectAgain = true
			infos = backend.search(res)
		}
	}

	for selectAgain == true {
		if len(infos) == 0 {
			fe.renderNoResults()
			fe.renderSearchPrompt()
			res := fe.getUserInput()
			if res == "exit" {
				selectAgain = false
			} else {
				selectAgain = true
				infos = backend.search(res)
			}
		} else {
			fe.renderTorrentInfos(infos)
			fe.renderSelectPrompt()
			i := fe.userInputTerminal()
			if i == -1 {
				selectAgain = false
			} else if i == -2 {
				fe.renderSearchPrompt()
				res := fe.getUserInput()
				if res == "exit" {
					selectAgain = false
				} else {
					selectAgain = true
					infos = backend.search(res)
				}
			} else if i <= len(infos) && i > 0 {
				i = i - 1
				selectAgain = false
				addAndStartTorrent(opts, infos[i])
			}
		}
	}
}

func (fe TerminalFE) userInputTerminal() int {
	for true {
		input := fe.getUserInput()
		i, err := strconv.Atoi(input)
		if input == "exit" {
			return -1
		} else if input == "search" {
			return -2
		} else if err == nil && i <= 5 && i > 0 {
			// TODO: Figure out unicode and calculate est download time
			//fmt.Printf("Est Download Time: %s", infos[i].Title)
			return i
		}
		return -3
	}
	return -1
}

// ========================
// ========= CMD ==========
// ========================
func addAndStartTorrent(opts Opts, info TorrentInfo) {
	log.Trace("Initiating addAndStartTorrent")
	log.Info("Adding Torrent")
	downloadFolder := fmt.Sprint(opts.DownloadFolder, "/", strings.ReplaceAll(info.Title, " ", "-"))
	title := strings.ReplaceAll(info.Title, " ", "+")
	log.Info("INNER FOLDER: ", downloadFolder)
	cmd := exec.Command("mkdir", downloadFolder)
	cmd.Start()
	err := cmd.Wait()

	if err != nil {
		log.Warn("Error making dir: ", err)
	}

	out, err := exec.Command("transmission-remote", "--auth", "transmission:transmission", "-w", downloadFolder, "-a", info.Magnet).Output()
	if err != nil {
		log.Fatal("Err Adding Torrent:%s\n", err)
		log.Fatal("Printing Torrent Info: %s\n\n", info.toString())
		log.Fatal("Most likely daemon is not running")
		log.Fatal("try command: transmission-daemon")
		panic("FATAL ERROR")
	}
	log.Debug("transmission-remote: %s", string(out))

	time.Sleep(100 * time.Millisecond)

	log.Debug("Getting Torrent ID...")
	id, err := exec.Command("bash", "-c", "transmission-remote --auth transmission:transmission -l | grep -i '"+title+"' |awk '{print$1}'").Output()
	if err != nil {
		log.Fatal("Err Getting Torrent ID:%s", err)
	}
	log.Debug("Torrent ID:" + string(id))

	log.Debug("Starting Torrent")
	out, err = exec.Command("transmission-remote", "--auth", "transmission:transmission", "-t", string(id), "-s").Output()
	if err != nil {
		log.Fatal("Err Starting Torrent:%s", err)
	}
	log.Debug("transmission-remote: %s", string(out))
	i, err := strconv.Atoi(string(id))
	if err != nil {
		log.Warn("Error parsing ID: ", i)
	}
	getTorrentStatusUntilFinished(string(title), i)

	log.Info("Torrent Finished")
	if len(opts.NoVlc) > 0{
		log.Info("Running vlc")
		cmd = exec.Command("./vlc_helper", "-f", downloadFolder)
		cmd.Start()
	}
}

func getTorrentStatusUntilFinished(title string, id int) {
	var finished = false
	var last = -1
	fmt.Printf("Starting Torrent. This may take a few seconds...")

	for !finished {
		time.Sleep(4 * time.Second)
		out, err := exec.Command("bash", "-c", "transmission-remote --auth transmission:transmission -l |grep -i '"+title+"' |awk '{print $2}'").Output()
		if err != nil {
			fmt.Printf("\nTorrent Status Err:%s", err)
		}
		if len(out) != 0 {
			ostr := string(out)
			ostr = strings.TrimSpace(ostr)
			ostr = strings.Split(ostr, "\n")[0]
			ostr = strings.Split(ostr, "%")[0]
			statusNum, err := strconv.Atoi(ostr)
			if err != nil {
				log.Warn("Error conv string in getTorrent: ", err)
			}

			if statusNum != last {
				last = statusNum
				fmt.Printf("Status: %s", out)

				if strings.Contains(string(out), "100") {
					time.Sleep(1 * time.Second)
					finished = true
				}
			}
		}
	}
}

// ========================
// ======= BACKEND ========
// ========================

type TorrentCache struct {
	infos map[string][]TorrentInfo
}

func (cache TorrentCache) search(search string) []TorrentInfo {
	if infos, ok := cache.infos[search]; ok {
		return infos
	} else {
		url := createPirateURL(search)
		dom := getDom(url)
		log.Info("Retreiving info from %s...", url)
		infos := scrapePirateBaySearch(dom)
		//cache.infos[search][0] = infos[0]
		for _, info := range infos {
			cache.infos[search] = append(cache.infos[search], info)
		}
		return cache.infos[search]
	}
}

// ========= DATA =========
type TorrentInfo struct {
	Title    string
	Link     string
	Magnet   string
	By       string
	Seeders  string
	Leechers string
	Comments string
	Size     string
	Date     string
	Details  string
}

func (r TorrentInfo) toString() string {
	var s = fmt.Sprintf("\n Title: %s\n Link: %s\n By: %s\n Seeders: %s\n Leechers: %s\n Comments: %s\n Size: %s\n Date: %s\n Magnet: ", r.Title, r.Link, r.By, r.Seeders, r.Leechers, r.Comments, r.Size, r.Date)
	s = fmt.Sprint(s, r.Magnet, "\n\n")
	return s
}
func (r TorrentInfo) toStringAbbr() string {
	var s = fmt.Sprintf("\n Title: %s\n Seeders: %s\n Leechers: %s\n Size: %s\n Date: %s\n", r.Title, r.Seeders, r.Leechers, r.Size, r.Date)
	return s
}

// ======== METHODS ========
func getDom(url string) *goquery.Document {
	log.Trace("Retrieving DOM from URL: ", url, "\n")
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal("The Pirate Bay is Probably Down\n%s\n", err)
		panic("TRY AGAIN IN A FEW MINUTES")
	}
	return doc
}

func getProgramDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func createPirateURL(search string) string {
	log.Trace("Constructing Pirate URL\n")
	search = URL.PathEscape(search)

	var domain = "https://thepiratebay.org"
	var pre = "/search/"
	var post = "/0/99/0" // Defines what page (0-99)
	var url = domain + pre + search + post

	log.Debug(url + "\n")
	return url
}

func scrapePirateBaySearch(doc *goquery.Document) []TorrentInfo {
	log.Trace("Entering scrapePirateBaySearch")

	// Get Links
	var results = List.New()
	doc.Find("#SearchResults .detLink").Each(func(index int, item *goquery.Selection) {
		//title := item.Text()
		//var linkTag = item.Find(".detLink")
		var link, _ = item.Attr("href")
		results.PushBack(link)
	})

	// Take 5 -> spits out an array of string urls
	var c = 0
	const LIMIT = 5
	var results_filter []string
	for l := results.Front(); l != nil; l = l.Next() {
		results_filter = append(results_filter, "https://thepiratebay.org/"+l.Value.(string))
		c++
		if c >= LIMIT {
			break
		}
	}

	log.Info("Scraping Torrent Data")
	var results_scrape []TorrentInfo
	for i := 0; i < len(results_filter); i++ {
		results_scrape = append(results_scrape, scrapePirateDesc(results_filter[i], pirateMap))
	}
	/*
		var info = TorrentInfo{
			Title:    "",
			Link:     "",
			Magnet:   "",
			By:       "",
			Seeders:  "",
			Leechers: "",
			Comments: "",
			Size:     "",
			Date:     "",
			Details:  "",
		}
	*/

	return results_scrape

}

func scrapePirateDesc(url string, scrapeMap func(*goquery.Selection) string) TorrentInfo {
	var doc = getDom(url)
	log.Trace("Scraping Info" + url)
	var info = TorrentInfo{
		Title:    "",
		Link:     "",
		Magnet:   "",
		By:       "",
		Seeders:  "",
		Leechers: "",
		Comments: "",
		Size:     "",
		Date:     "",
		Details:  "",
	}

	const by = "By"
	const seeders = "Seeders"
	const leechers = "Leechers"
	const comments = "Comments"
	const size = "Size"
	const date = "Date"
	var keys = [6]string{by, seeders, leechers, comments, size, date}

	info.Title = strings.TrimSpace(doc.Find("#title").Text())
	info.Link = url
	info.Magnet, _ = doc.Find(".download").Children().First().Attr("href")

	var scrape = ""
	var infoScrape = [2]string{"", ""}
	doc.Find(".col2").Children().Each(func(index int, item *goquery.Selection) {
		scrape = scrapeMap(item)
		if infoScrape[0] == "" {
			for i := 0; i < 6; i++ {
				if scrape == keys[i] {
					infoScrape[0] = scrape
					break
				}
			}
		} else {
			infoScrape[1] = scrape
			if infoScrape[0] == by {
				info.By = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == seeders {
				info.Seeders = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == leechers {
				info.Leechers = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == comments {
				info.Comments = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == size {
				info.Size = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == date {
				info.Date = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			}
		}
	})
	doc.Find(".col1").Children().Each(func(index int, item *goquery.Selection) {
		scrape = scrapeMap(item)
		if infoScrape[0] == "" {
			for i := 0; i < 6; i++ {
				if scrape == keys[i] {
					infoScrape[0] = scrape
					break
				}
			}
		} else {
			infoScrape[1] = scrape
			if infoScrape[0] == by {
				info.By = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == seeders {
				info.Seeders = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == leechers {
				info.Leechers = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == comments {
				info.Comments = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == size {
				info.Size = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			} else if infoScrape[0] == date {
				info.Date = infoScrape[1]
				infoScrape[0] = ""
				infoScrape[1] = ""
			}
		}
	})

	return info
}

func pirateMap(item *goquery.Selection) string {
	text := strings.Split(item.Text(), "\n")[0]
	//log.Trace("Text: " + text)
	//var title = "Title:"
	//var link = "Link:"
	//var magnet = "Magnet:"

	// Pirate specific matches
	const by = "By:"
	const seeders = "Seeders:"
	const leechers = "Leechers:"
	const comments = "Comments"
	const size = "Size:"
	const date = "Uploaded:"

	if strings.Contains(text, by) {
		return "By"
	} else if strings.Contains(text, seeders) {
		return "Seeders"
		log.Trace("Contains Seeders: " + text)
	} else if strings.Contains(text, leechers) {
		return "Leechers"
		log.Trace("Contains Leechers: " + text)
	} else if strings.Contains(text, comments) {
		return "Comments"
	} else if strings.Contains(text, size) {
		return "Size"
	} else if strings.Contains(text, date) {
		return "Date"
	}
	return text
}

// Sample Command: transmission-cli --finish [script to run after torrent finished] --download-dir [download directory]
// Opening vlc with fullscreen
// vlc --fullscreen
