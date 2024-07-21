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

        drawGrid()
        drawPhase()
        drawLog()
        //drawBench()
        //drawStarters()
        drawError()
        //drawTable1()
        //drawTable2()
        drawStartersTable1()
        drawBenchTable()
        //drawStartersTable2()
        //drawSelect()
        //ctx.fillRect(1150,40 , 1, 800);


    } else if (match_data.length > 0) {

        drawGrid()
        //drawBoard()
        drawSinglePoint(0,0,c.width/2,c.height/2,'red');

        let cent = calculateCenter(c.width/2,c.height/2);
        drawSinglePoint(cent[0],cent[1],c.width/2,c.height/2,'yellow');


        drawCircle(c.width/2,c.height/2,250);
        drawPoint(c.width/2,c.height/2,250,10,)


        drawPos()
        drawState()
        //drawStats()
        drawTurn()
        drawUnitsOfTime()
        drawStartersTable2()
        drawLog()
        drawHeatMap()

    }

}
anime()

function drawPos()  {

    let w = Math.round(c.width/2)-250
    let h = Math.round(c.height/2)-250

    let y_offset = 200
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
                                state[i][j][k].Position = state[i][j][k].Position - (state[i][j][k].B.SPD * 0.02)
                                if(
                                    (match_data[animating_state][i][j][k].Direction == 0) &&
                                    (state[i][j][k].Position < match_data[animating_state][i][j][k].Position)
                                ) {
                                    state[i][j][k].Position = match_data[animating_state][i][j][k].Position
                                }
                                else if(state[i][j][k].Position < 0){state[i][j][k].Position = 0}
                            } else if(state[i][j][k].Direction == 1) {
                                state[i][j][k].Position = state[i][j][k].Position + (state[i][j][k].B.SPD * 0.02)
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

        if(draw_attacks>0){draw_attacks = 59}

        let a1 = 0
        let a2 = 0
        for(let i = 0; i < state.length; i++) {
            for(let j = 0; j < state[i].length; j++) {
                for(let k = 0; k < state[i][j].length; k++) {
                    if( (state[i][j][k].H.HP > 0) && (state[i][j][k].Position==100) ) {

                        if(a1 < atk_data[animating_state].length) {
                            let b = JSON.parse(atk_data[animating_state][a1])
                            let a_0 = b.Attacker[0]
                            let a_1 = b.Attacker[1]
                            let a_2 = b.Attacker[2]
                            let x1 = w+(250*a_0)+(250*(state[a_0][a_1][a_2].Position/100))
                            let y1 = h+(a_2 * 125)+(a_1*700)
                            if((a_0%2)==1){
                                x1 = w+(250*a_0)+250-(250*(state[a_0][a_1][a_2].Position/100))
                            }

                            let perc = draw_attacks / 30
                            if(draw_attacks>30){
                                perc = 1
                            }

                            // calculate a vector to draw
                            let b_0 = b.Defender[0][0]
                            let b_1 = b.Defender[0][1]
                            let b_2 = b.Defender[0][2]
                            let x2 = w+(250*b_0)+(250*(state[b_0][b_1][b_2].Position/100))
                            let y2 = h+(b_2 * 125)+(b_1*700)
                            if((b_0%2)==1){
                                x2 = w+(250*b_0)+250-(250*(state[b_0][b_1][b_2].Position/100))
                            }
                            let x2_x1 = (x2 - x1)*perc
                            let y2_y1 = (y2 - y1)*perc
                            //dashedLine(x1,y1,x1+x2_x1,y1+y2_y1,[1,10])

                            if(draw_attacks>=30) {
                                ctx.fillStyle = 'white'
                                ctx.textAlign = "center";
                                ctx.font = '9px monospace';
                                //ctx.fillText("BOOM!",x2, y2-7);
                            }

                            a1++
                        }

                        if ((draw_attacks==59) && (a2 < atk_data[animating_state].length)) {
                            event_log.push(atk_data[animating_state][a2])
                            if(event_log.length >= 25) {
                                event_log.shift()
                            }
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
            if(event_log.length >= 25) {
                event_log.shift()
            }
        }
    }

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

                if(i%2==0) {
                    if(symbol=='d'){symbol=''}
                    ctx.textAlign = "center";
                    //ctx.fillRect(700+(270*i),y_offset+(k*120)+(j*700),250,1);
                    //ctx.fillText(symbol, w+(250*i)+(250*(state[i][j][k].Position/100)), h+(k*125)+(j*700)+3 );

                    ctx.fillStyle = 'white';
                    ctx.textAlign = "center";
                    ctx.font = '8px monospace';
                    //ctx.fillText(k, w+(250*i)+(250*(state[i][j][k].Position/100)), h+(k*125)+(j*700)+25 );
                    //ctx.fillText(state[i][j][k].Position.toFixed(1), w+(250*i)+(250*(state[i][j][k].Position/100)), h+(k*125)+(j*700)+15 );
                    ctx.font = '11px monospace';

                } else {
                    if(symbol=='>'){symbol='<'}
                    else if(symbol=='<'){symbol='>'}
                    else if(symbol=='d'){symbol=''}
                    ctx.textAlign = "center";
                    //ctx.fillRect(700+(270*i),y_offset+(k*120)+(j*700),250,1);
                    //ctx.fillText(symbol, w+(250*i)+250-(250*(state[i][j][k].Position/100)), h+(k*125)+(j*700)+3 );

                    ctx.fillStyle = 'white';
                    ctx.textAlign = "left";
                    ctx.font = '8px monospace';
                    //ctx.fillText(k, w+(250*i)+250-(250*(state[i][j][k].Position/100)), h+(k*125)+(j*700)+25 );
                    //ctx.fillText( state[i][j][k].Position.toFixed(1), w+(250*i)+250-(250*(state[i][j][k].Position/100))-5, h+(k*125)+(j*700)+15 );
                    ctx.font = '11px monospace';
                }

            }
        }
    }


}

