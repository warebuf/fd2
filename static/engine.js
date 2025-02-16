console.log("loaded js");

class AStarNode {
    constructor(Origin, x, y, g = 0, h = 0, parent = null) {

        this.X = x;
        this.Y = y;
        this.G = g; // cost from start
        this.H = h; // Heuristic cost
        this.F = g + h; // Total cost
        this.parent = parent;
        this.origin = Origin;
    }
}



var c = document.getElementById("myCanvas");
const ctx = c.getContext("2d");

//c.width = window.innerWidth;
//c.height = window.innerHeight;

function getMousePosition(event) {
    const rect = c.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    //console.log(`mouse pos x ${x}, y ${y}`)
    console.log(`coord pos x ${x-401}, y ${732-y}`)

}
//c.addEventListener("mousemove", getMousePosition);

c.width = document.body.clientWidth;
c.height = document.body.clientHeight;

const dimension = 500;
const width = Math.round(c.width/2)-(dimension/2);
const height = Math.round(c.height/2)+(dimension/2);

console.log("width, height", width, height);

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


let last = performance.now();
let frameTimes = [];
let fps = 0;

function anime() {

    const current = performance.now();
    frameTimes.push(current - last)
    if(frameTimes.length > 30) frameTimes.shift();
    fps = Math.round(1000 / (frameTimes.reduce((a, b) => a + b) / frameTimes.length));
    last = current;

    // Update FPS display
    document.getElementById("fps").textContent = `FPS: ${fps}`;

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
        //drawGrid()
        drawPhase()
        drawLog()
        //drawBench()
        //drawStarters()
        drawError()
        //drawTable1()
        //drawTable2()
        //drawStartersTable1()
        //drawBenchTable()
        //drawStartersTable2()
        //drawSelect()
        //ctx.fillRect(1150,40 , 1, 800);

    } else if (match_data.length > 0) {

        drawObstacles()
        drawCampCrosshair()
        calcPos()
        drawAttack()
        drawPos()
        drawHoverable()
        drawCamps()
        //drawPotentialPaths()
        //drawPaths()
        //drawStats()
        drawDamageNumbers()
        updateDamageNumbers()
        drawTurn()
        drawUnitsOfTime()
        drawLog()
        //drawSprite(); // I think it's too distracting right now
        drawSelected();

        if(view==0){
            drawHeatMap()
        }


        if( testPath.length > 0 ) {
            drawComplexPath(testPath, "blue");
        }

        if( paths.length > 0 ) {
            let selected = -1;
            for(let i = 0; i < paths.length; i++) {
                if(
                    (paths[i][paths[i].length-1][0] == hoverable.X) &&
                    (paths[i][paths[i].length-1][1] == hoverable.Y)
                ) {
                    selected = i;
                } else {
                    drawComplexPath(paths[i], "grey");
                }
            }
            if(selected >= 0) {drawComplexPath(paths[selected], "#FFC107");}
        }
    }

    frame_count += 1
    if(frame_count==15) {
        frame_count = 0;
        if(even==0) {even = 1}
        else {even=0}
    }

}
anime()

