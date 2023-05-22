package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var file *string = flag.String("file", "", "Specifies an absolute location of a single text file containing a word list. (Required)")
var plt *string = flag.String("plt", "IGWT", "Specifies a plate design to check against. Different plates have different character lengths and reg. requirements. \nDefaults to \"In God We Trust\" (7.5 Char Limit).")
var delay *int = flag.Int("d", 1500, "Specifies a delay (in milliseconds) between each call to the DMV website. \nIf running without proxies or with a large word list, the delay value should be higher. Defaults to 1500ms")
var retry *int = flag.Int("r", 3000, "Specifies a delay (in milliseconds) to wait if a check is unsuccessful. \nIf running with proxies, the retry value can be much lower. Defaults to 3000ms")
var pfile *string = flag.String("proxy", "", "Specifies an absolute location of a text file containing a proxy list. \nProxies should be in ip:port:user:pass format. Defaults to use localhost for all tasks.")

var wg sync.WaitGroup

var proxies [][]string
var pltlendec, displen int

func main() {
	flag.Parse()
	if *file != "" {
		logo()
		if *pfile != "" {
			getProxies(*pfile)
		}
		getPltLen()
		withFile(*file)
	} else {
		flag.Usage()
	}
}

func withFile(s string) {
	conv := filepath.ToSlash(s)
	_, err := os.Stat(conv)
	if err != nil {
		log.Fatal("Error finding word list. Either the path is not a valid file, or it is not absolute.")
	}
	words, err := readWords(conv)
	if err != nil {
		log.Fatal(err)
	}
	for i := range words {
		wg.Add(1)
		time.Sleep(time.Duration(*delay) * time.Millisecond)
		go postForm(words[i])
	}
	wg.Wait()
}

func getProxies(s string) {
	conv := filepath.ToSlash(s)
	_, err := os.Stat(conv)
	if err != nil {
		log.Fatal("Error finding proxy list. Either the path is not a valid file, or it is not absolute.")
	}
	proxies, err = readProxies(conv)
	if err != nil {
		log.Fatal(err)
	}
}

func readProxies(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var proxies [][]string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		proxy := scanner.Text()
		if proxy == "" {
			continue
		}
		split := strings.Split(proxy, ":")
		if len(split) != 4 {
			return nil, errors.New("incorrect proxy format")
		}
		proxies = append(proxies, split)
	}
	return proxies, scanner.Err()
}

func readWords(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		word := scanner.Text()
		if validateWord(word) {
			lines = append(lines, word)
		}
	}
	return lines, scanner.Err()
}

func validateWord(word string) bool {
	l := 0
	for i := range word {
		if strings.Contains("abcdefghijklmnopqrstuvwxyz0123456789", string(word[i])) {
			l += 10
		} else if strings.Contains("&- ", string(word[i])) {
			l += 5
		} else {
			return false
		}
	}
	if l <= pltlendec {
		return true
	}
	return false
}