function drawState() {

    let test1 = c.width/2
    let test2 = c.height/2

    x0 = Math.round(c.width/2)-250 - 45
    y0 = Math.round(c.height/2)-250
    let space = 590

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
                ctx.fillRect(x0+(space*i), y0-9 + (k*125)+(j*700), 5, 5);
                if (state[i][j][k].L.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}

                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0-5+(space*i), y0-4 + (k*125)+(j*700), 5, 7); // make it 270,5,7 OR 269,6,4
                if (state[i][j][k].R.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0+5+(space*i), y0-4 + (k*125)+(j*700), 5, 7);
                if (state[i][j][k].B.HP > 0) {
                    if(i==my_team && j==my_int) {ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}

                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(x0+(space*i), y0+1 + (k*125)+(j*700), 5, 7);

                // draw move
                ctx.fillStyle = 'white';
                ctx.textAlign = "left";
                let temp = '-'
                if(state[i][j][k].H.HP>0) {
                    if(state[i][j][k].Move==0){temp='H'}
                    else if(state[i][j][k].Move==1){temp='L'}
                    else if(state[i][j][k].Move==2){temp='R'}
                    else if(state[i][j][k].Move==3){temp='B'}
                }
                ctx.fillText(temp, x0+(space*i)+20, y0+1 + (k*125)+(j*700));

                if(turn.substring(0,4) == "TURN") {
                    if (
                        state[i][j][k].H.HP > 0 &&
                        state[i][j][k].Position == 0
                    ) {
                        ctx.fillStyle = 'white'
                        ctx.textAlign = "center";
                        if((i == indexer_t) && (j == indexer_u) && (k == indexer_b)) {
                            ctx.fillText("->", x0-15+(space*i),  y0+2 + (k*125)+(j*700) );

                            let p1 = 16*Math.PI/10
                            if(i==1){
                                p1 = (16*Math.PI/10) + (2*k*Math.PI/10)
                            }
                            else {
                                p1 = (14*Math.PI/10) - (2*k*Math.PI/10)
                            }
                            let x1 = Math.cos(p1);
                            let y1 = Math.sin(p1);

                            for(let a = 0; a < state.length; a++) {
                                for(let b = 0; b < state[a].length; b++) {
                                    for(let c = 0; c < state[a][b].length; c++) {
                                        if(a==i && b==j && c==k) {

                                        }
                                        else {
                                            let p2 = 16*Math.PI/10
                                            if(a==1){
                                                p2 = (16*Math.PI/10) + (2*c*Math.PI/10)
                                            }
                                            else {
                                                p2 = (14*Math.PI/10) - (2*c*Math.PI/10)
                                            }
                                            let x2 = Math.cos(p2);
                                            let y2 = Math.sin(p2);

                                            ctx.beginPath();
                                            ctx.moveTo( (x1*250)+test1, (y1*250)+test2 );
                                            ctx.lineTo((x2*250)+test1, (y2*250)+test2);
                                            ctx.strokeStyle = "grey";
                                            ctx.lineWidth = 1;
                                            ctx.stroke();
                                        }
                                    }
                                }
                            }


                        }
                        else if((state[indexer_t][indexer_u][indexer_b].Position != 0) && (my_team==i) && (my_int==j)) { // finds
                            indexer_t = i;
                            indexer_u = j;
                            indexer_b = k;
                            ctx.fillText("->", x0-15+(space*i),  y0+2 + (k*125)+(j*700) );
                        }
                        else {
                            ctx.fillText("*", x0-15+(space*i),  y0+2 + (k*125)+(j*700) );
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

    a0 = 20

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
    ctx.fillText("CD", a1 +(25*6), y0+20) // H CD
    ctx.fillText("CLU", a1 +(25*7), y0+20) // H CLU
    ctx.fillText("USE", a1 +(25*8), y0+20) // H USE
    ctx.fillText("WGH", a1 +(25*9), y0+20) // H WEIGHT

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
    ctx.fillText("CD", a2 +(25*6), y0+20) // ARM CD
    ctx.fillText("CLU", a2 +(25*7), y0+20) // ARM CLU
    ctx.fillText("USE", a2 +(25*8), y0+20) // ARM USE
    ctx.fillText("WGT", a2 +(25*9), y0+20) // ARM WEIGHT

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
    ctx.fillText("CD", a3 +(25*6), y0+20) // LEG CD
    ctx.fillText("CLU", a3 +(25*7), y0+20) // LEG  CLU
    ctx.fillText("USE", a3 +(25*8), y0+20) // LEG USE
    ctx.fillText("WGT", a3 +(25*9), y0+20) // LEG WEIGHT
    ctx.fillText("DOG", a3 +(25*10), y0+20) // LEG DOG
    ctx.fillText("SPD", a3 +(25*11), y0+20) // LEG SPE

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

                /*
                ctx.fillText(state[i][j][k].H.SERIAL, a1 +(25*0), ((i+j)*115) + y1+2 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].H.HP, a1 +(25*1), ((i+j)*115) + y1+2 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].H.ATK, a1 +(25*2), ((i+j)*115) + y1+2 + (k*20)) // H ATK
                ctx.fillText(state[i][j][k].H.DEF, a1 +(25*3), ((i+j)*115) + y1+2 + (k*20)) // H DEF
                ctx.fillText(state[i][j][k].H.ACC, a1 +(25*4), ((i+j)*115) + y1+2 + (k*20)) // H ACC
                ctx.fillText(state[i][j][k].H.CRT, a1 +(25*5), ((i+j)*115) + y1+2 + (k*20)) // H CRIT
                ctx.fillText(state[i][j][k].H.CD, a1 +(25*6), ((i+j)*115) + y1+2 + (k*20)) // H CD
                ctx.fillText(state[i][j][k].H.CLU, a1 +(25*7), ((i+j)*115) + y1+2 + (k*20)) // H CLU
                ctx.fillText(state[i][j][k].H.Use_current+"/"+state[i][j][k].H.Use_outof, a1 +(25*8), ((i+j)*115) + y1+2 + (k*20)) // H USE
                ctx.fillText(state[i][j][k].H.Weight, a1 +(25*9), ((i+j)*115) + y1+2 + (k*20)) // H WEIGHT

                ctx.fillText("L", a2 -(20*1), ((i+j)*115) + y1-3 + (k*20)) // L L
                ctx.fillText(state[i][j][k].L.SERIAL, a2 +(25*0), ((i+j)*115) + y1-3 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].L.HP, a2 +(25*1), ((i+j)*115) + y1-3 + (k*20)) // L HP
                ctx.fillText(state[i][j][k].L.ATK, a2 +(25*2), ((i+j)*115) + y1-3 + (k*20)) // L ATK
                ctx.fillText(state[i][j][k].L.DEF, a2 +(25*3), ((i+j)*115) + y1-3 + (k*20)) // L DEF
                ctx.fillText(state[i][j][k].L.ACC, a2 +(25*4), ((i+j)*115) + y1-3 + (k*20)) // L ACC
                ctx.fillText(state[i][j][k].L.CRT, a2 +(25*5), ((i+j)*115) + y1-3 + (k*20)) // L CRIT
                ctx.fillText(state[i][j][k].L.CD, a2 +(25*6), ((i+j)*115) + y1-3 + (k*20)) // L CD
                ctx.fillText(state[i][j][k].L.CLU, a2 +(25*7), ((i+j)*115) + y1-3 + (k*20)) // L CLU
                ctx.fillText(state[i][j][k].L.Use_current+"/"+state[i][j][k].L.Use_outof, a2 +(25*8), ((i+j)*115) + y1-3 + (k*20)) // L USE
                ctx.fillText(state[i][j][k].L.Weight, a2 +(25*9), ((i+j)*115) + y1-3 + (k*20)) // L WEIGHT

                ctx.fillText("R", a2 -(20*1), ((i+j)*115) + y1+7 + (k*20))   // R L
                ctx.fillText(state[i][j][k].R.SERIAL, a2 +(25*0), ((i+j)*115) + y1+7 + (k*20)) // H HP
                ctx.fillText(state[i][j][k].R.HP, a2 +(25*1), ((i+j)*115) + y1+7 + (k*20)) // R HP
                ctx.fillText(state[i][j][k].R.ATK, a2 +(25*2), ((i+j)*115) + y1+7 + (k*20)) // R ATK
                ctx.fillText(state[i][j][k].R.DEF, a2 +(25*3), ((i+j)*115) + y1+7 + (k*20)) // R DEF
                ctx.fillText(state[i][j][k].R.ACC, a2 +(25*4), ((i+j)*115) + y1+7 + (k*20)) // R ACC
                ctx.fillText(state[i][j][k].R.CRT, a2 +(25*5), ((i+j)*115) + y1+7 + (k*20)) // R CRIT
                ctx.fillText(state[i][j][k].R.CD, a2 +(25*6), ((i+j)*115) + y1+7 + (k*20)) // R CD
                ctx.fillText(state[i][j][k].R.CLU, a2 +(25*7), ((i+j)*115) + y1+7 + (k*20)) // R CLU
                ctx.fillText(state[i][j][k].R.Use_current+"/"+state[i][j][k].R.Use_outof, a2 +(25*8), ((i+j)*115) + y1+7 + (k*20)) // R USE
                ctx.fillText(state[i][j][k].R.Weight, a2 +(25*9), ((i+j)*115) + y1+7 + (k*20)) // R WEIGHT


                ctx.fillText(state[i][j][k].B.SERIAL, a3 +(25*0), ((i+j)*115) + y1+2 + (k*20)) // LEG HP
                ctx.fillText(state[i][j][k].B.HP, a3 +(25*1), ((i+j)*115) + y1+2 + (k*20)) // LEG HP
                ctx.fillText(state[i][j][k].B.ATK, a3 +(25*2), ((i+j)*115) + y1+2 + (k*20)) // LEG ATK
                ctx.fillText(state[i][j][k].B.DEF, a3 +(25*3), ((i+j)*115) + y1+2 + (k*20)) // LEG DEF
                ctx.fillText(state[i][j][k].B.ACC, a3 +(25*4), ((i+j)*115) + y1+2 + (k*20)) // LEG ACC
                ctx.fillText(state[i][j][k].B.CRT, a3 +(25*5), ((i+j)*115) + y1+2 + (k*20)) // LEG CRIT
                ctx.fillText(state[i][j][k].B.CD, a3 +(25*6), ((i+j)*115) + y1+2 + (k*20)) // LEG CD
                ctx.fillText(state[i][j][k].B.CLU, a3 +(25*7), ((i+j)*115) + y1+2 + (k*20)) // LEG  CLU
                ctx.fillText(state[i][j][k].B.Use_current+"/"+state[i][j][k].B.Use_outof, a3 +(25*8), ((i+j)*115) + y1+2 + (k*20)) // LEG USE
                ctx.fillText(state[i][j][k].B.Weight, a3 +(25*9), ((i+j)*115) + y1+2 + (k*20)) // LEG WEIGHT
                ctx.fillText(state[i][j][k].B.DOG, a3 +(25*10), ((i+j)*115) + y1+2 + (k*20)) // LEG DOG
                ctx.fillText(state[i][j][k].B.SPD, a3 +(25*11), ((i+j)*115) + y1+2 + (k*20)) // LEG SPE
                */
            }
        }
    }
}

function drawLog() {

    let x = 0
    let y = 800
    for(i =0; i<event_log.length;i++) {
        ctx.textAlign = "left";
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(event_log[i],x, y + (10*i));
    }
}

