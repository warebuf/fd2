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

function anime() {
    window.requestAnimFrame(anime);
    ctx.fillStyle = '#101010';
    //ctx.clearRect(0,0,innerWidth,innerHeight);
    ctx.fillRect(0,0,innerWidth,innerHeight);

    let start = Date.now().toLocaleString('en-CH');

    if(game_over == true) {
        ctx.textAlign = "center";
        ctx.fillStyle = 'blue';
        ctx.fillRect((c.width/2)-(ctx.measureText("GAMEOVER").width/2),42,ctx.measureText("GAMEOVER").width,ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText("GAMEOVER",c.width/2, 50);
    }
    else if(sent_time != null) {
        ctx.textAlign = "center";
        remaining_time = sent_time - start.replaceAll("’","")
        ctx.fillStyle = 'black';
        ctx.fillRect(c.width/2,40,ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(remaining_time,c.width/2, 50);
    }

    if(phase=="CHARACTER SELECTION") {
        drawPhase()
        drawLog()
        drawBench()
        drawStarters()
    } else if (match_data.length > 0) {
        drawPos()
        drawState()
        drawStats()
        drawTurn()
        drawUnitsOfTime()
        drawLog()
    }

}
anime()

function drawPos() {

    time_event_ready = false

    // if there are states updates, update the position
    if(all_up_to_date == false && draw_attacks <= 0) {

        // check if the states are up to date
        if(pos_up_to_date == false) {
            check = true
            for(let i = 0; i < state.length; i++) {
                for(let j = 0; j < state[i].length; j++) {
                    for(let k = 0; k < state[i][j].length; k++) {
                        if(
                            state[i][j][k].Position != match_data[animating_state][i][j][k].Position ||
                            state[i][j][k].Direction != match_data[animating_state][i][j][k].Direction
                        ) {
                            check = false
                            break
                        }
                    }
                    if(check==false){break}
                }
                if(check==false){break}
            }
            if (check == true) {pos_up_to_date=true}
        }

        if(act_up_to_date == false) {
            check = true
            for(let i = 0; i < state.length; i++) {
                for(let j = 0; j < state[i].length; j++) {
                    for(let k = 0; k < state[i][j].length; k++) {
                        if(
                            state[i][j][k].Move != match_data[animating_state][i][j][k].Move &&
                            state[i][j][k].Direction != match_data[animating_state][i][j][k].Direction
                        ) {
                            check = false
                            break
                        }
                    }
                    if(check==false){break}
                }
                if(check==false){break}
            }
            if (check == true) {
                act_up_to_date=true
            }
        }

        // if we are at this state, let's draw the next one
        // if we are not at this state, let's move the units so we can approach the next state
        if((pos_up_to_date == true) && (act_up_to_date == true)) {

            attack_event_ready = false
            for(let i = 0; i < state.length; i++) {
                for(let j = 0; j < state[i].length; j++) {
                    for(let k = 0; k < state[i][j].length; k++) {
                        if(
                            (match_data[animating_state][i][j][k].Position == 0) &&
                            (match_data[animating_state][i][j][k].Direction == 0) &&
                            (match_data[animating_state][i][j][k].Move == -1) &&
                            (match_data[animating_state][i][j][k].SPD != 0)
                        ) {
                            time_event_ready = true
                        } else if (
                            (match_data[animating_state][i][j][k].Position == 100) &&
                            (match_data[animating_state][i][j][k].Direction == 1) &&
                            (match_data[animating_state][i][j][k].Move != -1)
                        ) {
                            time_event_ready = false
                            attack_event_ready = true
                            break
                        }
                    }
                    if((attack_event_ready)){break}
                }
                if((attack_event_ready)){break}
            }

            state = JSON.parse(JSON.stringify(match_data[animating_state]))
            pos_up_to_date = false
            act_up_to_date = false

            turn = turn_queue[animating_state]

            animating_state++
            if(animating_state >= match_data.length) {
                all_up_to_date = true
            }
        } else {

            if(act_up_to_date == false) {
                for(let i = 0; i < state.length; i++) {
                    for(let j = 0; j < state[i].length; j++) {
                        for(let k = 0; k < state[i][j].length; k++) {

                            if( // if we already set a move on this unit, but the user requests a change on it, and the unit is at the starting point
                                (state[i][j][k].Position == 0) &&
                                (state[i][j][k].Direction == 1) &&
                                (state[i][j][k].Position == match_data[animating_state][i][j][k].Position) 
                            ) {
                                state[i][j][k].Move = match_data[animating_state][i][j][k].Move
                            }
                            else if( // if the unit is waiting for a move, and a new move comes in
                                (state[i][j][k].Position == 0) &&
                                (state[i][j][k].Direction == 0) &&
                                (state[i][j][k].Move == -1) &&
                                (state[i][j][k].Move != match_data[animating_state][i][j][k].Move)
                            ) {
                                state[i][j][k].Move = match_data[animating_state][i][j][k].Move
                                state[i][j][k].Direction = 1
                            } else if( // if the unit is ready to attack, attack and set direction to homebase
                                (state[i][j][k].Position == 100) &&
                                (state[i][j][k].Direction == 1) &&
                                (state[i][j][k].Move != -1)
                            ){
                                if(draw_attacks==0){
                                    draw_attacks=1
                                }
                            }
                        }
                    }
                }
            }

            // move all units
            if(pos_up_to_date == false) {
                for(let i = 0; i < state.length; i++) {
                    for(let j = 0; j < state[i].length; j++) {
                        for(let k = 0; k < state[i][j].length; k++) {
                            if(state[i][j][k].H.HP <= 0) {
                            } else if(state[i][j][k].Direction == 0) {
                                state[i][j][k].Position = state[i][j][k].Position - (state[i][j][k].B.SPD * 0.003)
                                if(
                                    (match_data[animating_state][i][j][k].Direction == 0) &&
                                    (state[i][j][k].Position < match_data[animating_state][i][j][k].Position)
                                ) {
                                    state[i][j][k].Position = match_data[animating_state][i][j][k].Position
                                }
                                else if(state[i][j][k].Position < 0){state[i][j][k].Position = 0}
                            } else if(state[i][j][k].Direction == 1) {
                                state[i][j][k].Position = state[i][j][k].Position + (state[i][j][k].B.SPD * 0.003)
                                if(
                                    (match_data[animating_state][i][j][k].Direction == 1) &&
                                    (state[i][j][k].Position > match_data[animating_state][i][j][k].Position)
                                ) {
                                    state[i][j][k].Position = match_data[animating_state][i][j][k].Position
                                }
                                else if(state[i][j][k].Position > 100){
                                    state[i][j][k].Position = 100
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    else if(draw_attacks>0) {

        let a1 = 0
        let a2 = 0
        for(let i = 0; i < state.length; i++) {
            for(let j = 0; j < state[i].length; j++) {
                for(let k = 0; k < state[i][j].length; k++) {
                    if( (state[i][j][k].H.HP > 0) && (state[i][j][k].Position==100) ) {

                        if(a1 < atk_data[animating_state].length) {
                            let b = JSON.parse(atk_data[animating_state][a1])
                            let pos = state[b.Defender[0][0]][b.Defender[0][1]][b.Defender[0][2]].Position
                            let x2 = 1085 + (pos*6.08)
                            let y2 = ((b.Defender[0][0]+b.Defender[0][1])*115) + 100 + (b.Defender[0][2]*20)

                            let perc = draw_attacks / 30
                            if(draw_attacks>30){
                                perc = 1
                            }

                            // calculate a vector to draw
                            let x1 = 1690
                            let y1 = ((i+j)*115) + 98 + (k*20)
                            let x2_x1 = (x2 - x1)*perc
                            let y2_y1 = (y2 - y1)*perc
                            dashedLine(x1,y1,x1+x2_x1,y1+y2_y1,[5,2])

                            if(draw_attacks>=30) {
                                ctx.fillStyle = 'white'
                                ctx.textAlign = "center";
                                ctx.font = '9px monospace';
                                ctx.fillText("BOOM!",x2, y2-7);
                            }

                            a1++
                        }

                        if ((draw_attacks==59) && (a2 < atk_data[animating_state].length)) {
                            event_log.push(atk_data[animating_state][a2])
                            a2++
                        }

                        if ((draw_attacks==59) && (state[i][j][k].Direction==1)) {
                            state[i][j][k].Move = -1
                            state[i][j][k].Direction = 0
                            time_event_ready = true
                        }

                    }
                }
            }
        }

        // have to draw BOOM animations and -100 animations



        draw_attacks++
        if(draw_attacks==60) {draw_attacks = 0}
    }

    if(time_event_ready==true) {
        if(time_queue.length > 0) {
            t = time_queue.shift()
            total_units_of_time += parseFloat(t)
            total_units_of_time = parseFloat(total_units_of_time.toFixed(3))
            event_log.push(t.toString() + " units of time has passed")
        }
    }


    // draw the position
    ctx.textAlign = "left";
    ctx.fillStyle = 'white';
    ctx.font = '11px monospace';
    symbol = '?'
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                if(state[i][j][k].H.HP <= 0) {
                    symbol = 'd'
                } else if(state[i][j][k].Direction == 0) {
                    symbol = '<'
                } else if(state[i][j][k].Direction == 1) {
                    symbol = '>'
                }
                let road = 'o' + "-".repeat(99) + 'x';
                road = road.substring(0, Math.round(state[i][j][k].Position)) + symbol + road.substring(Math.round(state[i][j][k].Position) + 1);
                ctx.fillText(road, 1085, ((i+j)*115) + 102 + (k*20));
            }
        }
    }
}

function drawState() {
    x0 = 1060
    y0 = 100

    ctx.textAlign = "left";
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                if (state[i][j][k].H.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0, ((i+j)*115) + y0-9 + (k*20), 5, 5);
                if (state[i][j][k].L.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0-5, ((i+j)*115) + y0-4 + (k*20), 5, 7); // make it 270,5,7 OR 269,6,4
                if (state[i][j][k].R.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0+5, ((i+j)*115) + y0-4 + (k*20), 5, 7);
                if (state[i][j][k].B.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0, ((i+j)*115) + y0+1 + (k*20), 5, 7);

                if(turn.substring(0,4) == "TURN") {
                    if (
                        state[i][j][k].H.HP > 0 &&
                        state[i][j][k].Position == 0
                    ) {
                        ctx.fillStyle = 'white'
                        ctx.textAlign = "center";
                        if((i == indexer_t) && (j == indexer_u) && (k == indexer_b)) {
                            ctx.fillText("->", x0-15,  ((i+j)*115) + y0+2 + (k*20));
                        }
                        else if((state[indexer_t][indexer_u][indexer_b].Position != 0) && (my_team==i) && (my_int==j)) {
                            indexer_t = i;
                            indexer_u = j;
                            indexer_b = k;
                            ctx.fillText("->", x0-15,  ((i+j)*115) + y0+2 + (k*20));
                        }
                        else {
                            ctx.fillText("*", x0-15,  ((i+j)*115) + y0+2 + (k*20));
                        }
                    }
                }
            }
        }
    }

}

function drawPhase() {
    ctx.textAlign = "center";
    ctx.fillStyle = 'black';
    ctx.fillRect(c.width/2,10,ctx.measureText(phase).width,ctx.measureText('M').width);

    ctx.fillStyle = 'white'
    ctx.textAlign = "center";
    ctx.fillText(phase,c.width/2, 20);
}

function drawTurn() {
    ctx.textAlign = "center";
    ctx.fillStyle = 'black';
    ctx.fillRect(c.width/2,10,ctx.measureText(turn).width,ctx.measureText('M').width);

    ctx.fillStyle = 'white'
    ctx.textAlign = "center";
    ctx.fillText(turn,c.width/2, 20);
}

function drawUnitsOfTime() {
    ctx.textAlign = "center";
    ctx.font = '11px monospace';
    ctx.fillStyle = 'white';
    time_string = total_units_of_time+" units of time"
    ctx.fillText(time_string,c.width/2, 35);
}
function drawStats() {
    ctx.fillStyle = 'white'

    y0 = 60
    y1 = 100

    a0 = 975
    ctx.textAlign = "center";
    ctx.fillText("UNIT", a0+(60/2), y0);
    ctx.beginPath()
    ctx.lineWidth = 1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a0-5, y0+5);
    ctx.lineTo(a0+60, y0+5);
    ctx.stroke();
    ctx.textAlign = "center";
    ctx.fillText("D", a0 +(25*0), y0+20);
    ctx.fillText("POS", a0 +(25*1), y0+20);
    ctx.fillText("ACT", a0 +(25*2), y0+20);

    a1 = 10
    ctx.fillText("HEAD", a1+(260/2), y0);
    ctx.beginPath()
    ctx.lineWidth = 1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a1-5, y0+5);
    ctx.lineTo(a1+260, y0+5);
    ctx.stroke();
    ctx.fillText("S#", a1 +(25*0), y0+20) // H HP
    ctx.fillText("HP", a1 +(25*1), y0+20) // H HP
    ctx.fillText("ATK", a1 +(25*2), y0+20) // H ATK
    ctx.fillText("DEF", a1 +(25*3), y0+20) // H DEF
    ctx.fillText("ACC", a1 +(25*4), y0+20) // H ACC
    ctx.fillText("CRT", a1 +(25*5), y0+20) // H CRT
    ctx.fillText("MOB", a1 +(25*6), y0+20) // H MOB
    ctx.fillText("CD", a1 +(25*7), y0+20) // H CD
    ctx.fillText("CLU", a1 +(25*8), y0+20) // H CLU
    ctx.fillText("USE", a1 +(25*9), y0+20) // H USE
    ctx.fillText("WGH", a1 +(25*10), y0+20) // H WEIGHT

    a2 = 300
    ctx.fillText("ARMS", a2+(250/2), y0);
    ctx.beginPath()
    ctx.lineWidth = 1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a2-5, y0+5);
    ctx.lineTo(a2+250, y0+5);
    ctx.stroke();
    ctx.fillText("S#", a2 +(25*0), y0+20) // H HP
    ctx.fillText("HP", a2 +(25*1), y0+20) // ARM HP
    ctx.fillText("ATK", a2 +(25*2), y0+20) // ARM ATK
    ctx.fillText("DEF", a2 +(25*3), y0+20) // ARM DEF
    ctx.fillText("ACC", a2 +(25*4), y0+20) // ARM ACC
    ctx.fillText("CRT", a2 +(25*5), y0+20) // ARM CRIT
    ctx.fillText("MOB", a2 +(25*6), y0+20) // ARM MOBI
    ctx.fillText("CD", a2 +(25*7), y0+20) // ARM CD
    ctx.fillText("CLU", a2 +(25*8), y0+20) // ARM CLU
    ctx.fillText("USE", a2 +(25*9), y0+20) // ARM USE
    ctx.fillText("WGT", a2 +(25*10), y0+20) // ARM WEIGHT

    a3 = 575
    ctx.fillText("LEGS", a3+(375/2), y0);
    ctx.beginPath()
    ctx.lineWidth = 1;
    ctx.strokeStyle = "white";
    ctx.moveTo(a3-5, y0+5);
    ctx.lineTo(a3+375, y0+5);
    ctx.stroke();
    ctx.fillText("S#", a3 +(25*0), y0+20) // LEG HP
    ctx.fillText("HP", a3 +(25*1), y0+20) // LEG HP
    ctx.fillText("ATK", a3 +(25*2), y0+20) // LEG ATK
    ctx.fillText("DEF", a3 +(25*3), y0+20) // LEG DEF
    ctx.fillText("ACC", a3 +(25*4), y0+20) // LEG ACC
    ctx.fillText("CRT", a3 +(25*5), y0+20) // LEG CRIT
    ctx.fillText("MOB", a3 +(25*6), y0+20) // LEG MOBI
    ctx.fillText("CD", a3 +(25*7), y0+20) // LEG CD
    ctx.fillText("CLU", a3 +(25*8), y0+20) // LEG  CLU
    ctx.fillText("USE", a3 +(25*9), y0+20) // LEG USE
    ctx.fillText("WGT", a3 +(25*10), y0+20) // LEG WEIGHT
    ctx.fillText("DOG", a3 +(25*11), y0+20) // LEG DOG
    ctx.fillText("SPD", a3 +(25*12), y0+20) // LEG SPE
    ctx.fillText("ACL", a3 +(25*13), y0+20) // LEG CRIT
    ctx.fillText("ANT", a3 +(25*14), y0+20) // LEG MOBI
    ctx.fillText("END", a3 +(25*15), y0+20) // LEG CD

    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                ctx.textAlign = "center";
                ctx.fillText(state[i][j][k].Direction, a0+(0*25), ((i+j)*115) + y1+2 + (k*20));
                ctx.fillText((Math.round(state[i][j][k].Position * 100) / 100).toFixed(1), a0+(1*25), ((i+j)*115) + y1+2 + (k*20))
                temp = '-'
                if(state[i][j][k].Move==0){temp='H'}
                else if(state[i][j][k].Move==1){temp='L'}
                else if(state[i][j][k].Move==2){temp='R'}
                else if(state[i][j][k].Move==3){temp='B'}
                ctx.fillText(temp, a0+(2*25), ((i+j)*115) + y1+2 + (k*20))

                ctx.fillText(state[i][j][k].H.SERIAL, a1 +(25*0), ((i+j)*115) + y1+2 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].H.HP, a1 +(25*1), ((i+j)*115) + y1+2 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].H.ATK, a1 +(25*2), ((i+j)*115) + y1+2 + (k*20)) // H ATK
                ctx.fillText(state[i][j][k].H.DEF, a1 +(25*3), ((i+j)*115) + y1+2 + (k*20)) // H DEF
                ctx.fillText(state[i][j][k].H.ACC, a1 +(25*4), ((i+j)*115) + y1+2 + (k*20)) // H ACC
                ctx.fillText(state[i][j][k].H.CRT, a1 +(25*5), ((i+j)*115) + y1+2 + (k*20)) // H CRIT
                ctx.fillText(state[i][j][k].H.MOB, a1 +(25*6), ((i+j)*115) + y1+2 + (k*20)) // H MOBI
                ctx.fillText(state[i][j][k].H.CD, a1 +(25*7), ((i+j)*115) + y1+2 + (k*20)) // H CD
                ctx.fillText(state[i][j][k].H.CLU, a1 +(25*8), ((i+j)*115) + y1+2 + (k*20)) // H CLU
                ctx.fillText(state[i][j][k].H.Use_current+"/"+state[i][j][k].H.Use_outof, a1 +(25*9), ((i+j)*115) + y1+2 + (k*20)) // H USE
                ctx.fillText(state[i][j][k].H.Weight, a1 +(25*10), ((i+j)*115) + y1+2 + (k*20)) // H WEIGHT

                ctx.fillText("L", a2 -(20*1), ((i+j)*115) + y1-3 + (k*20)) // L L
                ctx.fillText(state[i][j][k].L.SERIAL, a2 +(25*0), ((i+j)*115) + y1-3 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].L.HP, a2 +(25*1), ((i+j)*115) + y1-3 + (k*20)) // L HP
                ctx.fillText(state[i][j][k].L.ATK, a2 +(25*2), ((i+j)*115) + y1-3 + (k*20)) // L ATK
                ctx.fillText(state[i][j][k].L.DEF, a2 +(25*3), ((i+j)*115) + y1-3 + (k*20)) // L DEF
                ctx.fillText(state[i][j][k].L.ACC, a2 +(25*4), ((i+j)*115) + y1-3 + (k*20)) // L ACC
                ctx.fillText(state[i][j][k].L.CRT, a2 +(25*5), ((i+j)*115) + y1-3 + (k*20)) // L CRIT
                ctx.fillText(state[i][j][k].L.MOB, a2 +(25*6), ((i+j)*115) + y1-3 + (k*20)) // L MOBI
                ctx.fillText(state[i][j][k].L.CD, a2 +(25*7), ((i+j)*115) + y1-3 + (k*20)) // L CD
                ctx.fillText(state[i][j][k].L.CLU, a2 +(25*8), ((i+j)*115) + y1-3 + (k*20)) // L CLU
                ctx.fillText(state[i][j][k].L.Use_current+"/"+state[i][j][k].L.Use_outof, a2 +(25*9), ((i+j)*115) + y1-3 + (k*20)) // L USE
                ctx.fillText(state[i][j][k].L.Weight, a2 +(25*10), ((i+j)*115) + y1-3 + (k*20)) // L WEIGHT

                ctx.fillText("R", a2 -(20*1), ((i+j)*115) + y1+7 + (k*20))   // R L
                ctx.fillText(state[i][j][k].R.SERIAL, a2 +(25*0), ((i+j)*115) + y1+7 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].R.HP, a2 +(25*1), ((i+j)*115) + y1+7 + (k*20)) // R HP
                ctx.fillText(state[i][j][k].R.ATK, a2 +(25*2), ((i+j)*115) + y1+7 + (k*20)) // R ATK
                ctx.fillText(state[i][j][k].R.DEF, a2 +(25*3), ((i+j)*115) + y1+7 + (k*20)) // R DEF
                ctx.fillText(state[i][j][k].R.ACC, a2 +(25*4), ((i+j)*115) + y1+7 + (k*20)) // R ACC
                ctx.fillText(state[i][j][k].R.CRT, a2 +(25*5), ((i+j)*115) + y1+7 + (k*20)) // R CRIT
                ctx.fillText(state[i][j][k].R.MOB, a2 +(25*6), ((i+j)*115) + y1+7 + (k*20)) // R MOBI
                ctx.fillText(state[i][j][k].R.CD, a2 +(25*7), ((i+j)*115) + y1+7 + (k*20)) // R CD
                ctx.fillText(state[i][j][k].R.CLU, a2 +(25*8), ((i+j)*115) + y1+7 + (k*20)) // R CLU
                ctx.fillText(state[i][j][k].R.Use_current+"/"+state[i][j][k].R.Use_outof, a2 +(25*9), ((i+j)*115) + y1+7 + (k*20)) // R USE
                ctx.fillText(state[i][j][k].R.Weight, a2 +(25*10), ((i+j)*115) + y1+7 + (k*20)) // R WEIGHT


                ctx.fillText(state[i][j][k].B.SERIAL, a3 +(25*0), ((i+j)*115) + y1+2 + (k*20)) // LEG HP
                ctx.fillText(state[i][j][k].B.HP, a3 +(25*1), ((i+j)*115) + y1+2 + (k*20)) // LEG HP
                ctx.fillText(state[i][j][k].B.ATK, a3 +(25*2), ((i+j)*115) + y1+2 + (k*20)) // LEG ATK
                ctx.fillText(state[i][j][k].B.DEF, a3 +(25*3), ((i+j)*115) + y1+2 + (k*20)) // LEG DEF
                ctx.fillText(state[i][j][k].B.ACC, a3 +(25*4), ((i+j)*115) + y1+2 + (k*20)) // LEG ACC
                ctx.fillText(state[i][j][k].B.CRT, a3 +(25*5), ((i+j)*115) + y1+2 + (k*20)) // LEG CRIT
                ctx.fillText(state[i][j][k].B.MOB, a3 +(25*6), ((i+j)*115) + y1+2 + (k*20)) // LEG MOBI
                ctx.fillText(state[i][j][k].B.CD, a3 +(25*7), ((i+j)*115) + y1+2 + (k*20)) // LEG CD
                ctx.fillText(state[i][j][k].B.CLU, a3 +(25*8), ((i+j)*115) + y1+2 + (k*20)) // LEG  CLU
                ctx.fillText(state[i][j][k].B.Use_current+"/"+state[i][j][k].B.Use_outof, a3 +(25*9), ((i+j)*115) + y1+2 + (k*20)) // LEG USE
                ctx.fillText(state[i][j][k].B.Weight, a3 +(25*10), ((i+j)*115) + y1+2 + (k*20)) // LEG WEIGHT
                ctx.fillText(state[i][j][k].B.DOG, a3 +(25*11), ((i+j)*115) + y1+2 + (k*20)) // LEG DOG
                ctx.fillText(state[i][j][k].B.SPD, a3 +(25*12), ((i+j)*115) + y1+2 + (k*20)) // LEG SPE
                ctx.fillText(state[i][j][k].B.ACL, a3 +(25*13), ((i+j)*115) + y1+2 + (k*20)) // LEG ACL
                ctx.fillText(state[i][j][k].B.ANT, a3 +(25*14), ((i+j)*115) + y1+2 + (k*20)) // LEG ANT
                ctx.fillText(state[i][j][k].B.END, a3 +(25*15), ((i+j)*115) + y1+2 + (k*20)) // LEG END
                
            }
        }
    }
}

