package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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

func handler(w http.ResponseWriter, r *http.Request) {
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
				Name:         port.Name,
				IsUSB:        port.IsUSB,
				VID:          port.VID,
				PID:          port.PID,
				SerialNumber: port.SerialNumber,
				Product:      port.Product,
			})

			fmt.Printf("Found port: %s\n", port.Name)
			if port.IsUSB {
				fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
				fmt.Printf("   USB serial %s\n", port.SerialNumber)

				//OS依存なので、ラズパイで動くか注意
				fmt.Printf("   USB Product %s\n", port.Product)

			}
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
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid Query \n"+err.Error()))
			fmt.Println(err)
		}

		baudRate := 115200
		dataBits := 8
		parity := serial.NoParity
		stopBits := serial.OneStopBit

		//paramsから設定を取得
		fmt.Println(params)
		for key, value := range params {
			if key == "baud_rate" {
				baudRate, err = strconv.Atoi(value[0])
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid baudRate \n"+err.Error()))
					fmt.Println("Error Parsing baudRate", err)
				}
			}
			if key == "data_bits" {
				dataBits, err = strconv.Atoi(value[0])
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid dataBits \n"+err.Error()))
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
					conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid parity \n"+value[0]))
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
					conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid stopBits \n"+value[0]))
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
		port, err := serial.Open(com, mode)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Serial cannot open \n"+err.Error()))
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
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Serial cannot read \n"+err.Error()))
			fmt.Println(err)
		}
		// if n == 0 {
		// 	fmt.Println("\nEOF")
		// 	break
		// }
		fmt.Printf("%v", string(buff[:n]))
		if wserr := conn.WriteMessage(websocket.TextMessage, buff[:n]); wserr != nil {
			fmt.Println(err)
			break
		}
	}
}

func reader(conn *websocket.Conn, port serial.Port) {
	for {
		_, message, err := conn.ReadMessage()
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
	fmt.Println("Hello, World!")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
