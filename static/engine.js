console.log("loaded js");

var c = document.getElementById("myCanvas");
const ctx = c.getContext("2d");

//c.width = window.innerWidth;
//c.height = window.innerHeight;

c.width = document.body.clientWidth;
c.height = document.body.clientHeight;

const dimension = 500;
const width = Math.round(c.width / 2) - (dimension / 2);
const height = Math.round(c.height / 2) + (dimension / 2);

console.log("width, height", width, height);

// shim layer with setTimeout fallback
window.requestAnimFrame = (function () {
    return window.requestAnimationFrame ||
        window.webkitRequestAnimationFrame ||
        window.mozRequestAnimationFrame ||
        window.oRequestAnimationFrame ||
        window.msRequestAnimationFrame ||
        function (callback) {
            window.setTimeout(callback, 1000 / 60);
        };
})();

window.cancelAnimationFrame = window.cancelAnimationFrame || window.mozCancelAnimationFrame || function (requestID) { clearTimeout(requestID) };

function anime() {

    window.requestAnimFrame(anime);
    ctx.fillStyle = '#101010';
    //ctx.clearRect(0,0,innerWidth,innerHeight);
    ctx.fillRect(0, 0, innerWidth, innerHeight);

    let start = Date.now().toLocaleString('en-CH');

    if (game_over == true) {
        ctx.textAlign = "center";
        ctx.fillStyle = 'blue';
        ctx.fillRect((c.width / 2) - (ctx.measureText("GAMEOVER").width / 2), 42, ctx.measureText("GAMEOVER").width, ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText("GAMEOVER", c.width / 2, 50);
    }
    else if (sent_time != null) {
        ctx.textAlign = "center";
        remaining_time = sent_time - start.replaceAll("’", "")
        ctx.fillStyle = 'black';
        ctx.fillRect(c.width / 2, 40, ctx.measureText(remaining_time).width, ctx.measureText('M').width);
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(remaining_time, c.width / 2, 50);
    }

    if (phase == "CHARACTER SELECTION") {
        //drawGrid()
        drawPhase()
        drawLog()
        //drawStarters()
        //drawTable1()
        //drawTable2()
        //drawStartersTable1()
        //drawStartersTable2()
        //drawSelect()
        //ctx.fillRect(1150,40 , 1, 800);

    } else if (match_data.length > 0) {

        drawObstacles();
        calcPos();
        drawAttack();
        drawPos();
        drawHoverable();
        drawCamps();
        drawPotentialPaths();
        drawPaths();
        drawDynamicCircles();
        drawLastMileCircle();
        drawDamageNumbers();
        updateDamageNumbers();
        drawTurn();
        drawUnitsOfTime();
        drawLog();
        drawSelected();
        drawUnitSelect();

        //temp
        drawCampData();

        if (view == 0) {
            drawHeatMap()
        }

        //drawSprite(); // I think it's too distracting right now

        if( randomMoveFrameCount > 0) {
            ctx.font = '9px monospace';
            ctx.textAlign = "left";
            ctx.fillStyle = 'white'
            ctx.fillText(randomMove, 10, 10);
        }

    }

    frame_count += 1
    if (frame_count == 15) {
        frame_count = 0;
        if (even == 0) { even = 1 }
        else { even = 0 }
    }

    if (randomMoveFrameCount > 0) {
        randomMoveFrameCount -= 1;
    } 

}
anime()

function calcPos() {
    if ((animating_state < match_data.length) && (attacking_frames_count <= 0)) {

        let all_status_correct = true;
        for (let i = 0; i < state.length; i++) {
            for (let j = 0; j < state[i].length; j++) {
                for (let k = 0; k < state[i][j].length; k++) {
                    if (state[i][j][k].Status != match_data[animating_state][i][j][k].Status) {
                        all_status_correct = false;
                        if (state[i][j][k].Status == 0) {
                            state[i][j][k].Status = 1;
                            state[i][j][k].Start = match_data[animating_state][i][j][k].Start;
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft;
                            state[i][j][k].Path = match_data[animating_state][i][j][k].Path;
                            state[i][j][k].Path_Segment = 0;
                            state[i][j][k].Thetas = match_data[animating_state][i][j][k].Path;
                        } else if (state[i][j][k].Status == 1) {
                            if (state[i][j][k].DistanceLeft <= 0) {
                                if ( state[i][j][k].DestinationIndex.length == 1 && camps[state[i][j][k].DestinationIndex[0]].HP > 0) {
                                    state[i][j][k].Status = 2;
                                    state[i][j][k].Battle = camps[state[i][j][k].DestinationIndex[0]]
                                    camps[state[i][j][k].DestinationIndex[0]].Units.push([i, j, k])
                                } else {
                                    state[i][j][k].Status = 0;
                                }
                            }
                        } else if (state[i][j][k].Status == 2) {
                            if ((state[i][j][k].Battle != null) && (state[i][j][k].Battle.HP <= 0)) {
                                state[i][j][k].Status = 0;
                                state[i][j][k].TTD = state[i][j][k].AS;
                                state[i][j][k].Battle.Units = []
                                state[i][j][k].Battle = null;
                            }
                        }
                    }
                }
            }
        }

        let all_distances_correct = true;
        for (let i = 0; i < state.length; i++) {
            for (let j = 0; j < state[i].length; j++) {
                for (let k = 0; k < state[i][j].length; k++) {
                    if (state[i][j][k].Status == 1) {

                        if (!arraysEqual(state[i][j][k].DestinationIndex, match_data[animating_state][i][j][k].DestinationIndex) || !arraysEqual(state[i][j][k].Destination, match_data[animating_state][i][j][k].Destination)) {
                            state[i][j][k].Start = match_data[animating_state][i][j][k].Start;
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft;
                            state[i][j][k].Path = match_data[animating_state][i][j][k].Path;
                            state[i][j][k].Path_Segment = 0;
                        }
                        
                        if (state[i][j][k].DistanceLeft != match_data[animating_state][i][j][k].DistanceLeft) {
                            all_distances_correct = false;

                            let SPD = SPDTable[animating_state][i][j][k];
                            if (state[i][j][k].LastMile == 3) {
                                SPD = 0;
                            }

                            let toMove = SPD * 0.015;
                            state[i][j][k].DistanceLeft -= toMove;
                            while (toMove > 0) {
                                if (state[i][j][k].Path_Segment + 1 >= state[i][j][k].Path.length) {
                                    toMove = 0;
                                    state[i][j][k].DistanceLeft = 0;
                                } else {
                                    let seg_x = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][0] - state[i][j][k].Position[0]
                                    let seg_y = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][1] - state[i][j][k].Position[1]
                                    let segment_distance = Math.sqrt((seg_x * seg_x) + (seg_y * seg_y))
                                    toMove -= segment_distance
                                    if (toMove >= 0) {
                                        state[i][j][k].Position[0] = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][0]
                                        state[i][j][k].Position[1] = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][1]
                                        state[i][j][k].Path_Segment++
                                    } else { // the negative toMove distance left is the distance left in the segment
                                        let percentLeft = -toMove / segment_distance
                                        state[i][j][k].Position[0] = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][0] - (seg_x * percentLeft)
                                        state[i][j][k].Position[1] = state[i][j][k].Path[state[i][j][k].Path_Segment + 1][1] - (seg_y * percentLeft)
                                    }
                                }
                            }

                            if (state[i][j][k].DistanceLeft - 0.01 <= match_data[animating_state][i][j][k].DistanceLeft) {
                                state[i][j][k].Position = match_data[animating_state][i][j][k].Position
                                state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft
                                state[i][j][k].Path_Segment = match_data[animating_state][i][j][k].Path_Segment
                            }

                        }
                    } else if (state[i][j][k].Status == 2) {
                        if (state[i][j][k].TTD != match_data[animating_state][i][j][k].TTD) {
                            all_distances_correct = false;
                            state[i][j][k].TTD -= (0.015 * state[i][j][k].AS)
                            if (state[i][j][k].TTD <= match_data[animating_state][i][j][k].TTD) {
                                state[i][j][k].TTD = match_data[animating_state][i][j][k].TTD
                            }
                            if (state[i][j][k].TTD == 0) {
                                attacking_frames_count = 1
                            }
                        }
                    }
                }
            }
        }

        if (all_status_correct && all_distances_correct) {
            if ((time_queue[animating_state] != "") && (attacking_frames_count == 0)) { // update time stamp
                total_units_of_time += parseFloat(time_queue[animating_state])
                total_units_of_time = parseFloat(total_units_of_time.toFixed(3))
                event_log.push(time_queue[animating_state].toString() + " ticks have passed")
                if (event_log.length >= 25) {
                    event_log.shift()
                }
            }
            state = JSON.parse(JSON.stringify(match_data[animating_state])) // deep copy
            camps = camp_data[animating_state]
            turn = turn_queue[animating_state]
            animating_state++
            setupHoverable()
        }
    }
}