function calcPos() {
    if( (animating_state < match_data.length) && (attacking_frames_count <= 0) ) {

        let all_status_correct = true;
        for(let i = 0; i < state.length; i++) {
            for(let j = 0; j < state[i].length; j++) {
                for(let k = 0; k < state[i][j].length; k++) {
                    if(state[i][j][k].Status != match_data[animating_state][i][j][k].Status) {
                        all_status_correct = false;
                        if(state[i][j][k].Status == 0) {
                            state[i][j][k].Status = 1;
                            state[i][j][k].Start = match_data[animating_state][i][j][k].Start;
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft;
                            state[i][j][k].Path = match_data[animating_state][i][j][k].Path;
                            state[i][j][k].Path_Segment = 0;
                            state[i][j][k].Thetas = match_data[animating_state][i][j][k].Path;
                        } else if(state[i][j][k].Status == 1) {
                            if(state[i][j][k].DistanceLeft <= 0) {
                                if(camps[state[i][j][k].Destination.I].HP > 0) {
                                    state[i][j][k].Status = 2;
                                    state[i][j][k].Battle = camps[state[i][j][k].Destination.I]
                                    camps[state[i][j][k].Destination.I].Units.push([i,j,k])
                                } else {
                                    state[i][j][k].Status = 0;
                                }
                            }
                        } else if(state[i][j][k].Status == 2) {
                            if((state[i][j][k].Battle != null) && (state[i][j][k].Battle.HP <= 0)) {
                                state[i][j][k].Status = 0;
                                state[i][j][k].TTD = 100;
                                state[i][j][k].Battle.Units = []
                                state[i][j][k].Battle = null;
                            }
                        }
                    }
                }
            }
        }

        let all_distances_correct = true;
        for(let i = 0; i < state.length; i++) {
            for(let j = 0; j < state[i].length; j++) {
                for(let k = 0; k < state[i][j].length; k++) {
                    if(state[i][j][k].Status == 1) {
                        if(campEqual(state[i][j][k].Destination, match_data[animating_state][i][j][k].Destination) == false) {
                            state[i][j][k].Start = match_data[animating_state][i][j][k].Start;
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft;
                            state[i][j][k].Path = match_data[animating_state][i][j][k].Path;
                            state[i][j][k].Path_Segment = 0;
                        }
                        if(state[i][j][k].DistanceLeft != match_data[animating_state][i][j][k].DistanceLeft) {
                            all_distances_correct = false;

                            let toMove = state[i][j][k].SPD * 0.015;
                            state[i][j][k].DistanceLeft -= toMove;
                            while (toMove > 0) {
                                if(state[i][j][k].Path_Segment + 1 >= state[i][j][k].Path.length) {
                                    toMove = 0;
                                    state[i][j][k].DistanceLeft = 0;
                                } else {
                                    let seg_x = state[i][j][k].Path[state[i][j][k].Path_Segment+1][0] - state[i][j][k].Position.X
                                    let seg_y = state[i][j][k].Path[state[i][j][k].Path_Segment+1][1] - state[i][j][k].Position.Y
                                    let segment_distance = Math.sqrt((seg_x*seg_x) + (seg_y*seg_y))
                                    toMove -= segment_distance
                                    if (toMove >= 0) {
                                        state[i][j][k].Position.X = state[i][j][k].Path[state[i][j][k].Path_Segment+1][0]
                                        state[i][j][k].Position.Y = state[i][j][k].Path[state[i][j][k].Path_Segment+1][1]
                                        state[i][j][k].Path_Segment++
                                    } else { // the negative toMove distance left is the distance left in the segment
                                        let percentLeft = -toMove / segment_distance
                                        state[i][j][k].Position.X = state[i][j][k].Path[state[i][j][k].Path_Segment+1][0] - (seg_x*percentLeft)
                                        state[i][j][k].Position.Y = state[i][j][k].Path[state[i][j][k].Path_Segment+1][1] - (seg_y*percentLeft)
                                    }
                                }
                            }

                            if( state[i][j][k].DistanceLeft - 0.01 <= match_data[animating_state][i][j][k].DistanceLeft ){
                                state[i][j][k].Position = structuredClone(match_data[animating_state][i][j][k].Position)
                                state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft
                                state[i][j][k].Path_Segment = match_data[animating_state][i][j][k].Path_Segment
                            }

                        }
                    } else if(state[i][j][k].Status == 2) {
                        if(state[i][j][k].TTD != match_data[animating_state][i][j][k].TTD) {
                            all_distances_correct = false;
                            state[i][j][k].TTD -= (0.015*state[i][j][k].AS)
                            if(state[i][j][k].TTD <= match_data[animating_state][i][j][k].TTD) {
                                state[i][j][k].TTD = match_data[animating_state][i][j][k].TTD
                            }
                            if(state[i][j][k].TTD==0){
                                attacking_frames_count=1
                            }
                        }
                    }
                }
            }
        }
        followHoverable()

        if(all_status_correct && all_distances_correct) {
            if((time_queue[animating_state]!="") && (attacking_frames_count == 0)) { // update time stamp
                total_units_of_time += parseFloat(time_queue[animating_state])
                total_units_of_time = parseFloat(total_units_of_time.toFixed(3))
                event_log.push(time_queue[animating_state].toString() + " ticks have passed")
                if(event_log.length >= 25) {
                    event_log.shift()
                }
            }
            state = JSON.parse(JSON.stringify(match_data[animating_state]))
            camps = camp_data[animating_state]
            turn = turn_queue[animating_state]
            animating_state++
            setupHoverable()
        }
    }
}

function drawAttack() {

    if( (animating_state < match_data.length) && (attacking_frames_count > 0) ) {

        for(let i=0; i < atk_data[animating_state].length; i++) {
            let atk_obj = JSON.parse(atk_data[animating_state][i]);
            let atki = atk_obj.Attacker[0]
            let atkj = atk_obj.Attacker[1]
            let atkk = atk_obj.Attacker[2]

            let campOfInterest = state[atki][atkj][atkk].Battle
            campOfInterest.HP -= state[atki][atkj][atkk].ATK

            if( attacking_frames_count == 1) {
                createDamageNumber(state[atki][atkj][atkk].ATK, width+campOfInterest.XY.X, height-campOfInterest.XY.Y)
                event_log.push(atk_data[animating_state][i])
                if(event_log.length >= 25) {
                    event_log.shift()
                }
            }
        }

        attacking_frames_count++
        if(attacking_frames_count>=1) {attacking_frames_count = 0}
    }
}

