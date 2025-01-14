package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
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
	BaudRate int `json:"nmw_baud_rate"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set( "Access-Control-Allow-Methods","GET, POST, PUT, DELETE, OPTIONS" )
	if r.URL.Path == "/" {
		var portList PortList

		ports, err := enumerator.GetDetailedPortsList()
		if err != nil {
			fmt.Println(err)
		}
		if len(ports) == 0 {
			fmt.Println("No serial ports found!")
		}

		for _, port := range ports {
			portList.Ports = append(portList.Ports, Port{
				Name:         strings.TrimPrefix(port.Name, "/dev/"), 
				IsUSB:        port.IsUSB,
				VID:          port.VID,
				PID:          port.PID,
				SerialNumber: port.SerialNumber,
				Product:      port.Product,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(portList); err != nil {
			fmt.Println(err)
		}
		fmt.Fprint(w, buf.String())
	} else {
		//websocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()

		com := r.URL.Path[1:]
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Invalid Query \n"+err.Error()))
			fmt.Println(err)
		}

		baudRate := 9600
		dataBits := 8
		parity := serial.NoParity
		stopBits := serial.OneStopBit

		//paramsから設定を取得
		fmt.Println(params)
		for key, value := range params {
			if key == "baud_rate" {
				baudRate, err = strconv.Atoi(value[0])
				if err != nil {
					conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Invalid baudRate \n"+err.Error()))
					fmt.Println("Error Parsing baudRate", err)
				}
			}
			if key == "data_bits" {
				dataBits, err = strconv.Atoi(value[0])
				if err != nil {
					conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Invalid dataBits \n"+err.Error()))
					fmt.Println("Error Parsing data_bits", err)
				}
			}
			if key == "parity" {
				if value[0] == "none" {
					parity = serial.NoParity
				} else if value[0] == "odd" {
					parity = serial.OddParity
				} else if value[0] == "even" {
					parity = serial.EvenParity
				} else if value[0] == "mark" {
					parity = serial.MarkParity
				} else if value[0] == "space" {
					parity = serial.SpaceParity
				} else {
					conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Invalid parity \n"+value[0]))
					fmt.Println("Error Parsing parity", err)
				}
			}
			if key == "stop_bits" {
				if value[0] == "1" {
					stopBits = serial.OneStopBit
				} else if value[0] == "1.5" {
					stopBits = serial.OnePointFiveStopBits
				} else if value[0] == "2" {
					stopBits = serial.TwoStopBits
				} else {
					conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Invalid stopBits \n"+value[0]))
					fmt.Println("Error Parsing stop_bits", err)
				}
			}
		}

		//シリアルポートを開く
		mode := &serial.Mode{
			BaudRate: baudRate,
			DataBits: dataBits,
			Parity:   parity,
			StopBits: stopBits,
		}
		port, err := serial.Open("/dev/"+com, mode)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Serial cannot open \n"+err.Error()))
			fmt.Println(err)
		}
		go writer(conn, port)
		reader(conn, port)
	}
}

func writer(conn *websocket.Conn, port serial.Port) {
	buff := make([]byte, 100)
	for {
		n, err := port.Read(buff)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("Error: Serial cannot read \n"+err.Error()))
			fmt.Println(err)
		}
		// fmt.Printf("%v", string(buff[:n]))
		if wserr := conn.WriteMessage(websocket.BinaryMessage, buff[:n]); wserr != nil {
			break
		}
	}
}

func reader(conn *websocket.Conn, port serial.Port) {
	for {
		_, message, err := conn.ReadMessage()
		var msg nmw_message
		if err := json.Unmarshal(message, &msg); err == nil {
			mode := &serial.Mode{
				BaudRate: msg.BaudRate,
			}
			port.SetMode(mode)
			continue
		}
		if err != nil {
			port.Close()
			conn.Close()
			fmt.Println("connection closed")
			fmt.Println(err)
			break
		}
		port.Write(message)
	}
}

func main() {
	fmt.Println("Server started")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