function drawAttack() {

    if ((animating_state < match_data.length) && (attacking_frames_count > 0)) {

        for (let i = 0; i < atk_data[animating_state].length; i++) {
            let atk_obj = JSON.parse(atk_data[animating_state][i]);
            let atki = atk_obj.Attacker[0]
            let atkj = atk_obj.Attacker[1]
            let atkk = atk_obj.Attacker[2]

            let campOfInterest = state[atki][atkj][atkk].Battle
            campOfInterest.HP -= state[atki][atkj][atkk].ATK

            if (attacking_frames_count == 1) {
                createDamageNumber(state[atki][atkj][atkk].ATK, width + campOfInterest.Position[0], height - campOfInterest.Position[1]);
                event_log.push(atk_data[animating_state][i])
                if (event_log.length >= 25) {
                    event_log.shift()
                }
            }
        }

        attacking_frames_count++
        if (attacking_frames_count >= 1) { attacking_frames_count = 0 }
    }
}

function drawLevelUp() {

}

function drawPos() {

    ctx.fillStyle = "white";
    ctx.textAlign = "left";
    ctx.font = '11px monospace';
    for (let i = 0; i < state.length; i++) {
        for (let j = 0; j < state[i].length; j++) {
            for (let k = 0; k < state[i][j].length; k++) {
                if (state[i][j][k].HP > 0) {
                    //drawHPBoxes(state[i][j][k].HP,state[i][j][k].HP,width+state[i][j][k].Position[0],height-state[i][j][k].Position[1])
                    
                    if(i === 0) {
                        drawRotated('>', width + state[i][j][k].Position[0], height - state[i][j][k].Position[1], state[i][j][k].Thetas, state[i][j][k].Path_Segment, "11", 3, "yellow")
                    } else {
                        drawRotated('>', width + state[i][j][k].Position[0], height - state[i][j][k].Position[1], state[i][j][k].Thetas, state[i][j][k].Path_Segment, "11", 3, "red")
                    }
                }
            }
        }
    }
}

