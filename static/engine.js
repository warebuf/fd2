console.log("loaded js");

var c = document.getElementById("myCanvas");
const ctx = c.getContext("2d");

//c.width = window.innerWidth;
//c.height = window.innerHeight;

c.width = document.body.clientWidth;
c.height = document.body.clientHeight;

// shim layer with setTimeout fallback
window.requestAnimFrame = (function(){
    return  window.requestAnimationFrame       ||
        window.webkitRequestAnimationFrame ||
        window.mozRequestAnimationFrame    ||
        window.oRequestAnimationFrame      ||
        window.msRequestAnimationFrame     ||
        function( callback ){
            window.setTimeout(callback, 1000 / 60);
        };
})();

window.cancelAnimationFrame = window.cancelAnimationFrame || window.mozCancelAnimationFrame || function(requestID){clearTimeout(requestID)};

var once = true

function anime() {
    window.requestAnimFrame(anime);
    ctx.fillStyle = '#101010';
    //ctx.clearRect(0,0,innerWidth,innerHeight);
    ctx.fillRect(0,0,innerWidth,innerHeight);

    let start = Date.now().toLocaleString('en-CH');

    if(startCount != null) {
        ctx.textAlign = "center";
        remaining_time = startCount - start.replaceAll("’","")
        ctx.fillStyle = 'black';
        ctx.fillRect((c.width/2)-(ctx.measureText(remaining_time).width/2),25,ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(remaining_time,(c.width/2) - (ctx.measureText(remaining_time).width/2), 35);

        // need to change it so that it not only sends ur input when time expires, but sends it if you disconnect
        if(remaining_time<=0 && once) {
            socket.send(JSON.stringify({ "Event": 'timeUpMsg', "Message": "test" }))
            once = false
        }
    }

    if (match_data.length > 0) {
        drawPos()
        drawState()
        drawStats()
        drawTurn()
        drawLog()
    }

}
anime()

function drawPos() {
    ctx.font = '11px monospace';
    ctx.fillStyle = 'white';
    ctx.textAlign = "left";
    for(let i = 0; i < match_data[current_state].length; i++) { // team
        for(let j = 0; j < match_data[current_state][i].length; j++) { // client
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                if(match_data[current_state][i][j][k].Health > 0) {
                    let road = 'o' + "-".repeat(99) + 'x';
                    road = road.substring(0,
                        Math.round(match_data[current_state][i][j][k].Position)) +
                        (match_data[current_state][i][j][k].Direction ? '>' : '<') +
                        road.substring(Math.round(match_data[current_state][i][j][k].Position) + 1);
                    ctx.fillText(road, 1125, ((i+j)*115) + 82 + (k*20));
                }
            }

        }
    }
}

function drawState() {
    x0 = 1100
    ctx.textAlign = "left";
    for(let i = 0; i < match_data[current_state].length; i++) {
        for(let j = 0; j < match_data[current_state][i].length; j++) {
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                if (match_data[current_state][i][j][k].Health > 0) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0, ((i+j)*115) + 71 + (k*20), 5, 5);
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0-5, ((i+j)*115) + 76 + (k*20), 5, 7); // make it 270,5,7 OR 269,6,4
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0+5, ((i+j)*115) + 76 + (k*20), 5, 7);
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0, ((i+j)*115) + 81 + (k*20), 5, 7);

                if ((i == my_team) && (j == my_int)) {
                    ctx.fillStyle = 'yellow';
                    ctx.fillRect(x0+2, ((i+j)*115) + 78 + (k*20), 1, 1);
                }

                if(status.substring(0,4) == "TURN") {
                    if (
                        match_data[current_state][i][j][k].Health > 0 &&
                        match_data[current_state][i][j][k].Position == 0
                    ){
                        ctx.fillStyle = 'white'
                        ctx.textAlign = "center";
                        if((i == indexer_t) && (j == indexer_u) && (k == indexer_b)) {
                            ctx.fillText("->", x0-15,  ((i+j)*115) + 82 + (k*20));

                        }
                        else if((match_data[current_state][indexer_t][indexer_u][indexer_b].Position != 0) && (my_team==i) && (my_int==j)) {
                            indexer_t = i;
                            indexer_u = j;
                            indexer_b = k;
                            ctx.fillText("->", x0-15,  ((i+j)*115) + 82 + (k*20));
                        }
                        else {
                            ctx.fillText("*", x0-15,  ((i+j)*115) + 82 + (k*20));
                        }



                    }
                }
            }
        }
    }

}