function drawPos(){

    ctx.fillStyle = "white";

    ctx.textAlign = "left";
    ctx.fillStyle = 'white';
    ctx.font = '11px monospace';
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                if(state[i][j][k].HP > 0) {
                    //drawHPBoxes(state[i][j][k].HP,state[i][j][k].HP,width+state[i][j][k].Position.X,height-state[i][j][k].Position.Y)
                    drawRotated('>',width+state[i][j][k].Position.X,height-state[i][j][k].Position.Y,state[i][j][k].Thetas, state[i][j][k].Path_Segment,"11", 3, "white")
                }
            }
        }
    }
}

function drawCamps() {

    ctx.fillStyle = "green";

    for(let i = 0; i < camps.length; i++) {
        ctx.fillRect(width+camps[i].XY.X,height-camps[i].XY.Y,1,1);

        let x_shift_tracker = []
        x_shift_tracker.push([]) // list of my_team units at this camp
        x_shift_tracker.push([]) // list of enemy units at this camp
        for(let j = 0; j < camps[i].Units.length; j++) {
            if((camps[i].Units[j][0]==my_team) && (camps[i].Units[j][1]==my_int)) {x_shift_tracker[0].push(0)}
            else {x_shift_tracker[1].push(0)}
        }

        // this draws the attack circles
        for(let j = 0; j < camps[i].Units.length; j++) {
            let unitOfInterest = state[camps[i].Units[j][0]][camps[i].Units[j][1]][camps[i].Units[j][2]]
            if(unitOfInterest.Status == 2) {

                let x_shift = 0;
                let y_shift = 15;
                let odd = x_shift_tracker[0].length%2;
                if(x_shift_tracker[0].length==1){odd=0;}
                if((camps[i].Units[j][0]==my_team) && (camps[i].Units[j][1]==my_int)) {
                    y_shift = -15;
                    x_shift = -Math.floor(x_shift_tracker[0].length/2)*10-(odd*10)
                    x_shift += x_shift_tracker[0][0]*20
                    x_shift_tracker[0][0] += 1
                } else {
                    odd = x_shift_tracker[1].length%2;
                    if(x_shift_tracker[1].length==1){odd=0;}
                    x_shift = -Math.floor(x_shift_tracker[1].length/2)*10-(odd*10)
                    x_shift += x_shift_tracker[1][0]*20
                    x_shift_tracker[1][0] += 1
                }

                // draw percentage circle
                let percentage = (100-unitOfInterest.TTD)/100

                // Draw the progress circle
                ctx.beginPath();
                var startAngle = -Math.PI / 2; // Start at the top
                var endAngle = percentage * 2 * Math.PI + startAngle;
                ctx.arc(width+(unitOfInterest.Position.X) + x_shift, height - unitOfInterest.Position.Y + y_shift, 5, startAngle, endAngle, false);

                if((percentage != 0) && (percentage != 1)) { // Close the path to form a wedge
                    ctx.lineTo(width + unitOfInterest.Position.X + x_shift, height - unitOfInterest.Position.Y + y_shift); // Draw line to center of the circle
                    ctx.closePath(); // Automatically draws a line from the last point to the starting point
                }
                if((camps[i].Units[j][0]==my_team) && (camps[i].Units[j][1]==my_int)) {ctx.strokeStyle = 'blue';}
                else {ctx.strokeStyle = 'green';}
                ctx.stroke();
            }
        }

        // draw HP
        if(camps[i].HP > 0) {
            drawHPBoxes(camps[i].HP,camps[i].MaxHP,width+camps[i].XY.X,height-camps[i].XY.Y)
        }
    }

    ctx.fillStyle = "white";
}

function drawCampCrosshair() {
    if(state_machine == 1) {
        if(even) {

            ctx.fillStyle = "white";
            for(let i = 0; i < camps.length; i++) {
                ctx.fillRect(width+camps[i].XY.X-5,height-camps[i].XY.Y+5,1,1);
                ctx.fillRect(width+camps[i].XY.X-5,height-camps[i].XY.Y-5,1,1);
                ctx.fillRect(width+camps[i].XY.X+5,height-camps[i].XY.Y+5,1,1);
                ctx.fillRect(width+camps[i].XY.X+5,height-camps[i].XY.Y-5,1,1);
            }
        }
    }
}