function drawStarters() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {


        let x = 60
        let y1 = 0
        let x_offset = 75
        let x_offset_2 = 300
        let x_offset_3 = 525
        let x_offset_4 = 750

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
        ctx.textAlign = "center";
        ctx.fillRect(32 + x_offset_4, y1+83, ctx.measureText("STARTERS (BOTTOM)").width, 1);
        ctx.fillText("STARTERS (BOTTOM)",85 + x_offset_4, y1+ 80 );

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "left";

        for(let i = 0; i < state[my_team][my_int].length; i++) {

            // left bracket
            ctx.fillStyle = 'red';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20,90 + (60*i), 1, 55);
            ctx.fillRect(x_offset-20,90 + (60*i), 5, 1);
            ctx.fillRect(x_offset-20,90 + 55 + (60*i), 5, 1);

            // right bracket
            ctx.fillStyle = 'red';
            ctx.textAlign = "left";
            ctx.fillRect(180 + x_offset_4,90 + (60*i), 1, 55);
            ctx.fillRect(175 + x_offset_4,90 + (60*i), 5, 1);
            ctx.fillRect(175 + x_offset_4,90 + 55 + (60*i), 5, 1);

            // printing number
            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillText(i, x_offset-40,90 + 27 + (60*i));


            // HEAD STARTERS
            s   = state[my_team][my_int][i].H.SERIAL.toString()
            for (let k = s.length; k < 7; k++) {
                s = '0' + s
            }
            s = 'H' + s
            name = state[my_team][my_int][i].H.NAME
            HP  = "HP:  " + state[my_team][my_int][i].H.HP
            ATK = "ATK: " + state[my_team][my_int][i].H.ATK
            DEF = "DEF: " + state[my_team][my_int][i].H.DEF
            ACC = "ACC: " + state[my_team][my_int][i].H.ACC
            CRT = "CRT: " + state[my_team][my_int][i].H.CRT
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
                        parseInt(state[my_team][my_int][i].H.CD) +
                        parseInt(state[my_team][my_int][i].H.CLU)
                    ) / 7)

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
            ctx.fillText(CD,x*2 + x_offset, y1+ 120 + (60*i));
            ctx.fillText(CLU,x*0 + x_offset, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset, y1+ 140 + (60*i));

            // LEFT STARTERS
            s   = state[my_team][my_int][i].L.SERIAL.toString()
            for (let k = s.length; k < 7; k++) {
                s = '0' + s
            }
            s = 'L' + s
            name = state[my_team][my_int][i].L.NAME
            HP  = "HP:  " + state[my_team][my_int][i].L.HP
            ATK = "ATK: " + state[my_team][my_int][i].L.ATK
            DEF = "DEF: " + state[my_team][my_int][i].L.DEF
            ACC = "ACC: " + state[my_team][my_int][i].L.ACC
            CRT = "CRT: " + state[my_team][my_int][i].L.CRT
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
                        parseInt(state[my_team][my_int][i].L.CD) +
                        parseInt(state[my_team][my_int][i].L.CLU)
                    ) / 7)

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
            ctx.fillText(CD,x*2 + x_offset_2, y1+ 120 + (60*i));
            ctx.fillText(CLU,x*0 + x_offset_2, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset_2, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset_2, y1+ 140 + (60*i));

            // RIGHT STARTERS
            s   = state[my_team][my_int][i].R.SERIAL.toString()
            for (let k = s.length; k < 7; k++) {
                s = '0' + s
            }
            s = 'R' + s
            name = state[my_team][my_int][i].R.NAME
            HP  = "HP:  " + state[my_team][my_int][i].R.HP
            ATK = "ATK: " + state[my_team][my_int][i].R.ATK
            DEF = "DEF: " + state[my_team][my_int][i].R.DEF
            ACC = "ACC: " + state[my_team][my_int][i].R.ACC
            CRT = "CRT: " + state[my_team][my_int][i].R.CRT
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
                        parseInt(state[my_team][my_int][i].R.CD) +
                        parseInt(state[my_team][my_int][i].R.CLU)
                    ) / 7)

            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset_3,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset_3, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset_3, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset_3, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset_3, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset_3, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset_3, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset_3, y1+ 120 + (60*i));
            ctx.fillText(CD,x*2 + x_offset_3, y1+ 120 + (60*i));
            ctx.fillText(CLU,x*0 + x_offset_3, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset_3, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset_3, y1+ 140 + (60*i));

            // BOTTOM STARTERS
            s   = state[my_team][my_int][i].B.SERIAL.toString()
            for (let k = s.length; k < 7; k++) {
                s = '0' + s
            }
            s = 'B' + s
            name = state[my_team][my_int][i].B.NAME
            HP  = "HP:  " + state[my_team][my_int][i].B.HP
            ATK = "ATK: " + state[my_team][my_int][i].B.ATK
            DEF = "DEF: " + state[my_team][my_int][i].B.DEF
            ACC = "ACC: " + state[my_team][my_int][i].B.ACC
            CRT = "CRT: " + state[my_team][my_int][i].B.CRT
            CD  = "CD:  " + state[my_team][my_int][i].B.CD
            CLU = "CLU: " + state[my_team][my_int][i].B.CLU
            SPD  = "SPD:  " + state[my_team][my_int][i].B.SPD
            DOG = "DOG: " + state[my_team][my_int][i].B.DOG
            W   = "Weight: " + state[my_team][my_int][i].B.Weight
            AVG = "AVG: " +
                Math.floor(
                    (parseInt(state[my_team][my_int][i].B.HP)+
                        parseInt(state[my_team][my_int][i].B.ATK)+
                        parseInt(state[my_team][my_int][i].B.DEF) +
                        parseInt(state[my_team][my_int][i].B.ACC) +
                        parseInt(state[my_team][my_int][i].B.CRT) +
                        parseInt(state[my_team][my_int][i].B.CD) +
                        parseInt(state[my_team][my_int][i].B.CLU)
                    ) / 9)

            ctx.textAlign = "left";
            ctx.fillStyle = 'blue';
            ctx.fillRect(x + x_offset_4,y1+ 94 + (60*i),ctx.measureText(name).width,ctx.measureText('M').width);

            ctx.font = '11px monospace';
            ctx.fillStyle = 'white';
            ctx.fillText(s,x*0 + x_offset_4, y1+ 100 + (60*i));
            ctx.fillText(name,x*1 + x_offset_4, y1+ 100 + (60*i));
            ctx.fillText(HP,x*0 + x_offset_4, y1+ 110 + (60*i));
            ctx.fillText(ATK,x*1 + x_offset_4, y1+ 110 + (60*i));
            ctx.fillText(DEF,x*2 + x_offset_4, y1+ 110 + (60*i));
            ctx.fillText(ACC,x*0 + x_offset_4, y1+ 120 + (60*i));
            ctx.fillText(CRT,x*1 + x_offset_4, y1+ 120 + (60*i));
            ctx.fillText(CD,x*2 + x_offset_4, y1+ 120 + (60*i));
            ctx.fillText(CLU,x*0 + x_offset_4, y1+ 130 + (60*i));
            ctx.fillText(SPD,x*1 + x_offset_4, y1+ 130 + (60*i));
            ctx.fillText(DOG,x*2 + x_offset_4, y1+ 130 + (60*i));
            ctx.fillText(W,x*0 + x_offset_4, y1+ 140 + (60*i));
            ctx.fillText(AVG,x*2 + x_offset_4, y1+ 140 + (60*i));

        }
    }
}

