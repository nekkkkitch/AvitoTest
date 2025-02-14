import http from 'k6/http';
import { sleep } from 'k6';
export const options = {
    scenarios:{
        contacts:{
            executor:'ramping-vus',
            startVUs:100,
            stages:[
                {duration:'10s', target:100},
            ]
        }
    }
};
var available_users = ['1','2','3','4','5','6','7','8','9','10']
export default function () {
    const part = ['a', 'b', 'c', 'd', 'e', '1', '2', '3', '4'];
    let l = part.length
    let username = ''
    for(let i = 0; i < 20; i++){
        let r = Math.floor(Math.random()*l);
        username += part[r];
    }
    const url = 'http://localhost:8080/api';
    let payload = JSON.stringify({
        username: username,
        password:'123',
    });
    var params = {
        headers: {
          'Content-Type': 'application/json',
          'Authorization':'',
        },
    };
    let res = http.post(url+'/auth', payload, params);
    if(res.status != 200){
        console.log(res);
    }
    let rich_man = available_users[Math.floor(Math.random()*available_users.length)]
    sleep(0.01);
    params.headers.Authorization = 'Bearer ' + res.json().token;
    res = http.get(url+'/info', params);
    if(res.status != 200){
        console.log(res);
    }
    sleep(0.01);
    res = http.get(url+'/buy/cup', params);
    if(res.status != 200){
        console.log(res);
    }
    sleep(0.01);
    payload = JSON.stringify({
        toUser: rich_man,
        amount: 10,
    })
    res = http.post(url+'/sendCoin', payload, params)
    if(res.status != 200){
        console.log(res);
    }
    sleep(0.01);
  }