function drawHoverable() {
    if(hoverable!=null){
        ctx.fillStyle = 'pink'
        ctx.fillRect(width+hoverable.X-10,height-hoverable.Y+10,20,1);
        ctx.fillRect(width+hoverable.X-10,height-hoverable.Y-10,20,1);
        ctx.fillRect(width+hoverable.X-10,height-hoverable.Y-10,1,20);
        ctx.fillRect(width+hoverable.X+10,height-hoverable.Y-10,1,20);
        ctx.fillStyle = 'white'
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
    time_string = total_units_of_time+" ticks"
    ctx.fillText(time_string,c.width/2, 35);
}

function drawLog() {

    let x = 50
    let y = 50

    if(view==1) {
        x = Math.round(c.width/2) - 250
        y = Math.round(c.height/2) + 300
    }

    for(i =0; i<event_log.length;i++) {
        ctx.textAlign = "left";
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(event_log[i],x, y + (10*i));
    }
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

function drawHeatMap(){

    let x_offset = Math.round(c.width/2) - 400
    let y_offset = Math.round(c.height/2) - 250

    let x_shift = 0;
    let y_shift = 0;

    let t_shift = 0;

    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {

                if ( state[i][j][k].HP > 0 ){
                    if(state[i][j][k].Status == 0){
                        ctx.beginPath();
                        ctx.strokeStyle = 'green';
                        ctx.arc(x_offset+x_shift+t_shift-10, y_offset+y_shift-2, 3, 0, 2 * Math.PI); // (x, y, radius, startAngle, endAngle)
                        ctx.stroke();
                    }
                    else {
                        ctx.beginPath();
                        ctx.strokeStyle = 'white';
                        ctx.arc(x_offset+x_shift+t_shift-10, y_offset+y_shift-2, 3, 0, 2 * Math.PI); // (x, y, radius, startAngle, endAngle)
                        ctx.stroke();
                    }
                }

                ctx.textAlign = "left";
                ctx.font = '8px monospace';
                ctx.fillStyle = 'white'
                if(state_machine==0) {
                    if(hoverable.Type == "unit") {
                        if((hoverable.I == i) && (hoverable.J==j) && (hoverable.K==k)) {
                            ctx.fillStyle = 'red'
                        }
                    }
                } else if(state_machine==1) {
                    if(selected_hoverable.Type == "unit") {
                        if((selected_hoverable.I == i) && (selected_hoverable.J==j) && (selected_hoverable.K==k)) {
                            ctx.fillStyle = 'red'
                        }
                    }
                }
                ctx.fillText(state[i][j][k].Name,x_offset+x_shift+t_shift,y_offset+y_shift);
                x_shift += 24*4
                ctx.textAlign = "right";
                ctx.fillText("LV. " + (Math.floor(state[i][j][k].EXP/100)+1),x_offset+x_shift+t_shift,y_offset+y_shift);
                x_shift = 0;
                y_shift = y_shift+5

                ctx.textAlign = "center";
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.fillStyle = 'white';
                ctx.font = '10px monospace';
                ctx.fillText("HP",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.fillStyle = 'white';
                ctx.fillText("AS",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("DOG",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("WGT",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.fillStyle = 'white';
                ctx.fillText("ATK",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("ACC",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("CRT",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("CD",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("CLU",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.fillStyle = 'white';
                ctx.fillText("SPD",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = 0
                y_shift = y_shift+20;


                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift-25,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("H",x_offset+x_shift+t_shift+12-25,y_offset+y_shift+14);

                ctx.fillStyle = calcColour(state[i][j][k].HP)
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].HP,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25
                ctx.fillStyle = calcColour(state[i][j][k].AS)
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].AS,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.DOG)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.DOG,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.Weight)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.Weight,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                ctx.fillStyle = calcColour(state[i][j][k].ATK)
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].ATK,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.ACC)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.ACC,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.CRT)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.CRT,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.CD)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.CD,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].H.CLU)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].H.CLU,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;

                //y_shift=y_shift+20;
                //x_shift=0;

                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift-25,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("L",x_offset+x_shift+t_shift+12-25,y_offset+y_shift+14);

                //ctx.fillStyle = calcColour(state[i][j][k].L.HP)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.HP,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = calcColour(state[i][j][k].L.DEF)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.DEF,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.DOG)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.DOG,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.Weight)
                // ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.Weight,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.ATK)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.ATK,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.ACC)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.ACC,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.CRT)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.CRT,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.CD)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.CD,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].L.CLU)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].L.CLU,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;

                //y_shift=y_shift+20;
                //x_shift=0;

                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift-25,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("R",x_offset+x_shift+t_shift+12-25,y_offset+y_shift+14);

                //ctx.fillStyle = calcColour(state[i][j][k].R.HP)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.HP,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = calcColour(state[i][j][k].R.DEF)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.DEF,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.DOG)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.DOG,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.Weight)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.Weight,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.ATK)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.ATK,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.ACC)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.ACC,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.CRT)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.CRT,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.CD)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.CD,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].R.CLU)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].R.CLU,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;

                //y_shift=y_shift+20;
                //x_shift=0;

                //ctx.fillStyle = '#848884';
                //ctx.fillRect(x_offset+x_shift+t_shift-25,y_offset+y_shift,24,19)
                //ctx.fillStyle = 'white';
                //ctx.fillText("B",x_offset+x_shift+t_shift+12-25,y_offset+y_shift+14);

                //ctx.fillStyle = calcColour(state[i][j][k].B.HP)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.HP,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25
                //ctx.fillStyle = calcColour(state[i][j][k].B.DEF)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.DEF,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.DOG)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.DOG,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.Weight)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.Weight,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.ATK)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.ATK,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.ACC)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.ACC,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.CRT)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.CRT,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.CD)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.CD,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                //ctx.fillStyle = calcColour(state[i][j][k].B.CLU)
                //ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                //ctx.font = '10px monospace';
                //ctx.fillStyle = 'black';
                //ctx.fillText(state[i][j][k].B.CLU,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                //x_shift = x_shift+25;
                ctx.fillStyle = calcColour(state[i][j][k].SPD)
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].SPD,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+50;


                // draw percentage circle
                let percentage = (100-state[i][j][k].TTD)/100
                let radius = 10;

                // Draw the progress circle
                ctx.beginPath();
                var startAngle = -Math.PI / 2; // Start at the top
                var endAngle = percentage * 2 * Math.PI + startAngle;
                ctx.arc(x_offset+x_shift+t_shift, y_offset+y_shift, radius, startAngle, endAngle, false);


                if((percentage != 0) && (percentage != 1)) {
                    // Close the path to form a wedge
                    ctx.lineTo(x_offset+x_shift+t_shift, y_offset+y_shift); // Draw line to center of the circle
                    ctx.closePath(); // Automatically draws a line from the last point to the starting point
                }
                ctx.strokeStyle = '#4caf50'; // y_offset+y_shift for the progress
                ctx.stroke();

                // Add text inside the circle
                ctx.font = "7px monospace";
                ctx.fillStyle = 'white';
                ctx.fillText((100*percentage).toFixed(0) + '%', x_offset+x_shift+t_shift, y_offset+y_shift+2);

                x_shift=0;
                y_shift=y_shift+25;
                drawEXPBar(x_offset+x_shift+t_shift,y_offset+y_shift+2, state[i][j][k].EXP)

                y_shift=y_shift+75;
            }
        }
        y_shift = 0;
        x_shift = 0;
        t_shift = t_shift + 700;
    }

}

