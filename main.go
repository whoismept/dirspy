package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type ColorMode struct {
	enabled bool
	red     string
	green   string
	yellow  string
	blue    string
	purple  string
	reset   string
}

func newColorMode(enabled bool) ColorMode {
	if enabled {
		return ColorMode{
			enabled: true,
			red:     "\033[31m",
			green:   "\033[32m",
			yellow:  "\033[33m",
			blue:    "\033[34m",
			purple:  "\033[35m",
			reset:   "\033[0m",
		}
	}
	return ColorMode{
		enabled: false,
		red:     "",
		green:   "",
		yellow:  "",
		blue:    "",
		purple:  "",
		reset:   "",
	}
}

func (c ColorMode) Red(s string) string {
	if c.enabled {
		return c.red + s + c.reset
	}
	return s
}

func (c ColorMode) Green(s string) string {
	if c.enabled {
		return c.green + s + c.reset
	}
	return s
}

func (c ColorMode) Yellow(s string) string {
	if c.enabled {
		return c.yellow + s + c.reset
	}
	return s
}

func (c ColorMode) Blue(s string) string {
	if c.enabled {
		return c.blue + s + c.reset
	}
	return s
}

func (c ColorMode) Purple(s string) string {
	if c.enabled {
		return c.purple + s + c.reset
	}
	return s
}

type FileInfo struct {
	URL      string
	Size     int64
	Keywords []string
}

func parseStatusCodes(codes string) map[int]bool {
	ignoreCodes := make(map[int]bool)
	if codes == "" {
		return ignoreCodes
	}

	for _, code := range strings.Split(codes, ",") {
		code = strings.TrimSpace(code)
		if statusCode, err := strconv.Atoi(code); err == nil {
			ignoreCodes[statusCode] = true
		}
	}
	return ignoreCodes
}

func parseKeywords(keywords string) []string {
	if keywords == "" {
		return nil
	}
	var result []string
	for _, kw := range strings.Split(keywords, ",") {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			result = append(result, kw)
		}
	}
	return result
}

func searchKeywords(content string, keywords []string) []string {
	if len(keywords) == 0 {
		return nil
	}

	var found []string
	contentLower := strings.ToLower(content)

	for _, kw := range keywords {
		if strings.Contains(contentLower, strings.ToLower(kw)) {
			found = append(found, kw)
		}
	}
	return found
}

func main() {
	// Define command line flags
	baseURL := flag.String("u", "", "Base URL to crawl (required)")
	ignoreCodes := flag.String("i", "", "Comma-separated HTTP status codes to ignore (e.g., '404,403,500')")
	keywords := flag.String("k", "", "Comma-separated keywords to search in files (e.g., 'password,secret,key')")
	extensions := flag.String("e", "", "Comma-separated file extensions to ignore (e.g., '.txt,.jpg')")
	noColor := flag.Bool("c", false, "Disable colored output")
	proxy := flag.String("p", "", "Proxy URL (default: http://localhost:8080)")
	flag.Parse()

	// Initialize color mode
	colors := newColorMode(!*noColor)

	// Check if URL is provided
	if *baseURL == "" {
		fmt.Println("Error: URL parameter is required")
		fmt.Println("Usage: ./dirspy -u=http://example.com/ [-i='404,403'] [-k='password,secret'] [-e='.txt,.jpg'] [-c] [-p='http://localhost:8080']")
		fmt.Println("Examples:")
		fmt.Println("  ./dirspy -url=http://example.com/")
		fmt.Println("  ./dirspy -url=http://example.com/ -i='403,404' -k='password,api_key'")
		fmt.Println("  ./dirspy -url=http://example.com/ -e='.txt,.jpg'")
		fmt.Println("  ./dirspy -url=http://example.com/ -c")
		os.Exit(1)
	}

	// Initialize HTTP client with proxy if provided
	var httpClient *http.Client
	if *proxy != "" {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			fmt.Printf("Invalid proxy URL: %v\n", err)
			os.Exit(1)
		}
		httpClient = &http.Client{Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	} else {
		httpClient = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	}

	// Parse parameters
	ignoreMap := parseStatusCodes(*ignoreCodes)
	keywordsList := parseKeywords(*keywords)
	extList := *extensions

	files := make(map[string]FileInfo)
	visited := make(map[string]bool)

	crawl(*baseURL, *baseURL, files, visited, ignoreMap, keywordsList, extList, colors, httpClient)

	fmt.Printf("\n%s\n", colors.Purple("Results:"))
	for fileURL, info := range files {
		resultLine := fmt.Sprintf("%s: %d bytes", fileURL, info.Size)
		if len(info.Keywords) > 0 {
			resultLine += fmt.Sprintf(" %s", colors.Blue(fmt.Sprintf("[FOUND KEYWORDS: %s]", strings.Join(info.Keywords, ", "))))
		}
		fmt.Printf("%s%s%s\n", colors.Purple("-> "), resultLine, colors.reset)
	}
}

