const axios = require('axios');

const url = 'http://localhost:8080/serial/settings/';
const data = {
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
};

axios.put(url, data)
    .then(response => {
        console.log('Response:', response.data);
    })
    .catch(error => {
        console.error('Error:', error);
    });