function calcColour(n){
    if(n<0){
        return "rgb(0,0,0)"
    }

    if(n<=20){
        return "rgb("+80+(9*n)+",0,0)"
    }
    else if(n<71){
        return "rgb(255,"+(n-20)*5+",0)"
    }
    else {
        return "rgb("+(255-((n-71)*8))+",255,0)"
    }
    return "black"
}

function coordinateEqual(c1,c2) {
    if ((c1.X == c2.X) && (c1.Y == c2.Y)) {
        return true
    }
    return false
}

function drawRotated(s,x,y, thetas, path_segment, size, y_shift, colour) {

    let radians = 0;

    if ((thetas != null) && (thetas.length > 0)) {
        radians = thetas[path_segment];
    }

    ctx.save();
    ctx.fillStyle = colour;
    ctx.textAlign = "center";
    ctx.font = size+'px monospace';
    ctx.translate(x,y);
    ctx.rotate( (2*Math.PI) - radians );
    ctx.fillText(s, 0, y_shift ); // subtle y-axis adjustment since it is being from from top-left corner
    ctx.restore();
}

function getDistance(c1,c2) {
    let y_dist = c2.Y - c1.Y
    let x_dist = c2.X - c1.X
    return Math.sqrt((y_dist*y_dist)+(x_dist*x_dist))
}

function createDamageNumber(damage, x, y) {
    let color = 'white';
    let fontSize = 24;

    damageNumbers.push({
        x: x,
        y: y,
        damage: damage,
        color: color,
        fontSize: fontSize,
        scale: 1.2,  // Initial pop-up effect
        opacity: 1,
        lifetime: 0,  // Timer for animation
        duration: 1000,  // Total animation duration (in ms)
        isActive: true
    });
}