function drawSelect() {

    ctx.beginPath();
    ctx.lineWidth = "1";
    ctx.strokeStyle = "red";
    ctx.rect(72+(225*x_index_1), 90+(60*y_index_1)+(350*screen_index), 170, 55);
    ctx.stroke();

    if(second_index==true) {

        if(screen_index_2==0) {
            ctx.beginPath();
            ctx.lineWidth = "1";
            ctx.strokeStyle = "pink";
            ctx.rect(72+(225*x_index_2), 90+(60*y_index_2)+(350*screen_index_2), 170, 55);
            ctx.stroke();
        } else {

            if(x_index_2==0) {
                if((y_index_2<h_index+5) && (y_index_2>=h_index)) {
                    ctx.beginPath();
                    ctx.lineWidth = "1";
                    ctx.strokeStyle = "pink";
                    ctx.rect(72+(225*x_index_2), 90+(60*(y_index_2-h_index))+(350*screen_index_2), 170, 55);
                    ctx.stroke();
                }
            } else if(x_index_2==1) {
                if((y_index_2<l_index+5) && (y_index_2>=l_index)) {
                    ctx.beginPath();
                    ctx.lineWidth = "1";
                    ctx.strokeStyle = "pink";
                    ctx.rect(72+(225*x_index_2), 90+(60*(y_index_2-l_index))+(350*screen_index_2), 170, 55);
                    ctx.stroke();
                }
            } else if(x_index_2==2) {
                if((y_index_2<r_index+5) && (y_index_2>=r_index)) {
                    ctx.beginPath();
                    ctx.lineWidth = "1";
                    ctx.strokeStyle = "pink";
                    ctx.rect(72+(225*x_index_2), 90+(60*(y_index_2-r_index))+(350*screen_index_2), 170, 55);
                    ctx.stroke();
                }

            } else if(x_index_2==3) {
                if((y_index_2<b_index+5) && (y_index_2>=b_index)) {
                    ctx.beginPath();
                    ctx.lineWidth = "1";
                    ctx.strokeStyle = "pink";
                    ctx.rect(72+(225*x_index_2), 90+(60*(y_index_2-b_index))+(350*screen_index_2), 170, 55);
                    ctx.stroke();
                }
            }


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

function drawError(){
    if(err_msg_index == 0) {
        return
    }
    if(err_msg_index == 1) {
        ctx.font = '11px monospace';
        ctx.fillStyle = 'red';
        ctx.fillText("CANNOT SWAP INCOMPATIBLE BODY PARTS",50, 50);
        err_count = err_count + 1
    }

    if(err_count==50) {
        err_msg_index = 0
        err_count = 0
    }
}

function drawGrid() {
    for(i=0;i<200;i++) {
        for(j=0;j<100;j++) {
            ctx.textAlign = "left";
            ctx.fillStyle = 'grey';
            ctx.fillRect((i*10),(j*10),1,1);
        }
    }
}

function drawTable1() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {

        let x_offset = 1000
        let y_offset = 100

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.textAlign = "center";

        ctx.fillStyle = 'red';
        ctx.fillText("0",x_offset-20, y_offset )
        ctx.fillStyle = 'white';

        ctx.fillText("H",x_offset+5, y_offset );
        ctx.fillText("L",x_offset+25, y_offset )
        ctx.fillText("R",x_offset+45, y_offset )
        ctx.fillText("B",x_offset+65, y_offset )
        ctx.fillText("AVG",x_offset+90, y_offset )

        ctx.fillText("HP",x_offset-20, y_offset+20 )
        ctx.fillText("ATK",x_offset-20, y_offset+40 )
        ctx.fillText("DEF",x_offset-20, y_offset+60 )
        ctx.fillText("ACC",x_offset-20, y_offset+80 )
        ctx.fillText("CRT",x_offset-20, y_offset+100 )
        ctx.fillText("CD",x_offset-20, y_offset+120 )
        ctx.fillText("CLU",x_offset-20, y_offset+140 )
        ctx.fillText("SPD",x_offset-20, y_offset+160 )
        ctx.fillText("DOG",x_offset-20, y_offset+180 )
        ctx.fillText("WGT",x_offset-20, y_offset+200 )

        ctx.fillText(state[my_team][my_int][0].H.HP,x_offset+5, y_offset+20 )
        ctx.fillText(state[my_team][my_int][0].H.ATK,x_offset+5, y_offset+40 )
        ctx.fillText(state[my_team][my_int][0].H.DEF,x_offset+5, y_offset+60 )
        ctx.fillText(state[my_team][my_int][0].H.ACC,x_offset+5, y_offset+80 )
        ctx.fillText(state[my_team][my_int][0].H.CRT,x_offset+5, y_offset+100 )
        ctx.fillText(state[my_team][my_int][0].H.CD,x_offset+5, y_offset+120 )
        ctx.fillText(state[my_team][my_int][0].H.CLU,x_offset+5, y_offset+140 )
        ctx.fillText(state[my_team][my_int][0].H.Weight,x_offset+5, y_offset+200 )

        ctx.fillText(state[my_team][my_int][0].L.HP,x_offset+25, y_offset+20 )
        ctx.fillText(state[my_team][my_int][0].L.ATK,x_offset+25, y_offset+40 )
        ctx.fillText(state[my_team][my_int][0].L.DEF,x_offset+25, y_offset+60 )
        ctx.fillText(state[my_team][my_int][0].L.ACC,x_offset+25, y_offset+80 )
        ctx.fillText(state[my_team][my_int][0].L.CRT,x_offset+25, y_offset+100 )
        ctx.fillText(state[my_team][my_int][0].L.CD,x_offset+25, y_offset+120 )
        ctx.fillText(state[my_team][my_int][0].L.CLU,x_offset+25, y_offset+140 )
        ctx.fillText(state[my_team][my_int][0].L.Weight,x_offset+25, y_offset+200 )

        ctx.fillText(state[my_team][my_int][0].R.HP,x_offset+45, y_offset+20 )
        ctx.fillText(state[my_team][my_int][0].R.ATK,x_offset+45, y_offset+40 )
        ctx.fillText(state[my_team][my_int][0].R.DEF,x_offset+45, y_offset+60 )
        ctx.fillText(state[my_team][my_int][0].R.ACC,x_offset+45, y_offset+80 )
        ctx.fillText(state[my_team][my_int][0].R.CRT,x_offset+45, y_offset+100 )
        ctx.fillText(state[my_team][my_int][0].R.CD,x_offset+45, y_offset+120 )
        ctx.fillText(state[my_team][my_int][0].R.CLU,x_offset+45, y_offset+140 )
        ctx.fillText(state[my_team][my_int][0].R.Weight,x_offset+45, y_offset+200 )

        ctx.fillText(state[my_team][my_int][0].B.HP,x_offset+65, y_offset+20 )
        ctx.fillText(state[my_team][my_int][0].B.ATK,x_offset+65, y_offset+40 )
        ctx.fillText(state[my_team][my_int][0].B.DEF,x_offset+65, y_offset+60 )
        ctx.fillText(state[my_team][my_int][0].B.ACC,x_offset+65, y_offset+80 )
        ctx.fillText(state[my_team][my_int][0].B.CRT,x_offset+65, y_offset+100 )
        ctx.fillText(state[my_team][my_int][0].B.CD,x_offset+65, y_offset+120 )
        ctx.fillText(state[my_team][my_int][0].B.CLU,x_offset+65, y_offset+140 )
        ctx.fillText(state[my_team][my_int][0].B.SPD,x_offset+65, y_offset+160 )
        ctx.fillText(state[my_team][my_int][0].B.DOG,x_offset+65, y_offset+180 )
        ctx.fillText(state[my_team][my_int][0].B.Weight,x_offset+65, y_offset+200 )

        ctx.fillText(0 |(state[my_team][my_int][0].H.HP+state[my_team][my_int][0].L.HP+state[my_team][my_int][0].R.HP+state[my_team][my_int][0].B.HP)/4,x_offset+90, y_offset+20 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.ATK+state[my_team][my_int][0].L.ATK+state[my_team][my_int][0].R.ATK+state[my_team][my_int][0].B.ATK)/4,x_offset+90, y_offset+40 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.DEF+state[my_team][my_int][0].L.DEF+state[my_team][my_int][0].R.DEF+state[my_team][my_int][0].B.DEF)/4,x_offset+90, y_offset+60 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.ACC+state[my_team][my_int][0].L.ACC+state[my_team][my_int][0].R.ACC+state[my_team][my_int][0].B.ACC)/4,x_offset+90, y_offset+80 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.CRT+state[my_team][my_int][0].L.CRT+state[my_team][my_int][0].R.CRT+state[my_team][my_int][0].B.CRT)/4,x_offset+90, y_offset+100 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.CD+state[my_team][my_int][0].L.CD+state[my_team][my_int][0].R.CD+state[my_team][my_int][0].B.CD)/4,x_offset+90, y_offset+120 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.CLU+state[my_team][my_int][0].L.CLU+state[my_team][my_int][0].R.CLU+state[my_team][my_int][0].B.CLU)/4,x_offset+90, y_offset+140 )
        ctx.fillText(state[my_team][my_int][0].B.SPD,x_offset+90, y_offset+160 )
        ctx.fillText(state[my_team][my_int][0].B.DOG,x_offset+90, y_offset+180 )
        ctx.fillText(0 |(state[my_team][my_int][0].H.Weight+state[my_team][my_int][0].L.Weight+state[my_team][my_int][0].R.Weight+state[my_team][my_int][0].B.Weight)/4,x_offset+90, y_offset+200 )

        ctx.textAlign = "left";
        let ser   = state[my_team][my_int][0].H.SERIAL.toString()
        for (let k = ser.length; k < 7; k++) {
            ser = '0' + ser
        }
        ser = 'H' + ser
        ctx.fillText(ser,x_offset+110, y_offset+10 )
        ctx.fillStyle = 'blue';
        ctx.fillRect(x_offset+110,y_offset+13,ctx.measureText(state[my_team][my_int][0].H.NAME).width,ctx.measureText('M').width);
        ctx.fillStyle = 'white';
        ctx.fillText(state[my_team][my_int][0].H.NAME,x_offset+110, y_offset+20 )
        ctx.fillText("Attack closest enemy",x_offset+110, y_offset+30 )

        ser   = state[my_team][my_int][0].L.SERIAL.toString()
        for (let k = ser.length; k < 7; k++) {
            ser = '0' + ser
        }
        ser = 'L' + ser
        ctx.fillText(ser,x_offset+110, y_offset+60 )
        ctx.fillStyle = 'blue';
        ctx.fillRect(x_offset+110,y_offset+63,ctx.measureText(state[my_team][my_int][0].L.NAME).width,ctx.measureText('M').width);
        ctx.fillStyle = 'white';
        ctx.fillText(state[my_team][my_int][0].L.NAME,x_offset+110, y_offset+70 )
        ctx.fillText("Attack closest enemy",x_offset+110, y_offset+80 )

        ser   = state[my_team][my_int][0].R.SERIAL.toString()
        for (let k = ser.length; k < 7; k++) {
            ser = '0' + ser
        }
        ser = 'R' + ser
        ctx.fillText(ser,x_offset+110, y_offset+110 )
        ctx.fillStyle = 'blue';
        ctx.fillRect(x_offset+110,y_offset+113,ctx.measureText(state[my_team][my_int][0].R.NAME).width,ctx.measureText('M').width);
        ctx.fillStyle = 'white';
        ctx.fillText(state[my_team][my_int][0].R.NAME,x_offset+110, y_offset+120 )
        ctx.fillText("Attack closest enemy",x_offset+110, y_offset+130 )

        ser   = state[my_team][my_int][0].B.SERIAL.toString()
        for (let k = ser.length; k < 7; k++) {
            ser = '0' + ser
        }
        ser = 'B' + ser
        ctx.fillText(ser,x_offset+110, y_offset+160 )
        ctx.fillStyle = 'blue';
        ctx.fillRect(x_offset+110,y_offset+163,ctx.measureText(state[my_team][my_int][0].B.NAME).width,ctx.measureText('M').width);
        ctx.fillStyle = 'white';
        ctx.fillText(state[my_team][my_int][0].B.NAME,x_offset+110, y_offset+170 )
        ctx.fillText("Attack closest enemy",x_offset+110, y_offset+180 )

    }
}