function drawCamps() {

    ctx.fillStyle = "green";

    for (let i = 0; i < camps.length; i++) {
        ctx.fillRect(width + camps[i].Position[0], height - camps[i].Position[1], 1, 1);

        let x_shift_tracker = []
        x_shift_tracker.push([]) // list of my_team units at this camp
        x_shift_tracker.push([]) // list of enemy units at this camp
        for (let j = 0; j < camps[i].Units.length; j++) {
            if ((camps[i].Units[j][0] == my_team) && (camps[i].Units[j][1] == my_int)) { x_shift_tracker[0].push(0) }
            else { x_shift_tracker[1].push(0) }
        }

        // this draws the attack circles
        for (let j = 0; j < camps[i].Units.length; j++) {
            let unitOfInterest = state[camps[i].Units[j][0]][camps[i].Units[j][1]][camps[i].Units[j][2]]
            if (unitOfInterest.Status == 2) {

                let x_shift = 0;
                let y_shift = 15;
                let odd = x_shift_tracker[0].length % 2;
                if (x_shift_tracker[0].length == 1) { odd = 0; }
                if ((camps[i].Units[j][0] == my_team) && (camps[i].Units[j][1] == my_int)) {
                    y_shift = -15;
                    x_shift = -Math.floor(x_shift_tracker[0].length / 2) * 10 - (odd * 10)
                    x_shift += x_shift_tracker[0][0] * 20
                    x_shift_tracker[0][0] += 1
                } else {
                    odd = x_shift_tracker[1].length % 2;
                    if (x_shift_tracker[1].length == 1) { odd = 0; }
                    x_shift = -Math.floor(x_shift_tracker[1].length / 2) * 10 - (odd * 10)
                    x_shift += x_shift_tracker[1][0] * 20
                    x_shift_tracker[1][0] += 1
                }

                // draw percentage circle
                let percentage = (100 - unitOfInterest.TTD) / 100

                // Draw the progress circle
                ctx.beginPath();
                var startAngle = -Math.PI / 2; // Start at the top
                var endAngle = percentage * 2 * Math.PI + startAngle;
                ctx.arc(width + (unitOfInterest.Position[0]) + x_shift, height - unitOfInterest.Position[1] + y_shift, 5, startAngle, endAngle, false);

                if ((percentage != 0) && (percentage != 1)) { // Close the path to form a wedge
                    ctx.lineTo(width + unitOfInterest.Position[0] + x_shift, height - unitOfInterest.Position[1] + y_shift); // Draw line to center of the circle
                    ctx.closePath(); // Automatically draws a line from the last point to the starting point
                }
                if ((camps[i].Units[j][0] == my_team) && (camps[i].Units[j][1] == my_int)) { ctx.strokeStyle = 'blue'; }
                else { ctx.strokeStyle = 'green'; }
                ctx.stroke();
            }
        }

        // draw HP
        ctx.font = '9px monospace';
        ctx.textAlign = "center";
        //ctx.fillStyle = '#E0E0E0'
        ctx.fillStyle = '#F5F5F5'

        if (camps[i].HP > 0) {
            //drawHPBoxes(camps[i].HP, camps[i].MaxHP, width + camps[i].Position[0], height - camps[i].Position[1])
            ctx.fillText(camps[i].HP, width + camps[i].Position[0], height - camps[i].Position[1] - 10);
        }
    }

    ctx.fillStyle = "white";
}

