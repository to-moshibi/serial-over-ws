//websocketの入力をconsoleに出力する
var connection = new WebSocket('ws://localhost:8080/COM3?baud_rate=115200&data_bits=8&stop_bits=1&parity=none');

//許可される値
const data_bits =[5,6,7,8]
const  stop_bits = [1,1.5,2]
const parity = ['none','even','odd','mark','space']

connection.onopen = function () {
    // connection.send('Hello, Server!');
}
connection.onerror = function (error) {
    console.log('WebSocket Error ' + error);
}
connection.onmessage = function (e) {
    process.stdout.write(e.data);
}
connection.onclose = function () {
    console.log('WebSocket connection closed');
}