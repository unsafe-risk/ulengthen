package main

import (
	"embed"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lemon-mint/envaddr"
	"github.com/lemon-mint/godotenv"
	"github.com/lemon-mint/vbox"
	"github.com/unsafe-risk/ulengthen/proto"
	"github.com/valyala/bytebufferpool"
)

var box *vbox.BlackBox = vbox.NewBlackBox([]byte(
	"my super secret key",
))

func main() {
	godotenv.Load()
	{ // Initialize the VBox
		secretKey := os.Getenv("SECRET_KEY")
		if secretKey != "" {
			box = vbox.NewBlackBox([]byte(secretKey))
		}
	}
	h := &URLLengthenerHandler{}
	h.Init()
	srv := http.Server{
		Handler: h,
	}
	ln, err := net.Listen("tcp", envaddr.Get(":8080"))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Printf("Listening on %s\n", ln.Addr())
	go srv.Serve(ln)

	{
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("\nShutting down...")
		err = srv.Shutdown(nil)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Println("Done.")
	}
}

//go:embed public index.html
var publicFS embed.FS

type URLLengthenerHandler struct {
	urlbufpool       bytebufferpool.Pool
	publicFileServer http.Handler
}

var URLEncoding = base32.NewEncoding("0123456789ABCDEFGHJKNOPQRSTUVXYZ").
	WithPadding(base32.NoPadding)

func (h *URLLengthenerHandler) Init() {
	h.publicFileServer = http.FileServer(http.FS(publicFS))
}

func (h *URLLengthenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		req := r.Header.Get("Access-Control-Request-Method")
		if req != "" {
			w.Header().Set("Access-Control-Allow-Methods", req)
		}
		req = r.Header.Get("Access-Control-Request-Headers")
		if req != "" {
			w.Header().Set("Access-Control-Allow-Headers", req)
		}
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	path := r.URL.Path
	switch path {
	case "/", "":
		h.publicFileServer.ServeHTTP(w, r)
		return
	case "/new", "/new/":
		if r.Method != "POST" {
			http.Error(w, "Method not allowed, Use POST", http.StatusMethodNotAllowed)
			return
		}

		type Request struct {
			URL string `json:"url"`
		}
		req := Request{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		t := time.Now().UnixMilli()
		urlinfo := proto.URLInfoFromVTPool()
		urlinfo.Url = req.URL
		urlinfo.Timestamp = uint64(t)

		urlbuf := h.urlbufpool.Get()
		urlsize := urlinfo.SizeVT()
		if urlsize > cap(urlbuf.B) {
			urlbuf.B = make([]byte, urlsize)
		}
		urlbuf.B = urlbuf.B[:urlsize]
		urlinfo.MarshalToSizedBufferVT(urlbuf.B)
		urlinfo.ReturnToVTPool()
		e := box.Seal(urlbuf.B)
		h.urlbufpool.Put(urlbuf)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		type Response struct {
			Data string `json:"data"`
		}
		resp := Response{
			Data: URLEncoding.EncodeToString(e),
		}
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		if strings.HasPrefix(path, "/public/") {
			h.publicFileServer.ServeHTTP(w, r)
			return
		}

		TrimmedPath := strings.TrimPrefix(path, "/")
		decoded, err := URLEncoding.DecodeString(TrimmedPath)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		decrypted, ok := box.OpenOverWrite(decoded)
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		urlinfo := proto.URLInfoFromVTPool()
		defer urlinfo.ReturnToVTPool()
		err = urlinfo.UnmarshalVT(decrypted)
		if err != nil {
			urlinfo.ReturnToVTPool()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if urlinfo.GetRequireCaptcha() {
			// TODO: Implement captcha
		} else if urlinfo.GetRequirePassword() {
			// TODO: Implement password
		} else {
			http.Redirect(w, r, urlinfo.GetUrl(), http.StatusTemporaryRedirect)
		}
	}
}