function updateDamageNumbers() {
    for (let i = 0; i < damageNumbers.length; i++) {
        const dmgNum = damageNumbers[i];

        if (dmgNum.isActive) {
            const progress = dmgNum.lifetime / dmgNum.duration;

            dmgNum.y -= 1.5; // Float upwards slightly

            // Scale back to normal size from pop-up
            if (dmgNum.lifetime < 200) {
                dmgNum.scale = 1.2 - (dmgNum.lifetime / 200) * 0.2;  // Shrink slightly over time
            }

            // Fade-out effect
            dmgNum.opacity = Math.max(0, 1 - progress);  // Reduces opacity as the lifetime progresses
            dmgNum.lifetime += 16;  // Increment lifetime (assuming ~60fps)

            // Deactivate number when animation is complete
            if (progress >= 1) {
                dmgNum.isActive = false;
            }
        }
    }

    // Remove inactive damage numbers
    damageNumbers = damageNumbers.filter(dmgNum => dmgNum.isActive);
}

function drawDamageNumbers() {
    for (let dmgNum of damageNumbers) {
        ctx.save();
        ctx.globalAlpha = dmgNum.opacity;  // Apply opacity for fade effect
        ctx.textAlign = 'center';
        ctx.fillText(dmgNum.damage, dmgNum.x, dmgNum.y);  // Draw damage number
        ctx.restore();
    }
}

function drawEXPBar(locationX, locationY, EXP) {

    const totalBars = 20;
    const barWidth = 3;
    const barHeight = 4;
    const gap = 1;
    const borderRadius = 2;

    const filled = Math.floor(((EXP%100)/100)*totalBars)

    for (let i = 0; i < totalBars; i++) {
        const xPos = locationX + i * (barWidth + gap)+18;
        let fillColor = '#333'; // gray
        if (i < filled) {
            fillColor = '#4A90E2'; // blue
        }
        ctx.beginPath();
        ctx.moveTo(xPos + borderRadius, locationY);
        ctx.arcTo(xPos + barWidth, locationY, xPos + barWidth, locationY + barHeight, borderRadius);
        ctx.arcTo(xPos + barWidth, locationY + barHeight, xPos, locationY + barHeight, borderRadius);
        ctx.arcTo(xPos, locationY + barHeight, xPos, locationY, borderRadius);
        ctx.arcTo(xPos, locationY, xPos + barWidth, locationY, borderRadius);
        ctx.closePath();
        ctx.fillStyle = fillColor;
        ctx.fill();
    }
    ctx.fillStyle = 'white';
    ctx.textAlign = "left";
    //ctx.fillText(`${EXP%100}/100`, locationX+95, locationY+barHeight);
    //ctx.font = '6px monospace';
    ctx.fillText(`EXP`, locationX+1, locationY+barHeight);
}

function drawObstacles() {

    // have to draw everything from bottom left corner not the top left corner
    // so we mirror the y-coordinate and shift it down by the board length
    const boardXY = 500;

    ctx.strokeStyle = "white";
    for(let i = 0; i < obstacles.length; i++ ){
        if(obstacles[i].Shape == "BEZIER") {
            ctx.beginPath();
            ctx.moveTo(width+obstacles[i].points[0].x, height-obstacles[i].points[0].y);
            ctx.bezierCurveTo(
                width+obstacles[i].points[1].x, height-obstacles[i].points[1].y,
                width+obstacles[i].points[2].x, height-obstacles[i].points[2].y,
                width+obstacles[i].points[3].x, height-obstacles[i].points[3].y
            );
            ctx.stroke();
        }
    }

    ctx.fillStyle = "red";
    for(let i = 0; i < intersections.length; i++) {
        ctx.fillRect(width+intersections[i].X, height-intersections[i].Y, 1, 1);
    }
    ctx.fillStyle = "white";

    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(width, height);
    ctx.lineTo(width, height-boardXY);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width, height-boardXY);
    ctx.lineTo(width+boardXY, height-boardXY);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width+boardXY, height-boardXY);
    ctx.lineTo(width+boardXY, height);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width+boardXY, height);
    ctx.lineTo(width, height);
    ctx.stroke();

}

function heuristic(node, goal) { // euclidean distance
    const dx = goal.X - node.X;
    const dy = goal.Y - node.Y;
    return Math.sqrt(dx * dx + dy * dy);
}