function drawTable2() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {

        let x_offset = 50
        let y_offset = 170
        let unit_shift = 150
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';

        for(let i=0;i<state[my_team][my_int].length;i++){

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0,y_offset - (0 | state[my_team][my_int][i].H.HP/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.HP/2);
            ctx.fillText("HP",x_offset, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.HP,x_offset, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+30,y_offset - (0 | state[my_team][my_int][i].H.ATK/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.ATK/2);
            ctx.fillText("ATK",x_offset+30, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.ATK,x_offset+30, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+60,y_offset - (0 | state[my_team][my_int][i].H.DEF/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.DEF/2);
            ctx.fillText("DEF",x_offset+60, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.DEF,x_offset+60, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+90,y_offset - (0 | state[my_team][my_int][i].H.ACC/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.ACC/2);
            ctx.fillText("ACC",x_offset+90, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.ACC,x_offset+90, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+120,y_offset - (0 | state[my_team][my_int][i].H.CRT/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.CRT/2);
            ctx.fillText("CRT",x_offset+120, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CRT,x_offset+120, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150,y_offset - (0 | state[my_team][my_int][i].H.CD/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.CD/2);
            ctx.fillText("CD",x_offset+150, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CD,x_offset+150, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+180,y_offset - (0 | state[my_team][my_int][i].H.CLU/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.CLU/2);
            ctx.fillText("CLU",x_offset+180, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CLU,x_offset+180, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+210,y_offset - (0 | state[my_team][my_int][i].H.Weight/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].H.Weight/2);
            ctx.fillText("WGT",x_offset+210, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.Weight,x_offset+210, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].H.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'H' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.NAME,x_offset-15, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20,y_offset-110 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20,y_offset+30 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+250,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5,y_offset + (i*unit_shift) ,220,1);// draw bottom plot line

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+260,y_offset - (0 | state[my_team][my_int][i].L.HP/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.HP/2);
            ctx.fillText("HP",x_offset+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.HP,x_offset+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+30+260,y_offset - (0 | state[my_team][my_int][i].L.ATK/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.ATK/2);
            ctx.fillText("ATK",x_offset+30+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.ATK,x_offset+30+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+60+260,y_offset - (0 | state[my_team][my_int][i].L.DEF/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.DEF/2);
            ctx.fillText("DEF",x_offset+60+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.DEF,x_offset+60+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+90+260,y_offset - (0 | state[my_team][my_int][i].L.ACC/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.ACC/2);
            ctx.fillText("ACC",x_offset+90+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.ACC,x_offset+90+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+120+260,y_offset - (0 | state[my_team][my_int][i].L.CRT/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.CRT/2);
            ctx.fillText("CRT",x_offset+120+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CRT,x_offset+120+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+260,y_offset - (0 | state[my_team][my_int][i].L.CD/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.CD/2);
            ctx.fillText("CD",x_offset+150+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CD,x_offset+150+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+180+260,y_offset - (0 | state[my_team][my_int][i].L.CLU/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.CLU/2);
            ctx.fillText("CLU",x_offset+180+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CLU,x_offset+180+260, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+210+260,y_offset - (0 | state[my_team][my_int][i].L.Weight/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].L.Weight/2);
            ctx.fillText("WGT",x_offset+210+260, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.Weight,x_offset+210+260, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].L.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'L' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+260, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.NAME,x_offset-15+260, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+260, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+260,y_offset-110 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20+260,y_offset+30 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20+260,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+250+260,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+260,y_offset + (i*unit_shift) ,220,1);// draw bottom plot line

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+520,y_offset - (0 | state[my_team][my_int][i].R.HP/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.HP/2);
            ctx.fillText("HP",x_offset+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.HP,x_offset+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+30+520,y_offset - (0 | state[my_team][my_int][i].R.ATK/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.ATK/2);
            ctx.fillText("ATK",x_offset+30+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.ATK,x_offset+30+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+60+520,y_offset - (0 | state[my_team][my_int][i].R.DEF/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.DEF/2);
            ctx.fillText("DEF",x_offset+60+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.DEF,x_offset+60+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+90+520,y_offset - (0 | state[my_team][my_int][i].R.ACC/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.ACC/2);
            ctx.fillText("ACC",x_offset+90+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.ACC,x_offset+90+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+120+520,y_offset - (0 | state[my_team][my_int][i].R.CRT/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.CRT/2);
            ctx.fillText("CRT",x_offset+120+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CRT,x_offset+120+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+520,y_offset - (0 | state[my_team][my_int][i].R.CD/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.CD/2);
            ctx.fillText("CD",x_offset+150+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CD,x_offset+150+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+180+520,y_offset - (0 | state[my_team][my_int][i].R.CLU/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.CLU/2);
            ctx.fillText("CLU",x_offset+180+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CLU,x_offset+180+520, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+210+520,y_offset - (0 | state[my_team][my_int][i].R.Weight/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].R.Weight/2);
            ctx.fillText("WGT",x_offset+210+520, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.Weight,x_offset+210+520, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].R.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'R' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+520, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.NAME,x_offset-15+520, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+520, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+520,y_offset-110 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20+520,y_offset+30 + (i*unit_shift), 250, 1);
            ctx.fillRect(x_offset-20+520,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+250+520,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+520,y_offset + (i*unit_shift) ,220,1);// draw bottom plot line

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+780,y_offset - (0 | state[my_team][my_int][i].B.HP/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.HP/2);
            ctx.fillText("HP",x_offset+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.HP,x_offset+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+30+780,y_offset - (0 | state[my_team][my_int][i].B.ATK/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.ATK/2);
            ctx.fillText("ATK",x_offset+30+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.ATK,x_offset+30+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+60+780,y_offset - (0 | state[my_team][my_int][i].B.DEF/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.DEF/2);
            ctx.fillText("DEF",x_offset+60+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.DEF,x_offset+60+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+90+780,y_offset - (0 | state[my_team][my_int][i].B.ACC/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.ACC/2);
            ctx.fillText("ACC",x_offset+90+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.ACC,x_offset+90+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+120+780,y_offset - (0 | state[my_team][my_int][i].B.CRT/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.CRT/2);
            ctx.fillText("CRT",x_offset+120+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CRT,x_offset+120+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+780,y_offset - (0 | state[my_team][my_int][i].B.CD/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.CD/2);
            ctx.fillText("CD",x_offset+150+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CD,x_offset+150+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+180+780,y_offset - (0 | state[my_team][my_int][i].B.CLU/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.CLU/2);
            ctx.fillText("CLU",x_offset+180+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CLU,x_offset+180+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+210+780,y_offset - (0 | state[my_team][my_int][i].B.Weight/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.Weight/2);
            ctx.fillText("WGT",x_offset+210+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.Weight,x_offset+210+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+240+780,y_offset - (0 | state[my_team][my_int][i].B.SPD/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.SPD/2);
            ctx.fillText("SPD",x_offset+240+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.SPD,x_offset+240+780, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+270+780,y_offset - (0 | state[my_team][my_int][i].B.DOG/2) + (i*unit_shift),4,0 | state[my_team][my_int][i].B.DOG/2);
            ctx.fillText("DOG",x_offset+270+780, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.DOG,x_offset+270+780, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].B.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'B' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+780, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.NAME,x_offset-15+780, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+780, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+780,y_offset-110 + (i*unit_shift), 310, 1);
            ctx.fillRect(x_offset-20+780,y_offset+30 + (i*unit_shift), 310, 1);
            ctx.fillRect(x_offset-20+780,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+310+780,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+780,y_offset + (i*unit_shift) ,280,1);// draw bottom plot line



        }


    }
}

function drawStartersTable1() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {

        let x_offset = 50
        let y_offset = 170
        let unit_shift = 150
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';

        for(let i=0;i<state[my_team][my_int].length;i++){

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0,y_offset - (0 | state[my_team][my_int][i].H.HP/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.HP/2);
            ctx.fillText("HP",x_offset, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.HP,x_offset, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+25,y_offset - (0 | state[my_team][my_int][i].H.ATK/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.ATK/2);
            ctx.fillText("ATK",x_offset+25, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.ATK,x_offset+25, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+50,y_offset - (0 | state[my_team][my_int][i].H.DEF/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.DEF/2);
            ctx.fillText("DEF",x_offset+50, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.DEF,x_offset+50, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+75,y_offset - (0 | state[my_team][my_int][i].H.ACC/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.ACC/2);
            ctx.fillText("ACC",x_offset+75, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.ACC,x_offset+75, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+100,y_offset - (0 | state[my_team][my_int][i].H.CRT/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.CRT/2);
            ctx.fillText("CRT",x_offset+100, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CRT,x_offset+100, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+125,y_offset - (0 | state[my_team][my_int][i].H.CD/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.CD/2);
            ctx.fillText("CD",x_offset+125, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CD,x_offset+125, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150,y_offset - (0 | state[my_team][my_int][i].H.CLU/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.CLU/2);
            ctx.fillText("CLU",x_offset+150, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.CLU,x_offset+150, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+175,y_offset - (0 | state[my_team][my_int][i].H.Weight/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].H.Weight/2);
            ctx.fillText("WGT",x_offset+175, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.Weight,x_offset+175, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].H.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'H' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].H.NAME,x_offset-15, y_offset-85 + (i*unit_shift) ) // draw name
            ctx.fillText("  Basic close attack",x_offset-15, y_offset-65 + (i*unit_shift) ) // draw desc

            // draw outer box
            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20,y_offset-110 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20,y_offset+30 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+215,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5,y_offset + (i*unit_shift) ,185,1);// draw bottom plot line

            // START OF L
            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+225,y_offset - (0 | state[my_team][my_int][i].L.HP/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.HP/2);
            ctx.fillText("HP",x_offset+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.HP,x_offset+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+25+225,y_offset - (0 | state[my_team][my_int][i].L.ATK/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.ATK/2);
            ctx.fillText("ATK",x_offset+25+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.ATK,x_offset+25+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+50+225,y_offset - (0 | state[my_team][my_int][i].L.DEF/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.DEF/2);
            ctx.fillText("DEF",x_offset+50+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.DEF,x_offset+50+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+75+225,y_offset - (0 | state[my_team][my_int][i].L.ACC/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.ACC/2);
            ctx.fillText("ACC",x_offset+75+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.ACC,x_offset+75+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+100+225,y_offset - (0 | state[my_team][my_int][i].L.CRT/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.CRT/2);
            ctx.fillText("CRT",x_offset+100+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CRT,x_offset+100+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+125+225,y_offset - (0 | state[my_team][my_int][i].L.CD/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.CD/2);
            ctx.fillText("CD",x_offset+125+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CD,x_offset+125+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+225,y_offset - (0 | state[my_team][my_int][i].L.CLU/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.CLU/2);
            ctx.fillText("CLU",x_offset+150+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.CLU,x_offset+150+225, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+175+225,y_offset - (0 | state[my_team][my_int][i].L.Weight/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].L.Weight/2);
            ctx.fillText("WGT",x_offset+175+225, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.Weight,x_offset+175+225, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].L.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'L' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+225, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].L.NAME,x_offset-15+225, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+225, y_offset-65 + (i*unit_shift) )

            // draw outer box
            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+225,y_offset-110 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+225,y_offset+30 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+225,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+215+225,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+225,y_offset + (i*unit_shift) ,185,1);// draw bottom plot line

            // START OF R
            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+450,y_offset - (0 | state[my_team][my_int][i].R.HP/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.HP/2);
            ctx.fillText("HP",x_offset+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.HP,x_offset+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+25+450,y_offset - (0 | state[my_team][my_int][i].R.ATK/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.ATK/2);
            ctx.fillText("ATK",x_offset+25+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.ATK,x_offset+25+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+50+450,y_offset - (0 | state[my_team][my_int][i].R.DEF/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.DEF/2);
            ctx.fillText("DEF",x_offset+50+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.DEF,x_offset+50+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+75+450,y_offset - (0 | state[my_team][my_int][i].R.ACC/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.ACC/2);
            ctx.fillText("ACC",x_offset+75+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.ACC,x_offset+75+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+100+450,y_offset - (0 | state[my_team][my_int][i].R.CRT/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.CRT/2);
            ctx.fillText("CRT",x_offset+100+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CRT,x_offset+100+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+125+450,y_offset - (0 | state[my_team][my_int][i].R.CD/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.CD/2);
            ctx.fillText("CD",x_offset+125+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CD,x_offset+125+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+450,y_offset - (0 | state[my_team][my_int][i].R.CLU/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.CLU/2);
            ctx.fillText("CLU",x_offset+150+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.CLU,x_offset+150+450, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+175+450,y_offset - (0 | state[my_team][my_int][i].R.Weight/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].R.Weight/2);
            ctx.fillText("WGT",x_offset+175+450, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.Weight,x_offset+175+450, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].R.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'R' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+450, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].R.NAME,x_offset-15+450, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+450, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+450,y_offset-110 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+450,y_offset+30 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+450,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+215+450,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+450,y_offset + (i*unit_shift) ,185,1);// draw bottom plot line

            ctx.textAlign = "center";
            ctx.fillRect(x_offset-2+0+675,y_offset - (0 | state[my_team][my_int][i].B.HP/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.HP/2);
            ctx.fillText("HP",x_offset+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.HP,x_offset+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+25+675,y_offset - (0 | state[my_team][my_int][i].B.ATK/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.ATK/2);
            ctx.fillText("ATK",x_offset+25+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.ATK,x_offset+25+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+50+675,y_offset - (0 | state[my_team][my_int][i].B.DEF/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.DEF/2);
            ctx.fillText("DEF",x_offset+50+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.DEF,x_offset+50+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+75+675,y_offset - (0 | state[my_team][my_int][i].B.ACC/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.ACC/2);
            ctx.fillText("ACC",x_offset+75+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.ACC,x_offset+75+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+100+675,y_offset - (0 | state[my_team][my_int][i].B.CRT/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.CRT/2);
            ctx.fillText("CRT",x_offset+100+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CRT,x_offset+100+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+125+675,y_offset - (0 | state[my_team][my_int][i].B.CD/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.CD/2);
            ctx.fillText("CD",x_offset+125+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CD,x_offset+125+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+150+675,y_offset - (0 | state[my_team][my_int][i].B.CLU/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.CLU/2);
            ctx.fillText("CLU",x_offset+150+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.CLU,x_offset+150+675, y_offset+20 + (i*unit_shift) )
            ctx.fillRect(x_offset-2+175+675,y_offset - (0 | state[my_team][my_int][i].B.Weight/2) + (i*unit_shift),3,0 | state[my_team][my_int][i].B.Weight/2);
            ctx.fillText("WGT",x_offset+175+675, y_offset+10 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.Weight,x_offset+175+675, y_offset+20 + (i*unit_shift) )

            ser   = state[my_team][my_int][i].B.SERIAL.toString()
            for (let k = ser.length; k < 7; k++) {
                ser = '0' + ser
            }
            ser = 'B' + ser
            ctx.textAlign = "left";
            ctx.fillText(ser,x_offset-15+675, y_offset-95 + (i*unit_shift) )
            ctx.fillText(state[my_team][my_int][i].B.NAME,x_offset-15+675, y_offset-85 + (i*unit_shift) )
            ctx.fillText("  Basic close attack",x_offset-15+675, y_offset-65 + (i*unit_shift) )

            ctx.fillStyle = 'white';
            ctx.textAlign = "left";
            ctx.fillRect(x_offset-20+675,y_offset-110 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+675,y_offset+30 + (i*unit_shift), 215, 1);
            ctx.fillRect(x_offset-20+675,y_offset-110 + (i*unit_shift), 1, 140);
            ctx.fillRect(x_offset-20+215+675,y_offset-110 + (i*unit_shift), 1, 140);

            ctx.fillRect(x_offset-5+675,y_offset + (i*unit_shift) ,185,1);// draw bottom plot line



        }


    }
}

function drawBenchTable() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {

        let x_offset = 1000
        let y_offset = 190
        let unit_shift = 165
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';

        let key = my_team + ';' + my_int

        for (let i=0;i<bench[key].length;i++) {
            if(bench[key][i].SECTION == 3) {drawBPart(x_offset,y_offset +(i*unit_shift),bench[key][i])}
            else {drawPart(x_offset,y_offset +(i*unit_shift),bench[key][i])}
        }
    }
}

function drawStartersTable2() {
    if((state!=null) && (my_team != -1) && (my_int != -1)) {

        let x_offset = 60
        let y_offset = (c.height/2) - 250
        if(c.height%2==1){y_offset = Math.round(c.height/2) - 250}
        let unit_shift = 125
        let user_shift = 1300
        let user_count = 0;

        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';

        for (let j=0;j<(state.length);j++) {
            for (let k=0;k<(state[j].length);k++) {
                for (let i=0;i<(state[j][k].length);i++){

                    if(state[j][k][i].H.HP<=0){ctx.fillStyle='grey'}
                    ctx.textAlign = "center";
                    ctx.fillRect(x_offset-1-45+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.HP/2) + (i*unit_shift),3,0 | state[j][k][i].H.HP/2);
                    ctx.fillRect(x_offset-1-30+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.HP/2) + (i*unit_shift),3,0 | state[j][k][i].L.HP/2);
                    ctx.fillRect(x_offset-1-15+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.HP/2) + (i*unit_shift),3,0 | state[j][k][i].R.HP/2);
                    ctx.fillRect(x_offset-1-0+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.HP/2) + (i*unit_shift),3,0 | state[j][k][i].B.HP/2);
                    ctx.fillText("HP",x_offset-22.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.HP,x_offset-45+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.HP/2))
                    ctx.fillText(state[j][k][i].L.HP,x_offset-30+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.HP/2))
                    ctx.fillText(state[j][k][i].R.HP,x_offset-15+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.HP/2))
                    ctx.fillText(state[j][k][i].B.HP,x_offset-0+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.HP/2))
                    ctx.fillText("H",x_offset-45+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset-30+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset-15+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset-0+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+25+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].H.ATK/2);
                    ctx.fillRect(x_offset-1+40+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].L.ATK/2);
                    ctx.fillRect(x_offset-1+55+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].R.ATK/2);
                    ctx.fillRect(x_offset-1+70+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].B.ATK/2);
                    ctx.fillText("ATK",x_offset+47.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.ATK,x_offset+25+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.ATK/2))
                    ctx.fillText(state[j][k][i].L.ATK,x_offset+40+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.ATK/2))
                    ctx.fillText(state[j][k][i].R.ATK,x_offset+55+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.ATK/2))
                    ctx.fillText(state[j][k][i].B.ATK,x_offset+70+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.ATK/2))
                    ctx.fillText("H",x_offset+25+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+40+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+55+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+70+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+95+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].H.DEF/2);
                    ctx.fillRect(x_offset-1+110+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].L.DEF/2);
                    ctx.fillRect(x_offset-1+125+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].R.DEF/2);
                    ctx.fillRect(x_offset-1+140+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].B.DEF/2);
                    ctx.fillText("DEF",x_offset+117.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.DEF,x_offset+95+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.DEF/2))
                    ctx.fillText(state[j][k][i].L.DEF,x_offset+110+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.DEF/2))
                    ctx.fillText(state[j][k][i].R.DEF,x_offset+125+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.DEF/2))
                    ctx.fillText(state[j][k][i].B.DEF,x_offset+140+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.DEF/2))
                    ctx.fillText("H",x_offset+95+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+110+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+125+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+140+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+165+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].H.ACC/2);
                    ctx.fillRect(x_offset-1+180+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].L.ACC/2);
                    ctx.fillRect(x_offset-1+195+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].R.ACC/2);
                    ctx.fillRect(x_offset-1+210+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].B.ACC/2);
                    ctx.fillText("ACC",x_offset+187.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.ACC,x_offset+165+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.ACC/2))
                    ctx.fillText(state[j][k][i].L.ACC,x_offset+180+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.ACC/2))
                    ctx.fillText(state[j][k][i].R.ACC,x_offset+195+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.ACC/2))
                    ctx.fillText(state[j][k][i].B.ACC,x_offset+210+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.ACC/2))
                    ctx.fillText("H",x_offset+165+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+180+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+195+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+210+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+235+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.CRT/2) + (i*unit_shift),3,0 | state[j][k][i].H.CRT/2);
                    ctx.fillRect(x_offset-1+250+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.CRT/2) + (i*unit_shift),3,0 | state[j][k][i].L.CRT/2);
                    ctx.fillRect(x_offset-1+265+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.CRT/2) + (i*unit_shift),3,0 | state[j][k][i].R.CRT/2);
                    ctx.fillRect(x_offset-1+280+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.CRT/2) + (i*unit_shift),3,0 | state[j][k][i].B.CRT/2);
                    ctx.fillText("CRT",x_offset+257.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.CRT,x_offset+235+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.CRT/2))
                    ctx.fillText(state[j][k][i].L.CRT,x_offset+250+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.CRT/2))
                    ctx.fillText(state[j][k][i].R.CRT,x_offset+265+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.CRT/2))
                    ctx.fillText(state[j][k][i].B.CRT,x_offset+280+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.CRT/2))
                    ctx.fillText("H",x_offset+235+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+250+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+265+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+280+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+305+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.CD/2) + (i*unit_shift),3,0 | state[j][k][i].H.CD/2);
                    ctx.fillRect(x_offset-1+320+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.CD/2) + (i*unit_shift),3,0 | state[j][k][i].L.CD/2);
                    ctx.fillRect(x_offset-1+335+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.CD/2) + (i*unit_shift),3,0 | state[j][k][i].R.CD/2);
                    ctx.fillRect(x_offset-1+350+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.CD/2) + (i*unit_shift),3,0 | state[j][k][i].B.CD/2);
                    ctx.fillText("CD",x_offset+327.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.CD,x_offset+305+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.CD/2))
                    ctx.fillText(state[j][k][i].L.CD,x_offset+320+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.CD/2))
                    ctx.fillText(state[j][k][i].R.CD,x_offset+335+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.CD/2))
                    ctx.fillText(state[j][k][i].B.CD,x_offset+350+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.CD/2))
                    ctx.fillText("H",x_offset+305+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+320+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+335+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+350+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+375+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.CLU/2) + (i*unit_shift),3, 0 | state[j][k][i].H.CLU/2);
                    ctx.fillRect(x_offset-1+390+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.CLU/2) + (i*unit_shift),3,0 | state[j][k][i].L.CLU/2);
                    ctx.fillRect(x_offset-1+405+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.CLU/2) + (i*unit_shift),3,0 | state[j][k][i].R.CLU/2);
                    ctx.fillRect(x_offset-1+420+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.CLU/2) + (i*unit_shift),3,0 | state[j][k][i].B.CLU/2);
                    ctx.fillText("CLU",x_offset+397.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].H.CLU,x_offset+375+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.CLU/2) )
                    ctx.fillText(state[j][k][i].L.CLU,x_offset+390+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.CLU/2) )
                    ctx.fillText(state[j][k][i].R.CLU,x_offset+405+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.CLU/2) )
                    ctx.fillText(state[j][k][i].B.CLU,x_offset+420+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.CLU/2) )
                    ctx.fillText("H",x_offset+375+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+390+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+405+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+420+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+445+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.Weight/2) + (i*unit_shift),3,0 | state[j][k][i].H.Weight/2);
                    ctx.fillRect(x_offset-1+460+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.Weight/2) + (i*unit_shift),3,0 | state[j][k][i].L.Weight/2);
                    ctx.fillRect(x_offset-1+475+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.Weight/2) + (i*unit_shift),3,0 | state[j][k][i].R.Weight/2);
                    ctx.fillRect(x_offset-1+490+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.Weight/2) + (i*unit_shift),3,0 | state[j][k][i].B.Weight/2);
                    ctx.fillText("WGT",x_offset+467.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText("H",x_offset+445+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("L",x_offset+460+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("R",x_offset+475+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.fillText("B",x_offset+490+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    let total_prob = 0
                    if(state[j][k][i].H.HP>0){total_prob=total_prob+state[j][k][i].H.Weight}
                    if(state[j][k][i].L.HP>0){total_prob=total_prob+state[j][k][i].L.Weight}
                    if(state[j][k][i].R.HP>0){total_prob=total_prob+state[j][k][i].R.Weight}
                    if(state[j][k][i].B.HP>0){total_prob=total_prob+state[j][k][i].B.Weight}

                    ctx.font = '8px monospace';
                    if(state[j][k][i].H.HP>0) {
                        ctx.fillText('('+(0|(100*state[j][k][i].H.Weight/total_prob))+"%)", x_offset+445+(user_shift*user_count),  y_offset-15 + (i*unit_shift) - (0 | state[j][k][i].H.Weight/2) );
                        ctx.fillText(state[j][k][i].H.Weight,x_offset+445+(user_shift*user_count), y_offset-5 + (i*unit_shift) - (0 | state[j][k][i].H.Weight/2) )
                    } else {
                        ctx.fillText(state[j][k][i].H.Weight,x_offset+445+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.Weight/2) )
                    }
                    if (state[j][k][i].L.HP > 0) {
                        ctx.fillText('('+(0|(100*state[j][k][i].L.Weight/total_prob))+"%)", x_offset+460+(user_shift*user_count),  y_offset-15 + (i*unit_shift) - (0 | state[j][k][i].L.Weight/2) );
                        ctx.fillText(state[j][k][i].L.Weight,x_offset+460+(user_shift*user_count), y_offset-5 + (i*unit_shift) - (0 | state[j][k][i].L.Weight/2) )
                    } else {
                        ctx.fillText(state[j][k][i].L.Weight,x_offset+445+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.Weight/2) )
                    }
                    if (state[j][k][i].R.HP > 0) {
                        ctx.fillText('('+(0|(100*state[j][k][i].R.Weight/total_prob))+"%)", x_offset+475+(user_shift*user_count),  y_offset-15 + (i*unit_shift) - (0 | state[j][k][i].R.Weight/2) );
                        ctx.fillText(state[j][k][i].R.Weight,x_offset+475+(user_shift*user_count), y_offset-5 + (i*unit_shift) - (0 | state[j][k][i].R.Weight/2) )
                    } else {
                        ctx.fillText(state[j][k][i].R.Weight,x_offset+475+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.Weight/2) )
                    }
                    if (state[j][k][i].B.HP > 0) {
                        ctx.fillText('('+(0|(100*state[j][k][i].B.Weight/total_prob))+"%)", x_offset+490+(user_shift*user_count),  y_offset-15 + (i*unit_shift) - (0 | state[j][k][i].B.Weight/2) );
                        ctx.fillText(state[j][k][i].B.Weight,x_offset+490+(user_shift*user_count), y_offset-5 + (i*unit_shift) - (0 | state[j][k][i].B.Weight/2) )
                    } else {
                        ctx.fillText(state[j][k][i].B.Weight,x_offset+490+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.Weight/2) )
                    }
                    ctx.font = '11px monospace';


                    ctx.fillRect(x_offset-1+515+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.SPD/2) + (i*unit_shift),3,0 | state[j][k][i].B.SPD/2);
                    ctx.fillText("SPD",x_offset+515+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].B.SPD,x_offset+515+(user_shift*user_count), y_offset - 10 + (i*unit_shift) - (0 | state[j][k][i].B.SPD/2) )
                    ctx.fillText("B",x_offset+515+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+540+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.DOG/2) + (i*unit_shift),3,0 | state[j][k][i].B.DOG/2);
                    ctx.fillText("DOG",x_offset+540+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].B.DOG,x_offset+540+(user_shift*user_count), y_offset - 10 + (i*unit_shift) - (0 | state[j][k][i].B.DOG/2) )
                    ctx.fillText("B",x_offset+540+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    /*
                    // write part desc
                    ctx.font = '11px monospace';
                    ser   = state[my_team][my_int][0].B.SERIAL.toString()
                    for (let k = ser.length; k < 7; k++) {
                        ser = '0' + ser
                    }
                    ser = 'B' + ser
                    ctx.textAlign = "left";
                    ctx.fillText(ser,x_offset-15, y_offset-95+ (i*unit_shift) )
                    ctx.fillText(state[my_team][my_int][0].H.NAME,x_offset-15, y_offset-85+ (i*unit_shift) )
                    */

                    ctx.fillRect(x_offset-50+(user_shift*user_count),y_offset + (i*unit_shift) ,600,1);
                    ctx.fillStyle = 'white';
                }
                user_count = user_count + 1
            }
        }



        ctx.fillStyle = 'black';

    }
}

