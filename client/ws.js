process.stdin.setEncoding("utf8");

// ws://localhost:8080/{COMポート名}?baud_rate={ボーレート}&data_bits={データビット数}&stop_bits={ストップビット数}&parity={パリティ}
// !!Linuxでは、COMポート名は/dev/以下にあるデバイス名を指定してください。!!
// 例: ws://localhost:8080/ttyUSB0?baud_rate=115200&data_bits=8&stop_bits=1&parity=none
var connection = new WebSocket('ws://localhost:8080/serial/ws/');


//resetスイッチを押すと、たまにwebsocketが切れるので、再接続する

//許可される値
const data_bits =[5,6,7,8]
const stop_bits = [1,1.5,2]
const parity = ['none','even','odd','mark','space']

const json = {
    //string
    nmw_com: 'COM3',
    //int
    nmw_baud_rate:9600,
    //int
    nmw_data_bits:8,
    //one:0, one_point_five:1, two:2
    nmw_stop_bits:0,
    //none:0, odd:1, even:2, mark:3, space:4
    nmw_parity: 0
}

connection.onopen = function () {
    connection.send(JSON.stringify(json));
    console.log('WebSocket connection established');
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