func getPltLen() {
	var data = strings.NewReader(`PltType=` + *plt + `&HoldISA=N&PersonalMsg=&PltNo=&SavePltNo=&HoldLet1=&HoldLet2=&HoldLet3=&HoldLet4=&HoldLet5=&HoldLet6=&HoldLet7=&HoldLet8=&Message10=&Message15=&Message20=&Message25=&Message30=&Message35=&Message40=&Message45=&Message50=&Message55=&Message60=&Message65=&Message70=&Message75=&Message80=&TransStat10=&TransStat15=&TransStat20=&TransStat25=&TransStat30=&TransStat35=&TransStat40=&TransStat45=&TransStat50=&TransStat55=&TransStat60=&TransStat65=&TransStat70=&TransStat75=&TransStat80=`)
	req, err := http.NewRequest("POST", "https://www.dmv.virginia.gov/dmvnet/plate_purchase/s2end.asp", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://www.dmv.virginia.gov")
	req.Header.Set("Referer", "https://www.dmv.virginia.gov/dmvnet/plate_purchase/s2plttype.asp?PLT=&PLTNO=")
	req.Header.Set("Sec-Fetch-Dest", "frame")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	z := html.NewTokenizer(res.Body)
	tt := z.Next()
	for {
		tt = z.Next()
		if tt == html.ErrorToken {
			break
		}
		cur := z.Token()
		if tt == html.StartTagToken && cur.Data == "input" {
			for _, v := range cur.Attr {
				if v.Key == "name" {
					if v.Val == "PltChars" {
						for _, w := range cur.Attr {
							if w.Key == "value" {
								pltlendec, err = strconv.Atoi(w.Val)
								if err != nil {
									log.Fatal("Plate Length Not Integer")
								}
								break
							}
						}
						break
					}
					if v.Val == "NumCharsInt" {
						for _, w := range cur.Attr {
							if w.Key == "value" {
								displen, err = strconv.Atoi(w.Val)
								if err != nil {
									log.Fatal("Rounded Plate Length Not Integer")
								}
								break
							}
						}
						break
					}
				}
			}
		}
	}
}

func postForm(word string) {
	defer wg.Done()
	var client *http.Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(proxies) > 0 {
		singleproxy := proxies[rand.Intn(len(proxies))]
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{
					Scheme: "http",
					User:   url.UserPassword(singleproxy[2], singleproxy[3]),
					Host:   singleproxy[0] + ":" + singleproxy[1],
				}),
			},
			Jar: jar,
		}
	} else {
		client = &http.Client{
			Jar: jar,
		}
	}
	bufferedword := word
	for i := 0; i < displen; i++ {
		if i >= len(word) {
			bufferedword += "+"
		}
	}
	var data = strings.NewReader(`TransType=INQ&TransID=RESINQ&ReturnPage=%2Fdmvnet%2Fplate_purchase%2Fs2end.asp&HelpPage=&Choice=A&PltNo=` + bufferedword + `&HoldISA=N&HoldSavePltNo=&HoldCallHost=&NumCharsInt=` + fmt.Sprint(displen) + `&CurrentTrans=plate_purchase_reserve&PltType=` + *plt + `&PltNoAvail=&PersonalMsg=Y`)
	init, err := http.NewRequest("GET", "https://www.dmv.virginia.gov/dmvnet/plate_purchase/select_plate.asp", nil)
	if err != nil {
		log.Fatal(err)
	}
	init.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	init.Header.Set("Accept-Language", "en-US,en;q=0.9")
	init.Header.Set("Cache-Control", "max-age=0")
	init.Header.Set("Connection", "keep-alive")
	init.Header.Set("Referer", "https://www.dmv.virginia.gov/vehicles/")
	init.Header.Set("Sec-Fetch-Dest", "frame")
	init.Header.Set("Sec-Fetch-Mode", "navigate")
	init.Header.Set("Sec-Fetch-Site", "same-origin")
	init.Header.Set("Sec-Fetch-User", "?1")
	init.Header.Set("Upgrade-Insecure-Requests", "1")
	init.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	res, err := client.Do(init)
	if err != nil {
		log.Println("Request Error")
		time.Sleep(time.Duration(*retry) * time.Millisecond)
		wg.Add(1)
		postForm(word)
		return
	}
	res.Body.Close()
	req, err := http.NewRequest("POST", "https://www.dmv.virginia.gov/dmvnet/common/router.asp", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://www.dmv.virginia.gov")
	req.Header.Set("Referer", "https://www.dmv.virginia.gov/dmvnet/plate_purchase/s2end.asp")
	req.Header.Set("Sec-Fetch-Dest", "frame")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	res, err = client.Do(req)
	if err != nil {
		log.Println("Request Error")
		time.Sleep(time.Duration(*retry) * time.Millisecond)
		wg.Add(1)
		postForm(word)
		return
	}
	defer res.Body.Close()
	z := html.NewTokenizer(res.Body)
	tt := z.Next()
	for {
		tt = z.Next()
		if tt == html.ErrorToken {
			break
		}
		cur := z.Token()
		if tt == html.TextToken {
			if cur.Data == "Congratulations.  The message you requested is available." {
				writeSuccess(word)
				log.Println("Word " + word + " is Available!")
				return
			}
			if cur.Data == "If you have reserved this message or it is on a vehicle you own, click Purchase Plate Now; if not, try a new message." {
				return
			}
		}
	}
	log.Println("No Response, Regenning Cookies..")
	time.Sleep(time.Duration(*retry) * time.Millisecond)
	wg.Add(1)
	postForm(word)
}

func writeSuccess(word string) {
	file, err := os.OpenFile("./available.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(word + "\n")
	if err != nil {
		log.Fatal(err)
	}
}

func logo() {
	color.New(color.FgCyan).Add(color.Bold).Print("\n\n   ____  ______   ____  __      __           \n  / __ \\/ ____/  / __ \\/ /___ _/ /____  _____\n / / / / / __   / /_/ / / __ `/ __/ _ \\/ ___/\n/ /_/ / /_/ /  / ____/ / /_/ / /_/  __(__  ) \n\\____/\\____/  /_/   /_/\\__,_/\\__/\\___/____/  \n\n")
}