function drawCircle(center_x,center_y,radius){

    ctx.beginPath();
    ctx.strokeStyle = 'pink';
    ctx.arc(center_x, center_y, radius, 0, 2 * Math.PI);
    ctx.stroke();


}

function drawPoint(center_x,center_y,radius,num){

    if(num==0){return}

    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length;k++){

                if(state[i][j][k].H.HP>0) {

                    let p = 16*Math.PI/10
                    if(i==1){
                        p = (16*Math.PI/10) + (2*k*Math.PI/10)
                    }
                    else {
                        p = (14*Math.PI/10) - (2*k*Math.PI/10)
                    }

                    let x = center_x + radius * Math.cos(p) * (100-state[i][j][k].Position)/100;
                    let y = center_y + radius * Math.sin(p) * (100-state[i][j][k].Position)/100;

                    ctx.beginPath();
                    if(i==my_team && j==my_int){ctx.fillStyle = 'blue';}
                    else{ctx.fillStyle = 'green';}

                    ctx.arc(x, y, 2, 0, 2 * Math.PI);
                    ctx.fill();

                    ctx.fillStyle = 'white';
                    ctx.textAlign = "center";
                    ctx.font = '8px monospace';
                    ctx.fillText(k, x, y-10 );
                    ctx.fillText(state[i][j][k].Position.toFixed(1), x, y+20 );
                    //ctx.font = '11px monospace';
                }
            }
        }
    }
}