function drawLog() {
    for(i =0; i<event_log.length;i++) {
        ctx.textAlign = "left";
        ctx.fillStyle = 'black';
        ctx.fillRect(0,4000 + (10*i),ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(event_log[i],0, 4000 + (10*i));
    }
}

function drawStarters() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {


        let x = 60
        let y1 = 0
        let x_offset = 25
        let x_offset_2 = 250
        let x_offset_3 = 475


        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "center";
        ctx.fillRect(25 + x_offset, y1+83, ctx.measureText("STARTERS (HEAD)").width, 1);
        ctx.fillText("STARTERS (HEAD)",70 + x_offset, y1+ 80 );

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "center";
        ctx.fillRect(25 + x_offset_2, y1+83, ctx.measureText("STARTERS (LEFT ARM)").width, 1);
        ctx.fillText("STARTERS (LEFT ARM)",85 + x_offset_2, y1+ 80 );

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "center";
        ctx.fillRect(25 + x_offset_3, y1+83, ctx.measureText("STARTERS (RIGHT ARM)").width, 1);
        ctx.fillText("STARTERS (RIGHT ARM)",85 + x_offset_3, y1+ 80 );

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "left";

        for(let i = 0; i < state[my_team][my_int].length; i++) {

            s   = state[my_team][my_int][i].H.SERIAL
            name = state[my_team][my_int][i].H.NAME
            HP  = "HP:  " + state[my_team][my_int][i].H.HP
            ATK = "ATK: " + state[my_team][my_int][i].H.ATK
            DEF = "DEF: " + state[my_team][my_int][i].H.DEF
            ACC = "ACC: " + state[my_team][my_int][i].H.ACC
            CRT = "CRT: " + state[my_team][my_int][i].H.CRT
            MOB = "MOB: " + state[my_team][my_int][i].H.MOB
            CD  = "CD:  " + state[my_team][my_int][i].H.CD
            CLU = "CLU: " + state[my_team][my_int][i].H.CLU
            W   = "Weight: " + state[my_team][my_int][i].H.Weight
            AVG = "AVG: " +
                Math.floor(
                    (parseInt(state[my_team][my_int][i].H.HP)+
                        parseInt(state[my_team][my_int][i].H.ATK)+
                        parseInt(state[my_team][my_int][i].H.DEF) +
                        parseInt(state[my_team][my_int][i].H.ACC) +
                        parseInt(state[my_team][my_int][i].H.CRT) +
                        parseInt(state[my_team][my_int][i].H.MOB) +
                        parseInt(state[my_team][my_int][i].H.CD) +
                        parseInt(state[my_team][my_int][i].H.CLU)
                    ) / 8)

            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(MOB,x*2 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(CD,x*0 + x_offset, y1+ 130 + (60*i));
            ctx.fillText(CLU,x*1 + x_offset, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset, y1+ 140 + (60*i));

            s   = state[my_team][my_int][i].L.SERIAL
            name = state[my_team][my_int][i].L.NAME
            HP  = "HP:  " + state[my_team][my_int][i].L.HP
            ATK = "ATK: " + state[my_team][my_int][i].L.ATK
            DEF = "DEF: " + state[my_team][my_int][i].L.DEF
            ACC = "ACC: " + state[my_team][my_int][i].L.ACC
            CRT = "CRT: " + state[my_team][my_int][i].L.CRT
            MOB = "MOB: " + state[my_team][my_int][i].L.MOB
            CD  = "CD:  " + state[my_team][my_int][i].L.CD
            CLU = "CLU: " + state[my_team][my_int][i].L.CLU
            W   = "Weight: " + state[my_team][my_int][i].L.Weight
            AVG = "AVG: " +
                Math.floor(
                    (parseInt(state[my_team][my_int][i].L.HP)+
                        parseInt(state[my_team][my_int][i].L.ATK)+
                        parseInt(state[my_team][my_int][i].L.DEF) +
                        parseInt(state[my_team][my_int][i].L.ACC) +
                        parseInt(state[my_team][my_int][i].L.CRT) +
                        parseInt(state[my_team][my_int][i].L.MOB) +
                        parseInt(state[my_team][my_int][i].L.CD) +
                        parseInt(state[my_team][my_int][i].L.CLU)
                    ) / 8)

            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset_2,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset_2, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset_2, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(MOB,x*2 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(CD,x*0 + x_offset_2, y1+ 130 + (60*i));
            ctx.fillText(CLU,x*1 + x_offset_2, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset_2, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset_2, y1+ 140 + (60*i));

            s   = state[my_team][my_int][i].R.SERIAL
            name = state[my_team][my_int][i].R.NAME
            HP  = "HP:  " + state[my_team][my_int][i].R.HP
            ATK = "ATK: " + state[my_team][my_int][i].R.ATK
            DEF = "DEF: " + state[my_team][my_int][i].R.DEF
            ACC = "ACC: " + state[my_team][my_int][i].R.ACC
            CRT = "CRT: " + state[my_team][my_int][i].R.CRT
            MOB = "MOB: " + state[my_team][my_int][i].R.MOB
            CD  = "CD:  " + state[my_team][my_int][i].R.CD
            CLU = "CLU: " + state[my_team][my_int][i].R.CLU
            W   = "Weight: " + state[my_team][my_int][i].R.Weight
            AVG = "AVG: " +
                Math.floor(
                    (parseInt(state[my_team][my_int][i].R.HP)+
                        parseInt(state[my_team][my_int][i].R.ATK)+
                        parseInt(state[my_team][my_int][i].R.DEF) +
                        parseInt(state[my_team][my_int][i].R.ACC) +
                        parseInt(state[my_team][my_int][i].R.CRT) +
                        parseInt(state[my_team][my_int][i].R.MOB) +
                        parseInt(state[my_team][my_int][i].R.CD) +
                        parseInt(state[my_team][my_int][i].R.CLU)
                    ) / 8)

            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset_2,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset_2, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset_2, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset_2, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(MOB,x*2 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(CD,x*0 + x_offset_2, y1+ 130 + (60*i));
            ctx.fillText(CLU,x*1 + x_offset_2, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset_2, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset_2, y1+ 140 + (60*i));



        }
    }
}
function drawBench() {

    if((bench!=null) && (my_team != -1) && (my_int != -1)) {

        let x = 60
        let y1 = 350
        let x_offset = 25

        let key = my_team + ';' + my_int

        for (let i=0;i<bench[key].length;i++) {
            s   = "H000000"
            name = bench[key][i]["NAME"]
            HP  = "HP:  " + bench[key][i]["HP"]
            ATK = "ATK: " + bench[key][i]["ATK"]
            DEF = "DEF: " + bench[key][i]["DEF"]
            ACC = "ACC: " + bench[key][i]["ACC"]
            CRT = "CRT: " + bench[key][i]["CRT"]
            MOB = "MOB: " + bench[key][i]["MOB"]
            CD  = "CD:  " + bench[key][i]["CD"]
            CLU = "CLU: " + bench[key][i]["CLU"]
            W   = "Weight: " + bench[key][i]["Weight"]
            AVG = "AVG: " +
                Math.floor(
                    (parseInt(bench[key][i]["HP"])+
                        parseInt(bench[key][i]["ATK"])+
                        parseInt(bench[key][i]["DEF"]) +
                        parseInt(bench[key][i]["ACC"]) +
                        parseInt(bench[key][i]["CRT"]) +
                        parseInt(bench[key][i]["MOB"]) +
                        parseInt(bench[key][i]["CD"]) +
                        parseInt(bench[key][i]["CLU"])
                    ) / 8)



            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(MOB,x*2 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(CD,x*0 + x_offset, y1+ 130 + (60*i));
            ctx.fillText(CLU,x*1 + x_offset, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset, y1+ 140 + (60*i));

            if(i == 4) {i = bench[key].length} // only show 5 parts
        }

        for(let i = 0; i<bench[key].length;i++){
            let a = (i%5) * 5
            let b = Math.floor(i/5)*5

            ctx.textAlign = "left";
            ctx.fillStyle = 'green';
            ctx.fillRect(70 + a + x_offset,775+b,3,3);
        }

        if(h_index + 5 < bench[key].length) {
            ctx.textAlign = "left";
            ctx.fillStyle = 'white';
            ctx.font = '11px monospace';
            ctx.fillText("▼",80 + x_offset, y1+400);
        }
        if(h_index > 0) {
            ctx.textAlign = "left";
            ctx.fillStyle = 'white';
            ctx.font = '11px monospace';
            ctx.fillText("▲",80 + x_offset, y1+200);
        }

    }


}

function dashedLine(x1,y1,x2,y2,dashArr){
    // get the normalised line vector from start to end
    var nx = x2 - x1;
    var ny = y2 - y1;
    const dist = Math.sqrt(nx * nx + ny * ny);  // get the line length
    nx /= dist;
    ny /= dist;
    var dashIdx = 0;  // the index into the dash array
    var i = 0;        // the current line position in pixels
    ctx.beginPath();  // start a path
    ctx.strokeStyle = 'yellow';
    ctx.lineWidth = 1;
    while(i < dist){   // do while less than line length
        // get the line seg dash length
        var dashLen = dashArr[(dashIdx ++) % dashArr.length];
        // draw the dash
        ctx.moveTo(x1 + nx * i, y1 + ny * i);
        i = Math.min(dist,i + dashLen);
        ctx.lineTo(x1 + nx * i, y1 + ny * i);
        // add the spacing
        i += dashArr[(dashIdx ++) % dashArr.length];
        if(i <= 0) { // something is wrong so exit rather than endless loop
            break;
        }
    }
    ctx.stroke();  // stroke
}