function aStar(s, g) {
    const openList = []; // nodes to explore
    const closedList = new Set(); // nodes completely explored, prevents backtracking

    // create start and goal nodes
    let startNode = new AStarNode(s, s.X, s.Y, 0, 0, null);
    let goalNode = new AStarNode(g, g.X, g.Y, 0, 0, null);

    startNode.H = heuristic(startNode, goalNode);
    startNode.F = startNode.G + startNode.H;
    openList.push(startNode);

    while (openList.length > 0) {
        openList.sort((a, b) => a.F - b.F); // sort by lowest f value (priority queue)
        const currentNode = openList.shift();

        if ( (currentNode.X == goalNode.X) && (currentNode.Y == currentNode.Y) ) { // found the goal, reconstruct and return the path
            const path = [];
            let node = currentNode;
            while (node !== null) {
                path.push([node.X, node.Y]);
                node = node.parent;
            }
            return path.reverse();
        }

        closedList.add(`${currentNode.X},${currentNode.Y}`);

        for (let i = 0; i < currentNode.origin.Edges.length; i++) {

            let neighbor = null;
            const firstChar = currentNode.origin.Edges[i][0];

            if( firstChar == 'd' ) {
                neighbor = dynamics[ parseInt(currentNode.origin.Edges[i].slice(1), 10) ];
            } else if ( firstChar == 'i' ) {
                neighbor = intersections[ parseInt(currentNode.origin.Edges[i].slice(1), 10) ];
            }


            if (closedList.has(`${neighbor.X},${neighbor.Y}`)) {
                continue;
            }

            const tentativeG = currentNode.G + currentNode.origin.EdgeLength[i];

            let found = null;
            for (let j = 0; j < openList.length; j++) {
                if ( (openList.X == neighbor.X) && (openList.Y == neighbor.Y) ) {
                    found = openList[j];
                    break;
                }
            }

            if (found == null) {
                let newNode = new AStarNode(neighbor, neighbor.X, neighbor.Y, tentativeG, heuristic(neighbor, goalNode), currentNode);
                openList.push(newNode);
            } else if (tentativeG < found.G) {
                found.G = tentativeG;
                found.H = heuristic(found, goalNode);
                found.F = found.G + found.H;
                found.parent = currentNode;
            }

        }
    }

    return [];
}

function drawComplexPath(path, color) {
    ctx.strokeStyle = color;
    for (let i = 0; i < path.length - 1; i++) {
        ctx.beginPath();    // Begin a new path
        ctx.moveTo(width+path[i][0], height - path[i][1]); // Move to the starting point
        ctx.lineTo(width+path[i + 1][0], height - path[i + 1][1]); // Draw a line to the ending point
        ctx.stroke();       // Render the line
    }
    ctx.strokeStyle = "white";
}

function campEqual(c1,c2) {
    if(
        (c1.Type == c2.Type) &&
        (c1.I == c2.I) &&
        (c1.J == c2.J) &&
        (c1.K == c2.K) &&
        (c1.X == c2.X) &&
        (c1.Y == c2.Y)
    ) {
        return true
    }

    return false
}

function getHeatMapColor(t) {
    t = 1 - t;
    // Clamp t between 0 and 1
    t = Math.max(0, Math.min(1, t));

    let r = 0, g = 0, b = 0;

    if (t <= 0.5) {
        // Green to Yellow (0 <= t <= 0.5)
        r = Math.floor(2 * t * 255); // Increase red
        g = 255; // Green stays at max
    } else {
        // Yellow to Red (0.5 < t <= 1)
        r = 255; // Red stays at max
        g = Math.floor((1 - (t - 0.5) * 2) * 255); // Decrease green
    }

    return `rgb(${r}, ${g}, ${b})`;
}

function drawHPBoxes(hp, maxHP, x, y) {

    let maxMultiple = 1;
    let multiple = 1;
    let firstBarHP = hp;
    let firstBarMax = maxHP;
    if(maxHP>100){
        maxMultiple = maxHP / 100;
        multiple = hp / 100;
        firstBarHP = hp % 100;
        firstBarMax = 100;
        if(firstBarHP == 0) {
            multiple -= 1
            firstBarHP = 100
        }
    }

    const borderWidth = 1;
    const barWidth = 5;
    const barHeight = 5;
    const percentage = firstBarHP / firstBarMax;
    const truePercentage = hp / maxHP;

    let boxes = Math.round(firstBarMax / 10);
    if(boxes < 1) {
        boxes = 1;
    }
    const boxWidth = boxes*barWidth;
    const boxShift = boxWidth/2;

    ctx.font = "`12px Arial";
    ctx.textAlign = "right";
    ctx.fillStyle = "white"; // Fill color
    //ctx.fillText(hp, x - boxShift - 5 , y - 20);

    const color = getHeatMapColor(truePercentage);

    ctx.fillStyle = color;
    ctx.lineWidth = borderWidth;
    ctx.fillRect(x - boxShift, y - 25, barWidth * boxes * percentage, barHeight);
    ctx.strokeStyle = "black";
    for (let i = 0; i < boxes; i++) {
        ctx.strokeRect(x + (i * barWidth) - boxShift, y - 25, barWidth, barHeight);
        if(i < maxMultiple) {
            if(i < multiple) {
                ctx.fillStyle = color;
            } else {
                ctx.fillStyle = "white";
            }
            ctx.fillRect(x + (i * barWidth) - boxShift + 2, y - 20 + 2, 1, 1);
        }
    }
}