function drawTurn() {
    ctx.textAlign = "center";
    ctx.fillStyle = 'black';
    ctx.fillRect((c.width/2)-(ctx.measureText(remaining_time).width/2),10,ctx.measureText(status).width,ctx.measureText('M').width);

    ctx.fillStyle = 'white'
    ctx.textAlign = "center";
    ctx.fillText(status,(c.width/2) - (ctx.measureText(remaining_time).width/2), 20);
}

function drawStats() {
    ctx.fillStyle = 'white'

    a0 = 975
    ctx.textAlign = "center";
    ctx.fillText("UNIT", a0+(105/2), 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a0-5, 45);
    ctx.lineTo(a0+105, 45);
    ctx.stroke();
    ctx.textAlign = "center";
    ctx.fillText("D", a0 +(25*0), 60);
    ctx.fillText("POS", a0 +(25*1), 60);
    ctx.fillText("ACT", a0 +(25*2), 60);

    a1 = 10
    ctx.fillText("HEAD", a1+(260/2), 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a1-5, 45);
    ctx.lineTo(a1+260, 45);
    ctx.stroke();
    ctx.fillText("S#", a1 +(25*0), 60) // H HP
    ctx.fillText("HP", a1 +(25*1), 60) // H HP
    ctx.fillText("ATK", a1 +(25*2), 60) // H ATK
    ctx.fillText("DEF", a1 +(25*3), 60) // H DEF
    ctx.fillText("ACC", a1 +(25*4), 60) // H ACC
    ctx.fillText("CRT", a1 +(25*5), 60) // H CRT
    ctx.fillText("MOB", a1 +(25*6), 60) // H MOB
    ctx.fillText("CD", a1 +(25*7), 60) // H CD
    ctx.fillText("CLU", a1 +(25*8), 60) // H CLU
    ctx.fillText("USE", a1 +(25*9), 60) // H USE
    ctx.fillText("WGH", a1 +(25*10), 60) // H WEIGHT

    a2 = 300
    ctx.fillText("ARMS", a2+(250/2), 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a2-5, 45);
    ctx.lineTo(a2+250, 45);
    ctx.stroke();
    ctx.fillText("S#", a2 +(25*0), 60) // H HP
    ctx.fillText("HP", a2 +(25*1), 60) // ARM HP
    ctx.fillText("ATK", a2 +(25*2), 60) // ARM ATK
    ctx.fillText("DEF", a2 +(25*3), 60) // ARM DEF
    ctx.fillText("ACC", a2 +(25*4), 60) // ARM ACC
    ctx.fillText("CRT", a2 +(25*5), 60) // ARM CRIT
    ctx.fillText("MOB", a2 +(25*6), 60) // ARM MOBI
    ctx.fillText("CD", a2 +(25*7), 60) // ARM CD
    ctx.fillText("CLU", a2 +(25*8), 60) // ARM CLU
    ctx.fillText("USE", a2 +(25*9), 60) // ARM USE
    ctx.fillText("WGT", a2 +(25*10), 60) // ARM WEIGHT

    a3 = 575
    ctx.fillText("LEGS", a3+(375/2), 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a3-5, 45);
    ctx.lineTo(a3+375, 45);
    ctx.stroke();
    ctx.fillText("S#", a3 +(25*0), 60) // LEG HP
    ctx.fillText("HP", a3 +(25*1), 60) // LEG HP
    ctx.fillText("ATK", a3 +(25*2), 60) // LEG ATK
    ctx.fillText("DEF", a3 +(25*3), 60) // LEG DEF
    ctx.fillText("ACC", a3 +(25*4), 60) // LEG ACC
    ctx.fillText("CRT", a3 +(25*5), 60) // LEG CRIT
    ctx.fillText("MOB", a3 +(25*6), 60) // LEG MOBI
    ctx.fillText("CD", a3 +(25*7), 60) // LEG CD
    ctx.fillText("CLU", a3 +(25*8), 60) // LEG  CLU
    ctx.fillText("USE", a3 +(25*9), 60) // LEG USE
    ctx.fillText("WGT", a3 +(25*10), 60) // LEG WEIGHT
    ctx.fillText("DOG", a3 +(25*11), 60) // LEG DOG
    ctx.fillText("SPD", a3 +(25*12), 60) // LEG SPE
    ctx.fillText("ACL", a3 +(25*13), 60) // LEG CRIT
    ctx.fillText("ANT", a3 +(25*14), 60) // LEG MOBI
    ctx.fillText("END", a3 +(25*15), 60) // LEG CD

    for(let i = 0; i < match_data[current_state].length; i++) {
        for(let j = 0; j < match_data[current_state][i].length; j++) {
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                ctx.textAlign = "left";
                ctx.fillText(match_data[current_state][i][j][k].Direction, 960, ((i+j)*115) + 82 + (k*20));
                ctx.textAlign = "center";
                ctx.fillText((Math.round(match_data[current_state][i][j][k].Position * 100) / 100).toFixed(1), 985, ((i+j)*115) + 82 + (k*20))
                temp = '-'
                if(match_data[current_state][i][j][k].Move==0){temp='H'}
                else if(match_data[current_state][i][j][k].Move==1){temp='L'}
                else if(match_data[current_state][i][j][k].Move==2){temp='R'}
                else if(match_data[current_state][i][j][k].Move==3){temp='B'}
                ctx.fillText(temp, 1010, ((i+j)*115) + 82 + (k*20))

                ctx.fillText(match_data[current_state][i][j][k].H.SERIAL, a1 +(25*0), ((i+j)*115) + 82 + (k*20)) // H HP
                ctx.fillText(match_data[current_state][i][j][k].H.HP, a1 +(25*1), ((i+j)*115) + 82 + (k*20)) // H HP
                ctx.fillText(match_data[current_state][i][j][k].H.ATK, a1 +(25*2), ((i+j)*115) + 82 + (k*20)) // H ATK
                ctx.fillText(match_data[current_state][i][j][k].H.DEF, a1 +(25*3), ((i+j)*115) + 82 + (k*20)) // H DEF
                ctx.fillText(match_data[current_state][i][j][k].H.ACC, a1 +(25*4), ((i+j)*115) + 82 + (k*20)) // H ACC
                ctx.fillText(match_data[current_state][i][j][k].H.CRT, a1 +(25*5), ((i+j)*115) + 82 + (k*20)) // H CRIT
                ctx.fillText(match_data[current_state][i][j][k].H.MOB, a1 +(25*6), ((i+j)*115) + 82 + (k*20)) // H MOBI
                ctx.fillText(match_data[current_state][i][j][k].H.CD, a1 +(25*7), ((i+j)*115) + 82 + (k*20)) // H CD
                ctx.fillText(match_data[current_state][i][j][k].H.CLU, a1 +(25*8), ((i+j)*115) + 82 + (k*20)) // H CLU
                ctx.fillText(match_data[current_state][i][j][k].H.Use_current+"/"+match_data[current_state][i][j][k].H.Use_outof, a1 +(25*9), ((i+j)*115) + 82 + (k*20)) // H USE
                ctx.fillText(match_data[current_state][i][j][k].H.Weight, a1 +(25*10), ((i+j)*115) + 82 + (k*20)) // H WEIGHT

                ctx.fillText("L", a2 -(20*1), ((i+j)*115) + 77 + (k*20)) // L L
                ctx.fillText(match_data[current_state][i][j][k].L.SERIAL, a2 +(25*0), ((i+j)*115) + 77 + (k*20)) // H HP
                ctx.fillText(match_data[current_state][i][j][k].L.HP, a2 +(25*1), ((i+j)*115) + 77 + (k*20)) // L HP
                ctx.fillText(match_data[current_state][i][j][k].L.ATK, a2 +(25*2), ((i+j)*115) + 77 + (k*20)) // L ATK
                ctx.fillText(match_data[current_state][i][j][k].L.DEF, a2 +(25*3), ((i+j)*115) + 77 + (k*20)) // L DEF
                ctx.fillText(match_data[current_state][i][j][k].L.ACC, a2 +(25*4), ((i+j)*115) + 77 + (k*20)) // L ACC
                ctx.fillText(match_data[current_state][i][j][k].L.CRT, a2 +(25*5), ((i+j)*115) + 77 + (k*20)) // L CRIT
                ctx.fillText(match_data[current_state][i][j][k].L.MOB, a2 +(25*6), ((i+j)*115) + 77 + (k*20)) // L MOBI
                ctx.fillText(match_data[current_state][i][j][k].L.CD, a2 +(25*7), ((i+j)*115) + 77 + (k*20)) // L CD
                ctx.fillText(match_data[current_state][i][j][k].L.CLU, a2 +(25*8), ((i+j)*115) + 77 + (k*20)) // L CLU
                ctx.fillText(match_data[current_state][i][j][k].L.Use_current+"/"+match_data[current_state][i][j][k].L.Use_outof, a2 +(25*9), ((i+j)*115) + 77 + (k*20)) // L USE
                ctx.fillText(match_data[current_state][i][j][k].L.Weight, a2 +(25*10), ((i+j)*115) + 77 + (k*20)) // L WEIGHT

                ctx.fillText("R", a2 -(20*1), ((i+j)*115) + 87 + (k*20))   // R L
                ctx.fillText(match_data[current_state][i][j][k].R.SERIAL, a2 +(25*0), ((i+j)*115) + 87 + (k*20)) // H HP
                ctx.fillText(match_data[current_state][i][j][k].R.HP, a2 +(25*1), ((i+j)*115) + 87 + (k*20)) // R HP
                ctx.fillText(match_data[current_state][i][j][k].R.ATK, a2 +(25*2), ((i+j)*115) + 87 + (k*20)) // R ATK
                ctx.fillText(match_data[current_state][i][j][k].R.DEF, a2 +(25*3), ((i+j)*115) + 87 + (k*20)) // R DEF
                ctx.fillText(match_data[current_state][i][j][k].R.ACC, a2 +(25*4), ((i+j)*115) + 87 + (k*20)) // R ACC
                ctx.fillText(match_data[current_state][i][j][k].R.CRT, a2 +(25*5), ((i+j)*115) + 87 + (k*20)) // R CRIT
                ctx.fillText(match_data[current_state][i][j][k].R.MOB, a2 +(25*6), ((i+j)*115) + 87 + (k*20)) // R MOBI
                ctx.fillText(match_data[current_state][i][j][k].R.CD, a2 +(25*7), ((i+j)*115) + 87 + (k*20)) // R CD
                ctx.fillText(match_data[current_state][i][j][k].R.CLU, a2 +(25*8), ((i+j)*115) + 87 + (k*20)) // R CLU
                ctx.fillText(match_data[current_state][i][j][k].R.Use_current+"/"+match_data[current_state][i][j][k].R.Use_outof, a2 +(25*9), ((i+j)*115) + 87 + (k*20)) // R USE
                ctx.fillText(match_data[current_state][i][j][k].R.Weight, a2 +(25*10), ((i+j)*115) + 87 + (k*20)) // R WEIGHT


                ctx.fillText(match_data[current_state][i][j][k].B.SERIAL, a3 +(25*0), ((i+j)*115) + 82 + (k*20)) // LEG HP
                ctx.fillText(match_data[current_state][i][j][k].B.HP, a3 +(25*1), ((i+j)*115) + 82 + (k*20)) // LEG HP
                ctx.fillText(match_data[current_state][i][j][k].B.ATK, a3 +(25*2), ((i+j)*115) + 82 + (k*20)) // LEG ATK
                ctx.fillText(match_data[current_state][i][j][k].B.DEF, a3 +(25*3), ((i+j)*115) + 82 + (k*20)) // LEG DEF
                ctx.fillText(match_data[current_state][i][j][k].B.ACC, a3 +(25*4), ((i+j)*115) + 82 + (k*20)) // LEG ACC
                ctx.fillText(match_data[current_state][i][j][k].B.CRT, a3 +(25*5), ((i+j)*115) + 82 + (k*20)) // LEG CRIT
                ctx.fillText(match_data[current_state][i][j][k].B.MOB, a3 +(25*6), ((i+j)*115) + 82 + (k*20)) // LEG MOBI
                ctx.fillText(match_data[current_state][i][j][k].B.CD, a3 +(25*7), ((i+j)*115) + 82 + (k*20)) // LEG CD
                ctx.fillText(match_data[current_state][i][j][k].B.CLU, a3 +(25*8), ((i+j)*115) + 82 + (k*20)) // LEG  CLU
                ctx.fillText(match_data[current_state][i][j][k].B.Use_current+"/"+match_data[current_state][i][j][k].B.Use_outof, a3 +(25*9), ((i+j)*115) + 82 + (k*20)) // LEG USE
                ctx.fillText(match_data[current_state][i][j][k].B.Weight, a3 +(25*10), ((i+j)*115) + 82 + (k*20)) // LEG WEIGHT
                ctx.fillText(match_data[current_state][i][j][k].B.DOG, a3 +(25*11), ((i+j)*115) + 82 + (k*20)) // LEG DOG
                ctx.fillText(match_data[current_state][i][j][k].B.SPD, a3 +(25*12), ((i+j)*115) + 82 + (k*20)) // LEG SPE
                ctx.fillText(match_data[current_state][i][j][k].B.ACL, a3 +(25*13), ((i+j)*115) + 82 + (k*20)) // LEG ACL
                ctx.fillText(match_data[current_state][i][j][k].B.ANT, a3 +(25*14), ((i+j)*115) + 82 + (k*20)) // LEG ANT
                ctx.fillText(match_data[current_state][i][j][k].B.END, a3 +(25*15), ((i+j)*115) + 82 + (k*20)) // LEG END
                
            }
        }
    }
}

function drawLog() {
    for(i =0; i<event_log.length;i++) {
        ctx.textAlign = "left";
        ctx.fillStyle = 'black';
        ctx.fillRect(1250,10 + (10*i),ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(event_log[i],1250, 20 + (10*i));
    }
}