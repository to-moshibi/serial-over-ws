package main

import (
	// "bytes"
	// "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	// "net/url"
	// "strconv"
	// "strings"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
	// "go.bug.st/serial/enumerator"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Port struct {
	Name         string `json:"name"`
	IsUSB        bool   `json:"is_usb"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	Product      string `json:"product"`
}

type PortList struct {
	Ports []Port `json:"ports"`
}

type nmw_message struct {
	Com      *string          `json:"nmw_com"`
	BaudRate *int             `json:"nmw_baud_rate"`
	DataBits *int             `json:"nmw_data_bits"`
	StopBits *serial.StopBits `json:"nmw_stop_bits"`
	Parity   *serial.Parity   `json:"nmw_parity"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	cancel := make(chan struct{})
	toDevice := make(chan []byte)
	toClient := make(chan []byte)
	if r.URL.Path == "/serial/settings/" {
		if r.Method == "PUT" {
			message, err := io.ReadAll(r.Body)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error: Cannot read request body    " + err.Error()))
				return
			}
			com, mode, err := nmw_unmarshal(message[:])
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error: Cannnot unmarshal request body    " + err.Error()))
				return
			}
			resCancel := make(chan struct{})
			go serialConnect(w, *com, mode, toDevice, toClient, resCancel, cancel)
			<-resCancel
		}
	} else if r.URL.Path == "/serial/ws/" {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()
		fmt.Println("Client connected")
		go wsReader(conn, toDevice, cancel)
		go wsWriter(conn, toClient, cancel)
		<-cancel
		fmt.Println("Client disconnected")
	}
}
func serialConnect(w http.ResponseWriter, com string, mode *serial.Mode, toDevice chan []byte, toClient chan []byte, resCancel chan struct{}, cancel chan struct{}) {
	port, err := serial.Open(com, mode)
	if err != nil {
		//internal server error
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error: Cannot open serial port    " + err.Error()))
		close(resCancel)
		return
	}
	defer port.Close()
	fmt.Println("Serial port opened")
	//200 OK
	w.WriteHeader(http.StatusOK)
	close(resCancel)
	serialCancel := make(chan struct{})
	go serialReader(port, toClient, serialCancel, cancel)
	go serialWriter(port, toDevice, serialCancel, cancel)
	for {
		select {
		case <-cancel:
			fmt.Println("Serial port closed by cancel")
			return

		case <-serialCancel:
			fmt.Println("Serial port closed by serialCancel")
			return
		}
	}

}

func wsReader(conn *websocket.Conn, toDevice chan []byte, cancel chan struct{}) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			close(cancel)
			return
		}
		fmt.Println("ws message received")
		fmt.Println(string(message))
		select {
		case toDevice <- message:
		case <-cancel:
			fmt.Println("cancel wsReader")
			return
		default:
			continue
		}
	}
}

func wsWriter(conn *websocket.Conn, toClient chan []byte, cancel chan struct{}) {
	for {
		select {
		case <-cancel:
			fmt.Println("cancel wsWriter")
			return
		case message := <-toClient:
			fmt.Println(string(message))
			if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				fmt.Println(err)
				close(cancel)
				return
			}
		default:
			continue
		}
	}
}

func nmw_unmarshal(message []byte) (*string, *serial.Mode, error) {
	var msg *nmw_message
	var result *serial.Mode
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, nil, err
	}
	if msg.Com == nil {
		return nil, nil, errors.New("com port is required")
	}
	if msg.BaudRate == nil {
		return nil, nil, errors.New("baud rate is required")
	}
	if msg.DataBits == nil {
		return nil, nil, errors.New("data bits is required")
	}
	if msg.StopBits == nil {
		return nil, nil, errors.New("stop bits is required")
	}
	if msg.Parity == nil {
		return nil, nil, errors.New("parity is required")
	}

	result = &serial.Mode{
		BaudRate: *msg.BaudRate,
		DataBits: *msg.DataBits,
		StopBits: *msg.StopBits,
		Parity:   *msg.Parity,
	}
	return msg.Com, result, nil
}

func serialReader(port serial.Port, toClient chan []byte, serialCancel chan struct{}, cancel chan struct{}) {
	buff := make([]byte, 100)
	for {
		select {
		case <-cancel:
			fmt.Println("cancel serialReader by cancel")
			close(serialCancel)
			return
		case <-serialCancel:
			fmt.Println("cancel serialReader by serialCancel")
			return
		default:
			n, err := port.Read(buff)
			if err != nil {
				toClient <- []byte("Error: Serial cannot read     " + err.Error())
				fmt.Println(err)
				close(serialCancel)
				return
			}
			toClient <- buff[:n]
		}
	}
}

func serialWriter(port serial.Port, toDevice chan []byte, serialCancel chan struct{}, cancel chan struct{}) {
	for {
		select {
		case <-cancel:
			fmt.Println("cancel serialWriter")
			close(serialCancel)
			return
		case <-serialCancel:
			fmt.Println("cancel serialWriter by serialCancel")
			return
		case message := <-toDevice:
			fmt.Println("serial message received")
			fmt.Println(string(message))
			port.Write(message)
		default:
			continue
		}
	}
}

func main() {
	fmt.Println("Server started")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
