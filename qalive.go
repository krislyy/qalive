package main

import (
	"net"
	"net/http"
	"encoding/json"
	"log"
	"flag"
	"github.com/krislyy/qalive/configure"
	"io/ioutil"
	"fmt"
	"github.com/krislyy/qalive/core"
	"github.com/krislyy/qalive/rtmp"
)


var (
	operaAddr      = flag.String("manage-addr", ":8095", "HTTP manage interface server listen address")
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	flag.Parse()
}

type Response struct {
	w       http.ResponseWriter
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	return r.w.Write(resp)
}

type Server struct {
	Config *configure.Configure
}

func NewServer() *Server {
	return &Server{
	}
}

func (s *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/control/asynccopy", func(w http.ResponseWriter, r *http.Request) {
		s.handleAsyncCopy(w, r)
	})
	mux.HandleFunc("/control/push", func(w http.ResponseWriter, r *http.Request) {
		s.handlePush(w, r)
	})
	mux.HandleFunc("/control/stop", func(w http.ResponseWriter, r *http.Request) {
		s.handleStop(w, r)
	})

	http.Serve(l, mux)
	return nil
}

func (s *Server)handleAsyncCopy(w http.ResponseWriter, r *http.Request)  {
	var err error
	response := &Response{
		w: w,
		Status: 200,
		Message: "Asyc Copy done!",
	}
	defer response.SendJson()

	s.Config = configure.NewConfigure()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Status = 300
		response.Message = fmt.Sprintf("Read Request body error[%s].", err.Error())
		return
	}
	log.Printf("%s", body)

	if err = json.Unmarshal(body, s.Config); err != nil {
		response.Status = 301
		response.Message = fmt.Sprintf("Json unmarshal error[%s].", err.Error())
		return
	}

	if !s.Config.IsCopyValid() {
		response.Status = 302
		response.Message = "Invalid Parameters!"
		return
	}

	go core.AsyncCopyPackets(s.Config)
}

func (s *Server)handlePush(w http.ResponseWriter, r *http.Request)  {
	var err error
	response := &Response{
		w: w,
		Status: 200,
		Message: "Rtmp push stream success!",
	}
	defer response.SendJson()

	s.Config = configure.NewConfigure()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Status = 300
		response.Message = fmt.Sprintf("Read Request body error[%s].", err.Error())
		return
	}
	log.Printf("%s", body)

	if err = json.Unmarshal(body, s.Config); err != nil {
		response.Status = 301
		response.Message = fmt.Sprintf("Json unmarshal error[%s].", err.Error())
		return
	}

	if !s.Config.IsPushValid() {
		response.Status = 302
		response.Message = "Invalid Parameters!"
		return
	}

	go rtmp.RTMP_Publish(s.Config)
}

func (s *Server)handleStop(w http.ResponseWriter, r *http.Request)  {
	var err error
	response := &Response{
		w: w,
		Status: 200,
		Message: "Stop rtmp stream success!",
	}
	defer response.SendJson()

	s.Config = configure.NewConfigure()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Status = 300
		response.Message = fmt.Sprintf("Read Request body error[%s].", err.Error())
		return
	}
	log.Printf("%s", body)

	if err = json.Unmarshal(body, s.Config); err != nil {
		response.Status = 301
		response.Message = fmt.Sprintf("Json unmarshal error[%s].", err.Error())
		return
	}

	if s.Config.Crtmp_url == "" || s.Config.StreamName == "" {
		response.Status = 302
		response.Message = "Invalid Parameters!"
		return
	}

	go rtmp.RTMP_Stop(s.Config)
}

func main() {
	opListen, err := net.Listen("tcp", *operaAddr)
	defer opListen.Close()
	if err != nil {
		log.Fatal(err)
	}
	opServer := NewServer()
	defer func() {
		if r := recover(); r != nil {
			log.Println("HTTP-Operation server panic: ", r)
		}
	}()
	log.Println("HTTP-Operation listen On", *operaAddr)
	opServer.Serve(opListen)
}
