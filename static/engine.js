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
                    ctx.fillText(road, 1075, ((i+j)*115) + 82 + (k*20));
                }
            }

        }
    }
}

function drawState() {
    ctx.textAlign = "left";
    for(let i = 0; i < match_data[current_state].length; i++) {
        for(let j = 0; j < match_data[current_state][i].length; j++) {
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                if (match_data[current_state][i][j][k].Health > 0) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(1050, ((i+j)*115) + 71 + (k*20), 5, 5);
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(1045, ((i+j)*115) + 76 + (k*20), 5, 7); // make it 270,5,7 OR 269,6,4
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(1055, ((i+j)*115) + 76 + (k*20), 5, 7);
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(1050, ((i+j)*115) + 81 + (k*20), 5, 7);

                if ((i == my_team) && (j == my_int)) {
                    ctx.fillStyle = 'yellow';
                    ctx.fillRect(1052, ((i+j)*115) + 78 + (k*20), 1, 1);
                }

                if(status.substring(0,4) == "TURN") {
                    if (
                        match_data[current_state][i][j][k].Health > 0 &&
                        match_data[current_state][i][j][k].Position == 0
                    ){
                        ctx.fillStyle = 'white'
                        ctx.textAlign = "center";
                        if((i == indexer_t) && (j == indexer_u) && (k == indexer_b)) {
                            ctx.fillText("->", 1035,  ((i+j)*115) + 82 + (k*20));

                        }
                        else if((match_data[current_state][indexer_t][indexer_u][indexer_b].Position != 0) && (my_team==i) && (my_int==j)) {
                            indexer_t = i;
                            indexer_u = j;
                            indexer_b = k;
                            ctx.fillText("->", 1035,  ((i+j)*115) + 82 + (k*20));
                        }
                        else {
                            ctx.fillText("*", 1035,  ((i+j)*115) + 82 + (k*20));
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


    ctx.textAlign = "center";
    ctx.fillText("UNIT", 58, 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(10, 45);
    ctx.lineTo(105, 45);
    ctx.stroke();

    ctx.textAlign = "left";
    ctx.fillText("D", 960, 60);

    ctx.textAlign = "center";
    ctx.fillText("POS", 985, 60);
    ctx.fillText("ACT", 1010, 60);

    ctx.fillText("HEAD", 234, 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(113, 45);
    ctx.lineTo(355, 45);
    ctx.stroke();
    ctx.fillText("HP", 120, 60) // H HP
    ctx.fillText("ATK", 145, 60) // H ATK
    ctx.fillText("DEF", 170, 60) // H DEF
    ctx.fillText("ACC", 195, 60) // H ACC
    ctx.fillText("CRT", 220, 60) // H CRT
    ctx.fillText("MOB", 245, 60) // H MOB
    ctx.fillText("CD", 270, 60) // H CD
    ctx.fillText("CLU", 295, 60) // H CLU
    ctx.fillText("USE", 320, 60) // H USE
    ctx.fillText("WGH", 345, 60) // H WEIGHT

    ctx.fillText("ARMS", 513, 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(392, 45);
    ctx.lineTo(635, 45);
    ctx.stroke();
    ctx.fillText("HP", 400, 60) // ARM HP
    ctx.fillText("ATK", 425, 60) // ARM ATK
    ctx.fillText("DEF", 450, 60) // ARM DEF
    ctx.fillText("ACC", 475, 60) // ARM ACC
    ctx.fillText("CRT", 500, 60) // ARM CRIT
    ctx.fillText("MOB", 525, 60) // ARM MOBI
    ctx.fillText("CD", 550, 60) // ARM CD
    ctx.fillText("CLU", 575, 60) // ARM CLU
    ctx.fillText("USE", 600, 60) // ARM USE
    ctx.fillText("WGT", 625, 60) // ARM WEIGHT

    ctx.fillText("LEGS", 836, 40);
    ctx.lineWidth = 0.1;
    ctx.strokeStyle = "white";
    ctx.moveTo(652, 45);
    ctx.lineTo(1020, 45);
    ctx.stroke();
    ctx.fillText("HP", 660, 60) // LEG HP
    ctx.fillText("ATK", 685, 60) // LEG ATK
    ctx.fillText("DEF", 710, 60) // LEG DEF
    ctx.fillText("ACC", 735, 60) // LEG ACC
    ctx.fillText("CRT", 760, 60) // LEG CRIT
    ctx.fillText("MOB", 785, 60) // LEG MOBI
    ctx.fillText("CD", 810, 60) // LEG CD
    ctx.fillText("CLU", 835, 60) // LEG  CLU
    ctx.fillText("USE", 860, 60) // LEG USE
    ctx.fillText("WGT", 885, 60) // LEG WEIGHT
    ctx.fillText("DOG", 910, 60) // LEG DOG
    ctx.fillText("SPD", 935, 60) // LEG SPE
    ctx.fillText("ACL", 960, 60) // LEG CRIT
    ctx.fillText("ANT", 985, 60) // LEG MOBI
    ctx.fillText("END", 1010, 60) // LEG CD

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

                ctx.fillText(match_data[current_state][i][j][k].L.HP, 120, ((i+j)*115) + 82 + (k*20)) // H HP
                ctx.fillText(match_data[current_state][i][j][k].L.ATK, 145, ((i+j)*115) + 82 + (k*20)) // H ATK
                ctx.fillText(match_data[current_state][i][j][k].L.DEF, 170, ((i+j)*115) + 82 + (k*20)) // H DEF
                ctx.fillText(match_data[current_state][i][j][k].L.ACC, 195, ((i+j)*115) + 82 + (k*20)) // H ACC
                ctx.fillText(match_data[current_state][i][j][k].L.CRT, 220, ((i+j)*115) + 82 + (k*20)) // H CRIT
                ctx.fillText(match_data[current_state][i][j][k].L.MOB, 245, ((i+j)*115) + 82 + (k*20)) // H MOBI
                ctx.fillText(match_data[current_state][i][j][k].L.CD, 270, ((i+j)*115) + 82 + (k*20)) // H CD
                ctx.fillText(match_data[current_state][i][j][k].L.CLU, 295, ((i+j)*115) + 82 + (k*20)) // H CLU
                ctx.fillText(match_data[current_state][i][j][k].H.Use_current+"/"+match_data[current_state][i][j][k].H.Use_outof, 320, ((i+j)*115) + 82 + (k*20)) // H USE
                ctx.fillText(match_data[current_state][i][j][k].H.Weight, 345, ((i+j)*115) + 82 + (k*20)) // H WEIGHT

                ctx.fillText("L", 375, ((i+j)*115) + 77 + (k*20)) // L L
                ctx.fillText(match_data[current_state][i][j][k].L.HP, 400, ((i+j)*115) + 77 + (k*20)) // L HP
                ctx.fillText(match_data[current_state][i][j][k].L.ATK, 425, ((i+j)*115) + 77 + (k*20)) // L ATK
                ctx.fillText(match_data[current_state][i][j][k].L.DEF, 450, ((i+j)*115) + 77 + (k*20)) // L DEF
                ctx.fillText(match_data[current_state][i][j][k].L.ACC, 475, ((i+j)*115) + 77 + (k*20)) // L ACC
                ctx.fillText(match_data[current_state][i][j][k].L.CRT, 500, ((i+j)*115) + 77 + (k*20)) // L CRIT
                ctx.fillText(match_data[current_state][i][j][k].L.MOB, 525, ((i+j)*115) + 77 + (k*20)) // L MOBI
                ctx.fillText(match_data[current_state][i][j][k].L.CD, 550, ((i+j)*115) + 77 + (k*20)) // L CD
                ctx.fillText(match_data[current_state][i][j][k].L.CLU, 575, ((i+j)*115) + 77 + (k*20)) // L CLU
                ctx.fillText(match_data[current_state][i][j][k].L.Use_current+"/"+match_data[current_state][i][j][k].L.Use_outof, 600, ((i+j)*115) + 77 + (k*20)) // L USE
                ctx.fillText(match_data[current_state][i][j][k].L.Weight, 625, ((i+j)*115) + 77 + (k*20)) // L WEIGHT

                ctx.fillText("R", 375, ((i+j)*115) + 87 + (k*20))   // R L
                ctx.fillText(match_data[current_state][i][j][k].R.HP, 400, ((i+j)*115) + 87 + (k*20)) // R HP
                ctx.fillText(match_data[current_state][i][j][k].R.ATK, 425, ((i+j)*115) + 87 + (k*20)) // R ATK
                ctx.fillText(match_data[current_state][i][j][k].R.DEF, 450, ((i+j)*115) + 87 + (k*20)) // R DEF
                ctx.fillText(match_data[current_state][i][j][k].R.ACC, 475, ((i+j)*115) + 87 + (k*20)) // R ACC
                ctx.fillText(match_data[current_state][i][j][k].R.CRT, 500, ((i+j)*115) + 87 + (k*20)) // R CRIT
                ctx.fillText(match_data[current_state][i][j][k].R.MOB, 525, ((i+j)*115) + 87 + (k*20)) // R MOBI
                ctx.fillText(match_data[current_state][i][j][k].R.CD, 550, ((i+j)*115) + 87 + (k*20)) // R CD
                ctx.fillText(match_data[current_state][i][j][k].R.CLU, 575, ((i+j)*115) + 87 + (k*20)) // R CLU
                ctx.fillText(match_data[current_state][i][j][k].R.Use_current+"/"+match_data[current_state][i][j][k].R.Use_outof, 600, ((i+j)*115) + 87 + (k*20)) // R USE
                ctx.fillText(match_data[current_state][i][j][k].R.Weight, 625, ((i+j)*115) + 87 + (k*20)) // R WEIGHT


                ctx.fillText(match_data[current_state][i][j][k].B.HP, 660, ((i+j)*115) + 82 + (k*20)) // LEG HP
                ctx.fillText(match_data[current_state][i][j][k].B.ATK, 685, ((i+j)*115) + 82 + (k*20)) // LEG ATK
                ctx.fillText(match_data[current_state][i][j][k].B.DEF, 710, ((i+j)*115) + 82 + (k*20)) // LEG DEF
                ctx.fillText(match_data[current_state][i][j][k].B.ACC, 735, ((i+j)*115) + 82 + (k*20)) // LEG ACC
                ctx.fillText(match_data[current_state][i][j][k].B.CRT, 760, ((i+j)*115) + 82 + (k*20)) // LEG CRIT
                ctx.fillText(match_data[current_state][i][j][k].B.MOB, 785, ((i+j)*115) + 82 + (k*20)) // LEG MOBI
                ctx.fillText(match_data[current_state][i][j][k].B.CD, 810, ((i+j)*115) + 82 + (k*20)) // LEG CD
                ctx.fillText(match_data[current_state][i][j][k].B.CLU, 835, ((i+j)*115) + 82 + (k*20)) // LEG  CLU
                ctx.fillText(match_data[current_state][i][j][k].B.Use_current+"/"+match_data[current_state][i][j][k].B.Use_outof, 860, ((i+j)*115) + 82 + (k*20)) // LEG USE
                ctx.fillText(match_data[current_state][i][j][k].B.Weight, 885, ((i+j)*115) + 82 + (k*20)) // LEG WEIGHT
                ctx.fillText(match_data[current_state][i][j][k].B.DOG, 910, ((i+j)*115) + 82 + (k*20)) // LEG DOG
                ctx.fillText(match_data[current_state][i][j][k].B.SPD, 935, ((i+j)*115) + 82 + (k*20)) // LEG SPE
                ctx.fillText(match_data[current_state][i][j][k].B.ACL, 960, ((i+j)*115) + 82 + (k*20)) // LEG ACL
                ctx.fillText(match_data[current_state][i][j][k].B.ANT, 985, ((i+j)*115) + 82 + (k*20)) // LEG ANT
                ctx.fillText(match_data[current_state][i][j][k].B.END, 1010, ((i+j)*115) + 82 + (k*20)) // LEG END
                
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