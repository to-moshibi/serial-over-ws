# HTTPでCOMポートのリストを取得
http://localhost:8080
からJSONで取ってこれます
例：
```json
{
    "ports":[
      {
        "name": "COM1",
        "is_usb": false,
        "vid": "",
        "pid": "",
        "serial_number": "",
        "product": "通信ポート (COM1)"
      },
      {
        "name": "COM3",
        "is_usb": true,
        "vid": "10C4",
        "pid": "EA60",
        "serial_number": "0001",
        "product": "Silicon Labs CP210x USB to UART Bridge (COM3)"
      }
    ]
}
```

# Server
serial-over-ws.exeを接続したいCOMがある側で起動します

serviceに登録してRestart=alwaysにしておいてください

# Client
テスト用
コメント参照

wsから来たデータをstdoutに出してます

stdinに打ち込んで改行するとwsに送信されます