function drawHoverable() {

    ctx.font = '7px monospace';
    ctx.textAlign = "center";

    if (hovIdx.length !== 0) {

        let current;
        if( hovIdx.length === 3 ) {
            current = state[hovIdx[0]][hovIdx[1]][hovIdx[2]];
            ctx.fillText(current.Name, width + current.Position[0], height - current.Position[1] - 15 );
        } else if( hovIdx.length === 1 ) {
            current = camps[hovIdx[0]];
            ctx.fillText("CAMP " + hovIdx[0].toString(), width + current.Position[0], height - current.Position[1] - 15 );
        }

        ctx.fillStyle = 'pink'
        ctx.fillRect(width + current.Position[0] - 10, height - current.Position[1] + 10, 20, 1);
        ctx.fillRect(width + current.Position[0] - 10, height - current.Position[1] - 10, 20, 1);
        ctx.fillRect(width + current.Position[0] - 10, height - current.Position[1] - 10, 1, 20);
        ctx.fillRect(width + current.Position[0] + 10, height - current.Position[1] - 10, 1, 20);
        ctx.fillStyle = 'white'

    }
}

function drawPhase() {
    ctx.textAlign = "center";
    ctx.fillStyle = 'black';
    ctx.fillRect(c.width / 2, 10, ctx.measureText(phase).width, ctx.measureText('M').width);

    ctx.fillStyle = 'white'
    ctx.textAlign = "center";
    ctx.fillText(phase, c.width / 2, 20);
}

function drawTurn() {
    ctx.textAlign = "center";
    ctx.fillStyle = 'black';
    ctx.fillRect(c.width / 2, 10, ctx.measureText(turn).width, ctx.measureText('M').width);

    ctx.fillStyle = 'white'
    ctx.textAlign = "center";
    ctx.fillText(turn, c.width / 2, 20);
}

function drawUnitsOfTime() {
    ctx.textAlign = "center";
    ctx.font = '11px monospace';
    ctx.fillStyle = 'white';
    time_string = total_units_of_time + " ticks"
    ctx.fillText(time_string, c.width / 2, 35);
}

function drawLog() {

    let x = Math.round(c.width / 2) - 250;
    let y = Math.round(c.height / 2) + 300;

    if (view == 1) {
        x = Math.round(c.width / 2) - 250
        y = Math.round(c.height / 2) + 300
    }

    for (i = 0; i < event_log.length; i++) {
        ctx.textAlign = "left";
        ctx.font = '11px monospace';
        ctx.fillStyle = 'white';
        ctx.fillText(event_log[i], x, y + (10 * i));
    }
}