func ignoreFileExtension(fileURL string, extList string) bool {
	if extList == "" {
		return false
	}
	extensions := strings.Split(extList, ",")
	for _, ext := range extensions {
		ext = strings.TrimSpace(ext)
		if strings.HasSuffix(fileURL, ext) {
			return true
		}
	}
	return false
}

func crawl(baseURL, currentURL string, files map[string]FileInfo, visited map[string]bool,
	ignoreCodes map[int]bool, keywords []string, extList string, colors ColorMode, httpClient *http.Client) {
	if !ignoreFileExtension(currentURL, extList) {
		if visited[currentURL] {
			return
		}

		visited[currentURL] = true
		fmt.Printf("Crawling: %s\n", currentURL)

		resp, err := httpClient.Get(currentURL)
		if err != nil {
			fmt.Printf("%s\n", colors.Red(fmt.Sprintf("Access error %s: %v", currentURL, err)))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if !ignoreCodes[resp.StatusCode] {
				fmt.Printf("%s\n", colors.Red(fmt.Sprintf("Invalid status code %s: %d", currentURL, resp.StatusCode)))
			}
			if ignoreCodes[resp.StatusCode] {
				return
			}
		} else {
			fmt.Printf("%s\n", colors.Green(fmt.Sprintf("[200 OK] %s", currentURL)))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s\n", colors.Red(fmt.Sprintf("Error reading body %s: %v", currentURL, err)))
			return
		}
		bodyContent := string(bodyBytes)

		doc, err := html.Parse(strings.NewReader(bodyContent))
		if err != nil {
			fmt.Printf("%s\n", colors.Red(fmt.Sprintf("HTML parsing error %s: %v", currentURL, err)))
			return
		}

		foundKeywords := searchKeywords(bodyContent, keywords)
		if len(foundKeywords) > 0 {
			fmt.Printf("%s\n", colors.Blue(fmt.Sprintf("Found keywords in %s: %s", currentURL, strings.Join(foundKeywords, ", "))))
		}

		var processNode func(*html.Node)
		processNode = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						href := attr.Val
						if !strings.HasPrefix(href, "http") {
							baseURLObj, _ := url.Parse(baseURL)
							relativeURL, err := url.Parse(href)
							if err != nil {
								continue
							}
							href = baseURLObj.ResolveReference(relativeURL).String()
						}

						if strings.HasSuffix(href, "/") && strings.HasPrefix(href, baseURL) {
							crawl(baseURL, href, files, visited, ignoreCodes, keywords, extList, colors, httpClient)
						} else if strings.HasPrefix(href, baseURL) {
							fileResp, err := httpClient.Get(href)
							if err != nil {
								fmt.Printf("%s\n", colors.Red(fmt.Sprintf("File access error %s: %v", href, err)))
								continue
							}

							bodyBytes, err := io.ReadAll(fileResp.Body)
							fileResp.Body.Close()

							if err != nil {
								fmt.Printf("%s\n", colors.Red(fmt.Sprintf("Error reading file %s: %v", href, err)))
								continue
							}

							if fileResp.StatusCode == http.StatusOK {
								if !ignoreFileExtension(href, extList) {
									foundKeywords := searchKeywords(string(bodyBytes), keywords)
									files[href] = FileInfo{
										URL:      href,
										Size:     int64(len(bodyBytes)),
										Keywords: foundKeywords,
									}

									statusMsg := fmt.Sprintf("[200 OK] %s (%d bytes)", href, len(bodyBytes))
									if len(foundKeywords) > 0 {
										statusMsg += fmt.Sprintf(" [FOUND: %s]", strings.Join(foundKeywords, ", "))
										statusMsg = colors.Green(statusMsg) + " " + colors.Blue("[KEYWORDS FOUND]")
									} else {
										statusMsg = colors.Green(statusMsg)
									}
									fmt.Println(statusMsg)
								}
							} else if !ignoreCodes[fileResp.StatusCode] {
								fmt.Printf("%s\n", colors.Yellow(fmt.Sprintf("[%d] %s", fileResp.StatusCode, href)))
							}
						}
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c)
			}
		}

		processNode(doc)
	}
}