function wrap(ctx, text, fontSize, fontColor) {
    var max_width  = 250;
    var fontSize   =  12;
    var lines      =  new Array();
    var width = 0, i, j;
    var result;
    var color = fontColor || "white";

    // Font and size is required for ctx.measureText()
    ctx.font   = fontSize + "px Arial";


    // Start calculation
    while ( text.length ) {
        for( i=text.length; ctx.measureText(text.substr(0,i)).width > max_width; i-- );

        result = text.substr(0,i);

        if ( i !== text.length )
            for( j=0; result.indexOf(" ",j) !== -1; j=result.indexOf(" ",j)+1 );

        lines.push( result.substr(0, j|| result.length) );
        width = Math.max( width, ctx.measureText(lines[ lines.length-1 ]).width );
        text  = text.substr( lines[ lines.length-1 ].length, text.length );
    }


    // Calculate canvas size, add margin
    ctx.canvas.width  = 14 + width;
    ctx.canvas.height =  8 + ( fontSize + 5 ) * lines.length;
    ctx.font   = fontSize + "px Arial";

    // Render
    ctx.fillStyle = color;
    for ( i=0, j=lines.length; i<j; ++i ) {
        ctx.fillText( lines[i], 8, 5 + fontSize + (fontSize+5) * i );
    }
}

