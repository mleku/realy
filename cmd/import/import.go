// paketannahme stores files uploaded via HTTP on disk.
//
// It's simple enough that it can be used with curl.
// It doesn't serve the uploaded files back - it's one-way.
// It supports HTTP authentication, size and rate limiting.
package main

import (
	"crypto/subtle"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rakyll/globalconf"
)

// App contains all settings needed by our http.HandlerFuncs.
type App struct {
	url       string // Base URL to show to users
	directory string // Where to put uploaded files
	maxBytes  int64  // Maximum size for each uploaded file
	username  string // HTTP Basic Auth username
	password  string // HTTP Basic Auth password
}

// RateLimit is a token bucket based traffic policer wrapping http.HandlerFunc.
type RateLimit struct {
	bucketSize int              // Maximum number of tokens in bucket, ie. maximum burst size
	tokenDelay time.Duration    // How often to add a new token to the bucket
	tokens     chan interface{} // Returns tokens as they become available
}

// HandleRobots serves robots.txt. We don't really serve anything, so we disallow all robots.
func (app App) handleRobots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *")
	fmt.Fprintln(w, "Disallow: /")
}

// HandleIndex is our main http.HandlerFunc. It just delegates its work depending on HTTP method.
func (app App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		app.handleGet(w, r)
	} else if r.Method == "POST" {
		app.handlePost(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet returns a copy/pasteable curl command.
func (app App) handleGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, app.getCurlCommand(""))
}

// handlePost accepts an uploaded file.
func (app App) handlePost(w http.ResponseWriter, r *http.Request) {
	dateStr := time.Now().Format("2006-01-02_15:04:05")

	if r.ContentLength > app.maxBytes {
		http.Error(w, "too long", 500)
		return
	}

	var directoryName string
	var directoryPath string
	expr, _ := regexp.Compile("/([a-z0-9-_:]+)$")
	if matches := expr.FindStringSubmatch(r.RequestURI); matches != nil {
		directoryName = matches[1]
		directoryPath = path.Join(app.directory, directoryName)

		ex, err := pathExists(directoryPath)
		if err != nil {
			http.Error(w, "directory doesn't exist", 404)
			return
		} else if !ex {
			directoryName = matches[1]
			directoryPath = path.Join(app.directory, directoryName)
		}
	} else {
		directoryName = dateStr + "_" + randomString(8)
		directoryPath = path.Join(app.directory, directoryName)
		os.Mkdir(directoryPath, 0755)
	}

	fileName := dateStr + "_" + randomString(8) + ".bin"
	filePath := path.Join(directoryPath, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "error", 500)
		return
	}
	defer out.Close()

	reader := io.LimitReader(r.Body, app.maxBytes)
	written, err := io.Copy(out, reader)
	if err != nil {
		http.Error(w, "error", 500)
		return
	}

	fmt.Fprintf(w, "Accepted %d bytes.\n", written)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Folder: %s\n", directoryName)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Add related files? Pipe to:\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintln(w, app.getCurlCommand(directoryName))
}

// getCurlCommand returns a copy/pasteable command one can pipe to to upload files.
// Careful: this includes the authentication credentials!
func (app App) getCurlCommand(subdir string) string {
	result := "curl"

	if app.username != "" && app.password != "" {
		result += " -u" + strconv.Quote(app.username) + ":" + strconv.Quote(app.password)
	}

	result += " " + app.url

	if subdir != "" {
		result += "/" + subdir
	}

	result += " --data-binary @-"

	return result
}

// Protect does HTTP basic authentication for an http.HandlerFunc.
func Protect(expectedUsername, expectedPassword []byte,
	handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader, hasAuthHeader := r.Header["Authorization"]
		if !hasAuthHeader || len(authHeader) != 1 {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}
		auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
		}

		actualUsername := []byte(pair[0])
		actualPassword := []byte(pair[1])

		if 1 != ConstantTimeAnd(
			ConstantTimeSeqEq(expectedUsername, actualUsername),
			ConstantTimeSeqEq(expectedPassword, actualPassword)) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}

}