function drawSprite() {


    let x_offset = Math.round(c.width/2) - 400
    let y_offset = Math.round(c.height/2) - 250

    let row = 0;

    let sourceWidth = 32;
    let sourceHeight = 32;

    let radians = 0;
    if ( (state[0][0][0].Thetas != null) && (state[0][0][0].Thetas.length > 0) ) {
        radians = state[0][0][0].Thetas[state[0][0][0].Path_Segment];
    }

    if(radians < Math.PI/8) {
        row = 2;
    } else if(radians < 3*Math.PI/8) {
        row = 3;
    } else if(radians < 5*Math.PI/8) {
        row = 4;
    } else if(radians < 7*Math.PI/8) {
        row = 5;
    } else if(radians < 9*Math.PI/8) {
        row = 6;
    } else if(radians < 11*Math.PI/8) {
        row = 7;
    } else if(radians < 13*Math.PI/8) {
        row = 0;
    } else if(radians < 15*Math.PI/8) {
        row = 1;
    } else {
        row = 2;
    }

    let s;
    if(state[0][0][0].Status == 1) {
        s = sprites[1];
    } else if(state[0][0][0].Status == 2) {
        s = sprites[2];
        sourceWidth = 32;
        sourceHeight = 40;
    } else {
        s = sprites[0];
    }

    ctx.drawImage(
        s,
        ( sourceWidth * (frameIndex % 4) ), ( sourceHeight * row ),
        sourceWidth, sourceHeight,
        x_offset - 20, y_offset,
        32, 32
    ); // (image, extract_x, extract_y, extract_w, extract_h, dest_x, dest_y, dest_w, dest_h)

    frameCount++;
    if(frameCount > frameCountMax) {
        frameCount = 0;
        frameIndex = (frameIndex + 1) % totalFrames;
    }

}

function drawSelected() {

    if( hoverable != null && selected_hoverable != null && hoverable.I >= 0 && hoverable.J < 0 && hoverable.K < 0 ) {

        const x_offset = Math.round(c.width/2)
        const y_offset = Math.round(c.height/2) - 375

        const gap = 5;
        const width = 250-gap;
        const height = 100;

        ctx.font = '11px monospace';
        ctx.strokeStyle = "white";

        // draw unit (LS)
        ctx.strokeRect(x_offset-(500/2), y_offset, width, height)

        // draw camp (RS)
        ctx.strokeRect(x_offset+gap, y_offset, width, height)
        let camper = camps[hoverable.I]
        let campString = "";
        if(hoverable.Type=="camp") {
            campString = "[CAMP]";
        }
        let iteratorY = ctx.measureText("M").width + 10; // trick to get height
        let iteratorX = 5;
        ctx.fillStyle = '#89CFF0';
        ctx.fillText(campString, x_offset+gap+iteratorX, y_offset+iteratorY)
        iteratorX += 5;
        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillStyle = 'white';
        ctx.fillText("HP " + camper.HP + " / " + camper.MaxHP, x_offset+gap+iteratorX, y_offset+iteratorY)


    }
    else if(hoverable != null && hoverable.I >= 0 && hoverable.J >= 0 && hoverable.K >= 0) {

        const x_offset = Math.round(c.width/2)
        const y_offset = Math.round(c.height/2) - 375

        const width = 500;
        const height = 100;

        ctx.strokeStyle = "white";
        ctx.strokeRect(x_offset-(width/2), y_offset, width, height)

        let unit = state[hoverable.I][hoverable.J][hoverable.K]

        ctx.font = '11px monospace';

        let botString = ""
        if(unit.Bot == true) {
            botString = "[BOT] "
        }
        let iteratorY = ctx.measureText("M").width + 10; // trick to get height
        let iteratorX = 5;
        ctx.fillStyle = 'red';
        ctx.fillText(botString, x_offset-(width/2)+iteratorX, y_offset+iteratorY)
        ctx.fillStyle = 'white';
        const botStringWidth = ctx.measureText(botString).width;

        if(uiTest == 0) {
            const levelString = "LV. " + (unit.EXP/100).toString();
            ctx.fillText(levelString, x_offset-(width/2) + botStringWidth + iteratorX, y_offset+iteratorY)
            const levelStringWidth = ctx.measureText(levelString).width;
            ctx.fillText(unit.Name, x_offset-(width/2) + botStringWidth + levelStringWidth + 5 + iteratorX, y_offset+iteratorY)
            
        } else if(uiTest == 1) {
            ctx.fillText(unit.Name, x_offset-(width/2) + botStringWidth + iteratorX, y_offset+iteratorY)
            const nameStringWidth = ctx.measureText(unit.Name).width
            ctx.fillText(" LV. " + (Math.floor(unit.EXP/100)), x_offset-(width/2) + botStringWidth + nameStringWidth + iteratorX, y_offset+iteratorY)
            
        }

        


    }


}