function drawPart(x,y,part) {
    ctx.textAlign = "center";
    ctx.fillRect(x-2+0,y - (0 | part.HP/2),3,0 | part.HP/2);
    ctx.fillText("HP",x, y+15 )
    ctx.fillText(part.HP,x, y-10 - (0 | part.HP/2) )
    ctx.fillRect(x-2+25,y - (0 | part.ATK/2),3,0 | part.ATK/2);
    ctx.fillText("ATK",x+25, y+15 )
    ctx.fillText(part.ATK,x+25, y-10 - (0 | part.ATK/2) )
    ctx.fillRect(x-2+50,y - (0 | part.DEF/2),3,0 | part.DEF/2);
    ctx.fillText("DEF",x+50, y+15 )
    ctx.fillText(part.DEF,x+50, y-10 - (0 | part.DEF/2) )
    ctx.fillRect(x-2+75,y - (0 | part.ACC/2),3,0 | part.ACC/2);
    ctx.fillText("ACC",x+75, y+15 )
    ctx.fillText(part.ACC,x+75, y-10 - (0 | part.ACC/2) )
    ctx.fillRect(x-2+100,y - (0 | part.CRT/2),3,0 | part.CRT/2);
    ctx.fillText("CRT",x+100, y+15 )
    ctx.fillText(part.CRT,x+100, y-10 - (0 | part.CRT/2) )
    ctx.fillRect(x-2+125,y - (0 | part.CD/2),3,0 | part.CD/2);
    ctx.fillText("CD",x+125, y+15 )
    ctx.fillText(part.CD,x+125, y-10 - (0 | part.CD/2) )
    ctx.fillRect(x-2+150,y - (0 | part.CLU/2),3,0 | part.CLU/2);
    ctx.fillText("CLU",x+150, y+15 )
    ctx.fillText(part.CLU,x+150, y-10 - (0 | part.CLU/2) )
    ctx.fillRect(x-2+175, y - (0 | part.Weight/2),3,0 | part.Weight/2);
    ctx.fillText("WGT",x+175, y+15 )
    ctx.fillText(part.Weight,x+175, y-10 - (0 | part.Weight/2) )

    ser   = part.SERIAL.toString()
    for (let k = ser.length; k < 7; k++) {
        ser = '0' + ser
    }
    if(part.SECTION == 0){
        ser = 'H' + ser
    } else if(part.SECTION == 1){
        ser = 'L' + ser
    } else if(part.SECTION == 2){
        ser = 'R' + ser
    } else {
        ser = 'B' + ser
    }

    ctx.textAlign = "left";
    ctx.fillText(ser,x-15, y-115 )
    ctx.fillText(part.NAME,x-15, y-105 ) // draw name
    ctx.fillText("  Basic close attack",x-15, y-85 ) // draw desc

    // draw outer box
    ctx.fillStyle = 'white';
    ctx.textAlign = "left";
    ctx.fillRect(x-20,y-130, 215, 1);
    ctx.fillRect(x-20,y+25, 215, 1);
    ctx.fillRect(x-20,y-130, 1, 155);
    ctx.fillRect(x-20+215,y-130, 1, 155);

    ctx.fillRect(x-5,y ,185,1);// draw bottom plot line
}

function drawBPart(x,y,part) {
    ctx.textAlign = "center";
    ctx.fillRect(x-2+0,y - (0 | part.HP/2),3,0 | part.HP/2);
    ctx.fillText("HP",x, y+15 )
    ctx.fillText(part.HP,x, y -10 - (0 | part.HP/2) )
    ctx.fillRect(x-2+25,y - (0 | part.ATK/2),3,0 | part.ATK/2);
    ctx.fillText("ATK",x+25, y+15 )
    ctx.fillText(part.ATK,x+25, y-10 - (0 | part.ATK/2) )
    ctx.fillRect(x-2+50,y - (0 | part.DEF/2),3,0 | part.DEF/2);
    ctx.fillText("DEF",x+50, y+15 )
    ctx.fillText(part.DEF,x+50, y-10 - (0 | part.DEF/2) )
    ctx.fillRect(x-2+75,y - (0 | part.ACC/2),3,0 | part.ACC/2);
    ctx.fillText("ACC",x+75, y+15 )
    ctx.fillText(part.ACC,x+75, y-10 - (0 | part.ACC/2) )
    ctx.fillRect(x-2+100,y - (0 | part.CRT/2),3,0 | part.CRT/2);
    ctx.fillText("CRT",x+100, y+15 )
    ctx.fillText(part.CRT,x+100, y-10 - (0 | part.CRT/2) )
    ctx.fillRect(x-2+125,y - (0 | part.CD/2),3,0 | part.CD/2);
    ctx.fillText("CD",x+125, y+15 )
    ctx.fillText(part.CD,x+125, y-10 - (0 | part.CD/2) )
    ctx.fillRect(x-2+150,y - (0 | part.CLU/2),3,0 | part.CLU/2);
    ctx.fillText("CLU",x+150, y+15 )
    ctx.fillText(part.CLU,x+150, y-10 - (0 | part.CLU/2) )
    ctx.fillRect(x-2+175, y - (0 | part.Weight/2),3,0 | part.Weight/2);
    ctx.fillText("WGT",x+175, y+15 )
    ctx.fillText(part.Weight,x+175, y-10 - (0 | part.Weight/2) )
    ctx.fillRect(x-2+200, y - (0 | part.SPD/2),3,0 | part.SPD/2);
    ctx.fillText("SPD",x+200, y+15 )
    ctx.fillText(part.SPD,x+200, y-10 - (0 | part.SPD/2) )
    ctx.fillRect(x-2+225, y - (0 | part.DOG/2),3,0 | part.DOG/2);
    ctx.fillText("DOG",x+225, y+15 )
    ctx.fillText(part.DOG,x+225, y-10 - (0 | part.DOG/2) )

    ser   = part.SERIAL.toString()
    for (let k = ser.length; k < 7; k++) {
        ser = '0' + ser
    }
    if(part.SECTION == 0){
        ser = 'H' + ser
    } else if(part.SECTION == 1){
        ser = 'L' + ser
    } else if(part.SECTION == 2){
        ser = 'R' + ser
    } else {
        ser = 'B' + ser
    }

    ctx.textAlign = "left";
    ctx.fillText(ser,x-15, y-115 )
    ctx.fillText(part.NAME,x-15, y-105 ) // draw name
    ctx.fillText("  Basic close attack",x-15, y-85 ) // draw desc

    // draw outer box
    ctx.fillStyle = 'white';
    ctx.textAlign = "left";
    ctx.fillRect(x-20,y-130, 265, 1);
    ctx.fillRect(x-20,y+25, 265, 1);
    ctx.fillRect(x-20,y-130, 1, 155);
    ctx.fillRect(x-20+265,y-130, 1, 155);

    ctx.fillRect(x-5,y ,235,1);// draw bottom plot line
}

function drawImage(){
    img1 = new Image();
    img1.src = "/images/1.png"
    ctx.drawImage(img1, 10,10, img1.width*.3, img1.height*.3);
    img2 = new Image();
    img2.src = "/images/2.png"
    ctx.drawImage(img2, 50,50, img2.width*.3, img2.height*.3);
    img3 = new Image();
    img3.src = "/images/3.png"
    ctx.drawImage(img3, 100,100, img3.width*.3, img3.height*.3);
    img4 = new Image();
    img4.src = "/images/4.png"
    ctx.drawImage(img4, 150,150, img4.width*.3, img4.height*.3);
    img5 = new Image();
    img5.src = "/images/5.png"
    ctx.drawImage(img5, 200,200, img5.width*.3, img5.height*.3);
    img6 = new Image();
    img6.src = "/images/6.png"
    ctx.drawImage(img6, 250,250, img6.width*.3, img6.height*.3);
    img7 = new Image();
    img7.src = "/images/7.png"
    ctx.drawImage(img7, 300,300, img7.width*.3, img7.height*.3);
    img8 = new Image();
    img8.src = "/images/8.png"
    ctx.drawImage(img8, 350,350, img8.width*.3, img8.height*.3);
}

function drawBoard() {

    let w = Math.round(c.width/2)-250
    let h = Math.round(c.height/2)-250
    ctx.textAlign = "left";
    ctx.fillStyle = 'grey';

    ctx.fillRect(w,h ,500,1);
    ctx.fillRect(w,h+500 ,500,1);
    ctx.fillRect(w,h,1,500);
    ctx.fillRect(w+500,h ,1,500);

    ctx.fillRect(Math.round(c.width/2),h, 1, 500);
}

function drawSinglePoint(shift_x,shift_y,center_x,center_y, color){
    //console.log(center_x,center_y)
    ctx.beginPath();
    ctx.fillStyle = color;

    ctx.arc(center_x+shift_x, center_y+shift_y, 3, 0, 2 * Math.PI);
    ctx.fill();
}

function calculateCenter(center_x,center_y) {
    let sum_x = 0;
    let sum_y = 0;

    count = 0
    for(let i = 0; i < state.length;i++){
        for(let j = 0; j < state[i].length;j++) {
            for(let k = 0; k < state[i][j].length;k++) {

                if(state[i][j][k].H.HP > 0) {

                    let p = 16*Math.PI/10
                    if(i==1){
                        p = (16*Math.PI/10) + (2*k*Math.PI/10)
                    }
                    else {
                        p = (14*Math.PI/10) - (2*k*Math.PI/10)
                    }

                    let x = Math.cos(p) * (100-state[i][j][k].Position)/100;
                    let y = Math.sin(p) * (100-state[i][j][k].Position)/100;

                    sum_x = sum_x + x
                    sum_y = sum_y + y
                    count = count + 1
                }

            }
        }
    }

    return [ 250*sum_x/count, 250*sum_y/count ]
}

function drawHeatMap(){

    let x_offset = 0
    let y_offset = 0

    let x_shift = 0;
    let y_shift = 0;
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {

                ctx.fillStyle = 'blue';
                ctx.fillRect(x_offset+x_shift,y_offset+y_shift,10,10)
                ctx.font = '5px monospace';
                ctx.fillStyle = 'white';
                ctx.fillText(state[i][j][k].H.HP,x_offset+x_shift,y_offset+y_shift);
                x_shift = x_shift+10;

                ctx.textAlign = "left";
                ctx.fillStyle = 'blue';
                ctx.fillRect(x_offset+x_shift,y_offset+y_shift,10,10)
                ctx.font = '5px monospace';
                ctx.fillStyle = 'white';
                ctx.fillText(state[i][j][k].H.DEF,x_offset+x_shift,y_offset+y_shift);
                x_shift = x_shift+10;


                y_shift=y_shift+10;
                x_shift=0;
            }
            x_shift = 0;
            y_shift=y_shift+10;
        }
    }
}