function drawHeatMap() {

    let x_offset = Math.round(c.width / 2) - 450
    let y_offset = Math.round(c.height / 2) - 250

    let x_shift = 0;
    let y_shift = 0;

    let t_shift = 0;

    for (let i = 0; i < state.length; i++) {
        for (let j = 0; j < state[i].length; j++) {
            for (let k = 0; k < state[i][j].length; k++) {

                // ready for action or not
                if (state[i][j][k].HP > 0) {
                    if (state[i][j][k].Status == 0) {
                        ctx.beginPath();
                        ctx.strokeStyle = 'green';
                        ctx.arc(x_offset + x_shift + t_shift - 10, y_offset + y_shift - 2, 3, 0, 2 * Math.PI); // (x, y, radius, startAngle, endAngle)
                        ctx.stroke();
                    }
                    else {
                        ctx.beginPath();
                        ctx.strokeStyle = 'white';
                        ctx.arc(x_offset + x_shift + t_shift - 10, y_offset + y_shift - 2, 3, 0, 2 * Math.PI); // (x, y, radius, startAngle, endAngle)
                        ctx.stroke();
                    }
                }

                // write name and level
                ctx.textAlign = "left";
                ctx.font = '8px monospace';
                ctx.fillStyle = 'white'
                if (state_machine == 0) {
                    if ( hovIdx.length === 3 ) { // is an unit
                        if ((hovIdx[0] == i) && (hovIdx[1] == j) && (hovIdx[2] == k)) {
                            ctx.fillStyle = 'red'
                        }
                    }
                } else if (state_machine == 1) {
                    if ( hovIdx2.length === 3 ) { // is an unit
                        if ((hovIdx2[0] == i) && (hovIdx2[1] == j) && (hovIdx2[2] == k)) {
                            ctx.fillStyle = 'red'
                        }
                    }
                }
                ctx.fillText(state[i][j][k].Name, x_offset + x_shift + t_shift, y_offset + y_shift);
                x_shift += 24 * 5
                ctx.textAlign = "right";
                ctx.fillText("LV. " + (Math.floor(state[i][j][k].EXP / 100) + 1), x_offset + x_shift + t_shift, y_offset + y_shift);
                x_shift = 0;
                y_shift = y_shift + 5

                // write headers
                ctx.textAlign = "center";
                ctx.fillStyle = '#848884';

                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.fillStyle = 'white';
                ctx.font = '10px monospace';
                ctx.fillText("HP", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.fillStyle = 'white';
                ctx.fillText("AS", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.fillStyle = 'white';
                ctx.fillText("ATK", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.fillStyle = 'white';
                ctx.fillText("DEF", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                
                ctx.fillStyle = '#848884';
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.fillStyle = 'white';
                ctx.fillText("SPD", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = 0
                y_shift = y_shift + 20;


                // write stats
                ctx.fillStyle = calcColour(state[i][j][k].HP)
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].HP, x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                ctx.fillStyle = calcColour(state[i][j][k].AS)
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].AS, x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25;

                ctx.fillStyle = calcColour(state[i][j][k].ATK)
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].ATK, x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25;

                ctx.fillStyle = calcColour(state[i][j][k].DEF)
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].DEF,x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25;

                ctx.fillStyle = calcColour(state[i][j][k].SPD)
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].SPD, x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 50;


                // draw percentage circle
                let percentage = (100 - state[i][j][k].TTD) / 100
                let radius = 10;

                // Draw the progress circle
                ctx.beginPath();
                var startAngle = -Math.PI / 2; // Start at the top
                var endAngle = percentage * 2 * Math.PI + startAngle;
                ctx.arc(x_offset + x_shift + t_shift, y_offset + y_shift, radius, startAngle, endAngle, false);


                if ((percentage != 0) && (percentage != 1)) {
                    // Close the path to form a wedge
                    ctx.lineTo(x_offset + x_shift + t_shift, y_offset + y_shift); // Draw line to center of the circle
                    ctx.closePath(); // Automatically draws a line from the last point to the starting point
                }
                ctx.strokeStyle = '#4caf50'; // y_offset+y_shift for the progress
                ctx.stroke();

                // Add text inside the circle
                ctx.font = "7px monospace";
                ctx.fillStyle = 'white';
                ctx.fillText((100 * percentage).toFixed(0) + '%', x_offset + x_shift + t_shift, y_offset + y_shift + 2);

                // draw scaling
                y_shift = y_shift + 20;
                x_shift = 0;
                ctx.fillStyle = "gray";
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].Scale[0]+"%", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25
                ctx.fillStyle = "gray";
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].Scale[1]+"%", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25;
                ctx.fillStyle = "gray";
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].Scale[2]+"%", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);
                x_shift = x_shift + 25;
                ctx.fillStyle = "gray";
                ctx.fillRect(x_offset+x_shift+t_shift,y_offset+y_shift,24,19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].Scale[3]+"%",x_offset+x_shift+t_shift+12,y_offset+y_shift+14);
                x_shift = x_shift+25;
                ctx.fillStyle = "gray";
                ctx.fillRect(x_offset + x_shift + t_shift, y_offset + y_shift, 24, 19)
                ctx.font = '10px monospace';
                ctx.fillStyle = 'black';
                ctx.fillText(state[i][j][k].Scale[4]+"%", x_offset + x_shift + t_shift + 12, y_offset + y_shift + 14);

                ctx.font = "7px monospace";
                ctx.fillStyle = 'white';
                x_shift = 0;
                y_shift = y_shift + 25;
                drawEXPBar(x_offset + x_shift + t_shift, y_offset + y_shift + 2, state[i][j][k].EXP)

                x_shift += 110;
                ctx.fillText( state[i][j][k].EXP, x_offset + x_shift + t_shift, y_offset + y_shift + 5);

                x_shift += 30;
                ctx.fillText( state[i][j][k].Status, x_offset + x_shift + t_shift, y_offset + y_shift + 5);
                
                x_shift += 10;
                ctx.fillText( state[i][j][k].LastMile, x_offset + x_shift + t_shift, y_offset + y_shift + 5);

                x_shift += 10;
                ctx.fillText( state[i][j][k].DestinationIndex, x_offset + x_shift + t_shift, y_offset + y_shift + 5);
                
                x_shift = 0;
                ctx.fillText( state[i][j][k].Start[0].toString() + " " + state[i][j][k].Start[1].toString(), x_offset + x_shift + t_shift, y_offset + y_shift + 15);
                ctx.fillText( state[i][j][k].Position[0].toString() + " " + state[i][j][k].Position[1].toString(), x_offset + x_shift + t_shift, y_offset + y_shift + 25);
                ctx.fillText( state[i][j][k].Destination[0].toString() + " " + state[i][j][k].Destination[1].toString(), x_offset + x_shift + t_shift, y_offset + y_shift + 35);
                ctx.fillText( state[i][j][k].DistanceLeft, x_offset + x_shift + t_shift, y_offset + y_shift + 45);

                x_shift = 0;
                y_shift = y_shift + 75;
            }
        }
        y_shift = 0;
        x_shift = 0;
        t_shift = t_shift + 775;
    }

}

function calcColour(n) {
    if (n < 0) {
        return "rgb(0,0,0)"
    }

    if (n <= 20) {
        return "rgb(" + 80 + (9 * n) + ",0,0)"
    }
    else if (n < 71) {
        return "rgb(255," + (n - 20) * 5 + ",0)"
    }
    else {
        return "rgb(" + (255 - ((n - 71) * 8)) + ",255,0)"
    }
    return "black"
}