// ConstantTimeAnd Returns 1 iff no value equals 0, 0 otherwise.
// The runtime is independent from the values.
func ConstantTimeAnd(values ...int) int {
	result := 1
	for _, value := range values {
		result &= int(value)
	}
	return result
}

// ConstantTimeSeqEq Returns 1 if expected == actual, 0 otherwise.
// The runtime is independent from the length and content of the actual sequence.
func ConstantTimeSeqEq(expected, actual []byte) int {
	if len(expected) > 255 || len(actual) > 255 || len(expected) == 0 || len(actual) == 0 {
		return 0
	}

	actualLen := len(actual)
	expectedLen := len(expected)

	result := uint8(expectedLen) ^ uint8(actualLen)
	for i := range expected {
		result |= expected[i] ^ actual[i%actualLen]
	}

	return subtle.ConstantTimeByteEq(result, 0)
}

func NewRateLimit(delay, burst int) RateLimit {
	result := RateLimit{
		tokenDelay: time.Duration(delay) * time.Second,
		bucketSize: burst,
		tokens:     make(chan interface{}, burst),
	}

	// Initially, the bucket should be full
	for i := 1; i <= result.bucketSize; i++ {
		result.tokens <- nil
	}

	// Add tokens to bucket regularly (unless it's already full)
	go func() {
		for {
			time.Sleep(result.tokenDelay)
			result.tokens <- nil
		}
	}()

	return result
}

func (rl RateLimit) Limit(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		select {
		case <-rl.tokens:
			handler.ServeHTTP(writer, request)
		default:
			writer.WriteHeader(429) // too many requests
		}
	}
}

// randomString is used to generate file/directory names.
func randomString(length int) string {
	var letters = []rune("0123456789abcdef")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// pathExists returns true iff path exists in the filesystem.
func pathExists(path string) (result bool, err error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func main() {
	var addr = flag.String("addr", "", "IP address to bind to (default: any)")
	var port = flag.Int("port", 8080, "TCP port to bind to")
	var directory = flag.String("directory", "", "path to directory to drop files in")
	var urlPtr = flag.String("url", "", "base url to tell clients")
	var delay = flag.Int("delay", 30, "rate limit: seconds per request")
	var burst = flag.Int("burst", 16, "rate limit: maximum burst size")
	var username = flag.String("user", "", "username for HTTP authentication (dangerous!)")
	var password = flag.String("password", "", "password for HTTP authentication (dangerous!)")
	var maxByte = flag.Int("size-limit", 1024, "maximum file size, in kB")
	var config = flag.String("config", "", "config file path")
	flag.Parse()

	if *config != "" {
		conf, err := globalconf.NewWithOptions(&globalconf.Options{
			Filename: *config,
		})
		if err != nil {
			panic(err)
		}
		conf.ParseAll()
	}

	var url string
	if *urlPtr != "" {
		url = *urlPtr
	} else if *addr != "" {
		url = "http://" + *addr + ":" + strconv.Itoa(*port)
	} else {
		url = "http://localhost:" + strconv.Itoa(*port)
	}

	app := App{
		url:       url,
		directory: *directory,
		maxBytes:  int64(*maxByte) * 1000,
		username:  *username,
		password:  *password,
	}

	// If credentials have been supplied, use HTTP authentication
	ProtectMaybe := func(f http.HandlerFunc) http.HandlerFunc {
		if *username != "" && *password != "" {
			return Protect([]byte(*username), []byte(*password), f)
		} else {
			return f
		}
	}

	http.HandleFunc("/robots.txt", app.handleRobots)
	http.HandleFunc("/", NewRateLimit(*delay, *burst).Limit(ProtectMaybe(app.handleIndex)))
	err := http.ListenAndServe(string(*addr)+":"+strconv.Itoa(*port), nil)
	if err != nil {
		panic(err)
	}
}
