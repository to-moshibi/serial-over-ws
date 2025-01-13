process.stdin.setEncoding("utf8");

// ws://localhost:8080/{COMポート名}?baud_rate={ボーレート}&data_bits={データビット数}&stop_bits={ストップビット数}&parity={パリティ}
// !!Linuxでは、COMポート名は/dev/以下にあるデバイス名を指定してください。!!
// 例: ws://localhost:8080/ttyUSB0?baud_rate=115200&data_bits=8&stop_bits=1&parity=none
var connection = new WebSocket('ws://localhost:8080/COM3?baud_rate=115200&data_bits=8&stop_bits=1&parity=none');



// パラメータが指定されていない時のデフォルト値
// baudRate := 9600
// dataBits := 8
// parity := serial.NoParity
// stopBits := serial.OneStopBit


//resetスイッチを押すと、たまにwebsocketが切れるので、再接続する

//許可される値
const data_bits =[5,6,7,8]
const stop_bits = [1,1.5,2]
const parity = ['none','even','odd','mark','space']

connection.onopen = function () {
    // connection.send('Hello, Server!');
}
connection.onerror = function (error) {
    console.log('WebSocket Error ');
    console.log(error);
}
connection.onmessage = function (e) {
    process.stdout.write(e.data);
}
connection.onclose = function () {
    console.log('WebSocket connection closed');
}

process.stdin.setEncoding("utf8");


var reader = require("readline").createInterface({
  input: process.stdin
});

reader.on("line", (line) => {
    connection.send(line);
});
reader.on("close", () => {
  console.log("stdin closed"); 
});