function coordinateEqual(c1, c2) {
    if ((c1[0] == c2[0]) && (c1[1] == c2[1])) {
        return true
    }
    return false
}

function drawRotated(s, x, y, thetas, path_segment, size, y_shift, colour) {

    let radians = 0;

    if ((thetas != null) && (thetas.length > 0)) {
        radians = thetas[path_segment];
    }

    ctx.save();
    ctx.fillStyle = colour;
    ctx.textAlign = "center";
    ctx.font = size + 'px monospace';
    ctx.translate(x, y);
    ctx.rotate((2 * Math.PI) - radians);
    ctx.fillText(s, 0, y_shift); // subtle y-axis adjustment since it is being from from top-left corner
    ctx.restore();
}

function getDistance(c1, c2) {
    let y_dist = c2[1] - c1[1];
    let x_dist = c2[0] - c1[0];
    return Math.sqrt((y_dist * y_dist) + (x_dist * x_dist));
}

function createDamageNumber(damage, x, y) {
    let color = 'white';
    let fontSize = 24;

    damageNumbers.push({
        Position: [x,y],
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
        ctx.fillText(dmgNum.damage, dmgNum.Position[0], dmgNum.Position[1]);  // Draw damage number
        ctx.restore();
    }
}

function drawEXPBar(locationX, locationY, EXP) {

    const totalBars = 20;
    const barWidth = 3;
    const barHeight = 4;
    const gap = 1;
    const borderRadius = 2;

    const filled = Math.floor(((EXP % 100) / 100) * totalBars)

    for (let i = 0; i < totalBars; i++) {
        const xPos = locationX + i * (barWidth + gap) + 18;
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
    ctx.fillText(`EXP`, locationX + 1, locationY + barHeight);
}

function drawObstacles() {

    // have to draw everything from bottom left corner not the top left corner
    // so we mirror the y-coordinate and shift it down by the board length
    const boardXY = 500;

    ctx.strokeStyle = "white";
    for (let i = 0; i < obstacles.length; i++) {
        if (obstacles[i].Shape == "BEZIER") {
            ctx.beginPath();
            ctx.moveTo(width + obstacles[i].points[0].x, height - obstacles[i].points[0].y);
            ctx.bezierCurveTo(
                width + obstacles[i].points[1].x, height - obstacles[i].points[1].y,
                width + obstacles[i].points[2].x, height - obstacles[i].points[2].y,
                width + obstacles[i].points[3].x, height - obstacles[i].points[3].y
            );
            ctx.stroke();
        }
    }

    if(intersectionsOn) {
        ctx.fillStyle = "red";
        for (let i = 0; i < intersections.length; i++) {
            ctx.fillRect(width + intersections[i].Position[0], height - intersections[i].Position[1], 1, 1);
        }
    }

    ctx.fillStyle = "white";

    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(width, height);
    ctx.lineTo(width, height - boardXY);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width, height - boardXY);
    ctx.lineTo(width + boardXY, height - boardXY);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width + boardXY, height - boardXY);
    ctx.lineTo(width + boardXY, height);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width + boardXY, height);
    ctx.lineTo(width, height);
    ctx.stroke();

}

function drawComplexPath(path, color) {
    ctx.strokeStyle = color;
    for (let i = 0; i < path.length - 1; i++) {
        ctx.beginPath();    // Begin a new path
        ctx.moveTo(width + path[i][0], height - path[i][1]); // Move to the starting point
        ctx.lineTo(width + path[i + 1][0], height - path[i + 1][1]); // Draw a line to the ending point
        ctx.stroke();       // Render the line
    }
    ctx.strokeStyle = "white";
}

function arraysEqual(a, b) {
    if (a.length !== b.length) return false;
    
    for (let i = 0; i < a.length; i++) {
        if (a[i] !== b[i]) return false;
    }

    return true;
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
    if (maxHP > 100) {
        maxMultiple = maxHP / 100;
        multiple = hp / 100;
        firstBarHP = hp % 100;
        firstBarMax = 100;
        if (firstBarHP == 0) {
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
    if (boxes < 1) {
        boxes = 1;
    }
    const boxWidth = boxes * barWidth;
    const boxShift = boxWidth / 2;

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
        if (i < maxMultiple) {
            if (i < multiple) {
                ctx.fillStyle = color;
            } else {
                ctx.fillStyle = "white";
            }
            ctx.fillRect(x + (i * barWidth) - boxShift + 2, y - 20 + 2, 1, 1);
        }
    }
}

function drawSprite() {

    let x_offset = Math.round(c.width / 2) - 400
    let y_offset = Math.round(c.height / 2) - 250

    let row = 0;

    let sourceWidth = 32;
    let sourceHeight = 32;

    let radians = 0;
    if ((state[0][0][0].Thetas != null) && (state[0][0][0].Thetas.length > 0)) {
        radians = state[0][0][0].Thetas[state[0][0][0].Path_Segment];
    }

    if (radians < Math.PI / 8) {
        row = 2;
    } else if (radians < 3 * Math.PI / 8) {
        row = 3;
    } else if (radians < 5 * Math.PI / 8) {
        row = 4;
    } else if (radians < 7 * Math.PI / 8) {
        row = 5;
    } else if (radians < 9 * Math.PI / 8) {
        row = 6;
    } else if (radians < 11 * Math.PI / 8) {
        row = 7;
    } else if (radians < 13 * Math.PI / 8) {
        row = 0;
    } else if (radians < 15 * Math.PI / 8) {
        row = 1;
    } else {
        row = 2;
    }

    let s;
    if (state[0][0][0].Status == 1) {
        s = sprites[1];
    } else if (state[0][0][0].Status == 2) {
        s = sprites[2];
        sourceWidth = 32;
        sourceHeight = 40;
    } else {
        s = sprites[0];
    }

    ctx.drawImage(
        s,
        (sourceWidth * (frameIndex % 4)), (sourceHeight * row),
        sourceWidth, sourceHeight,
        x_offset - 20, y_offset,
        32, 32
    ); // (image, extract_x, extract_y, extract_w, extract_h, dest_x, dest_y, dest_w, dest_h)

    frameCount++;
    if (frameCount > frameCountMax) {
        frameCount = 0;
        frameIndex = (frameIndex + 1) % totalFrames;
    }

}

function drawSelected() {

    let unitDraw = hovIdx.length === 3; // draw the hovered unit's stats
    let opponentDraw = (hovIdx2.length === 3) && (hovIdx.length === 1);

    if (unitDraw || opponentDraw) {

        const x_offset = Math.round(c.width / 2);
        const y_offset = Math.round(c.height / 2) - 400;

        const width = 500;
        const height = 125;

        ctx.strokeStyle = "white";

        let unit;

        // drawing the square
        if (unitDraw) {
            ctx.strokeRect(x_offset - (width / 2), y_offset, width, height);
            if(hovIdx.length === 1) {
                unit = camps[hovIdx[0]];
            } else if(hovIdx.length === 3) {
                unit = state[hovIdx[0]][hovIdx[1]][hovIdx[2]];
            }
        } else if (opponentDraw) {
            if(hovIdx2.length === 1) {
                unit = camps[hovIdx2[0]];
            } else if(hovIdx2.length === 3) {
                unit = state[hovIdx2[0]][hovIdx2[1]][hovIdx2[2]];
            }
        }

        ctx.font = '11px monospace';

        let botString = "";
        if (unit.Bot == true) {
            botString = "[BOT] ";
        }
        let iteratorY = ctx.measureText("M").width + 10; // trick to get height
        let iteratorX = 5;
        ctx.fillStyle = 'red';
        ctx.fillText(botString, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);
        ctx.fillStyle = 'white';
        const botStringWidth = ctx.measureText(botString).width;

        ctx.fillText(unit.Name, x_offset - (width / 2) + botStringWidth + iteratorX, y_offset + iteratorY);
        const nameStringWidth = ctx.measureText(unit.Name).width;
        ctx.fillText(" LV. " + (unit.EXP / 100), x_offset - (width / 2) + botStringWidth + nameStringWidth + iteratorX, y_offset + iteratorY);

        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillText("HP " + unit.HP + " / " + unit.MaxHP, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);

        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillText("AS " + unit.AS, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);

        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillText("ATK " + unit.ATK, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);

        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillText("DPS " + unit.ATK * unit.AS, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);

        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillText("SPD " + unit.SPD, x_offset - (width / 2) + iteratorX, y_offset + iteratorY);


        let iteratorX2 = 125;
        let iteratorY2 = (ctx.measureText("M").width + 10) * 3;

        if (unitDraw) {
            let ETA;
            if(unit.SPD === 0) {
                ETA = "NAN";
            } else {
                ETA = (unit.DistanceLeft / unit.SPD).toFixed(2);
            }
            ctx.fillText("ETA " + ETA, x_offset - (width / 2) + iteratorX2, y_offset + iteratorY2);
            iteratorY2 += ctx.measureText("M").width + 10;

        } else if (opponentDraw) {
            let ETA;
            if(unit.SPD === 0) {
                ETA = "NAN";
            } else {
                let pathIndex = -1;
                for(let i = 0; i < paths.length; i++) {
                    if( (paths[i][paths[i].length - 1][0] === unit.Position[0]) &&
                        (paths[i][paths[i].length - 1][1] === unit.Position[1]) ) {
                        pathIndex = i;
                        break;
                    }
                }
                if(pathIndex >= 0) {
                    ETA = (pathLengths[pathIndex] / unit.SPD).toFixed(2);
                    ctx.fillText("ETA " + ETA, x_offset - (width / 2) + iteratorX2, y_offset + iteratorY2);
                    iteratorY2 += ctx.measureText("M").width + 10;
                } else {
                    ctx.fillText("ETA -", x_offset - (width / 2) + iteratorX2, y_offset + iteratorY2);
                    iteratorY2 += ctx.measureText("M").width + 10;
                }
            }
        }

        ctx.fillText("TTD " + (unit.TTD/unit.AS).toFixed(2), x_offset - (width / 2) + iteratorX2, y_offset + iteratorY2);
        iteratorY2 += ctx.measureText("M").width + 10;
        const TOT = unit.HP + unit.AS + unit.ATK + unit.DEF + unit.SPD;
        ctx.fillText("TOT " + TOT, x_offset - (width / 2) + iteratorX2, y_offset + iteratorY2);

    }



    if (opponentDraw) {

        const x_offset = Math.round(c.width / 2)
        const y_offset = Math.round(c.height / 2) - 400

        const gap = 5;
        const width = 250 - gap;
        const height = 125;

        ctx.font = '11px monospace';
        ctx.strokeStyle = "white";

        // draw unit (LS)
        ctx.strokeRect(x_offset - (500 / 2), y_offset, width, height)

        // draw camp (RS)
        ctx.strokeRect(x_offset + gap, y_offset, width, height)
        let camper = camps[hovIdx[0]];
        let campString = "";
        if ( hovIdx.length === 1 ) {
            campString = "[CAMP]";
        }
        let iteratorY = ctx.measureText("M").width + 10; // trick to get height
        let iteratorX = 5;
        ctx.fillStyle = '#89CFF0';
        ctx.fillText(campString, x_offset + gap + iteratorX, y_offset + iteratorY)
        iteratorX += 5;
        iteratorY += ctx.measureText("M").width + 10;
        ctx.fillStyle = 'white';
        ctx.fillText("HP " + camper.HP + " / " + camper.MaxHP, x_offset + gap + iteratorX, y_offset + iteratorY)

    }

}

function drawUnitSelect() {
    const x_offset = Math.round(c.width / 2)
    const y_offset = Math.round(c.height / 2) - 415

    const width = 500;
    const height = 125;

    let increment = 0;
    let oppIncrement = 0;

    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {

                if( hovIdx.length === 3 ) {
                    ctx.strokeStyle = "red";
                }
                else if ( state[i][j][k].Status == 0 ) {
                    ctx.strokeStyle = "green";
                }
                else {
                    ctx.strokeStyle = "white";
                }

                if( (i == my_team) && (j == my_int) ) {
                    ctx.strokeRect(x_offset - (width / 2) + increment, y_offset, 5, 5)
                    increment += 15;
                } else {
                    ctx.strokeRect(x_offset - (width / 2) - oppIncrement + width - 5, y_offset, 5, 5)
                    oppIncrement += 15;
                }
            }
        }
    }

}

function drawPotentialPaths() {
    if (paths.length > 0) {
        let current = null;
        if(hovIdx.length === 1) {
            current = camps[hovIdx[0]];
        } else if(hovIdx.length === 3) {
            current = state[hovIdx[0]][hovIdx[1]][hovIdx[2]];
        } 
        
        for (let i = 0; i < paths.length; i++) {
            if (
                current &&
                (paths[i][paths[i].length - 1][0] == current.Position[0]) &&
                (paths[i][paths[i].length - 1][1] == current.Position[1])
            ) {

                if(pathType[i] == 1) { drawComplexPath(paths[i], "rgba(255, 0, 0, 1)"); }
                else { drawComplexPath(paths[i], "#FFC107"); }

            } else {
                drawComplexPath(paths[i], "rgba(255, 255, 255, 0.3)");
            }
        }
    }
}

function drawPaths() {
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                drawComplexPath(state[i][j][k].Path, "rgba(255, 255, 255, 0.3)");
            }
        }
    }
}

function drawDynamicCircles() {
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                if(state[i][j][k].DestinationIndex.length > 1) {
                    ctx.beginPath();
                    ctx.arc(width + state[i][j][k].DynamicCircle[0], height - state[i][j][k].DynamicCircle[1], state[i][j][k].DistanceLeft/3, 0, 2 * Math.PI);
                    ctx.stroke();
                }
            }
        }
    }
}

function drawLastMileCircle() {
    ctx.strokeStyle = 'red';
    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {
                if(state[i][j][k].DestinationIndex.length > 1) {
                    let target = state[ state[i][j][k].DestinationIndex[0] ][ state[i][j][k].DestinationIndex[1] ][ state[i][j][k].DestinationIndex[2] ]
                    ctx.beginPath();
                    ctx.arc(width + target.Position[0], height - target.Position[1], state[i][j][k].LastMileRadius, 0, 2 * Math.PI);
                    ctx.stroke();
                }
            }
        }
    }
}

function drawCampData() {
    let x = 100;
    let y = 100;
    for(let i = 0; i < camps.length; i++) {
        ctx.fillStyle = 'white';
        ctx.fillText(i + " " + camps[i].Position + " " + camps[i].HP, x, y);
        y += 10
    }
}