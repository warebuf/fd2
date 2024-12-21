

import { Bezier } from "https://cdn.jsdelivr.net/npm/bezier-js@6.1.4/src/bezier.min.js";

document.onkeydown = function (e) {
    switch (e.key) {

        case 'E':
        case 'e':
        case "Enter":
            if (phase == "CHARACTER SELECTION") {

                if (second_index == false) {
                    second_index = true

                    screen_index_2 = screen_index
                    x_index_2 = x_index_1
                    if (screen_index == 1) {
                        if (x_index_1 == 0) {
                            y_index_2 = y_index_1 + h_index
                        } else if (x_index_1 == 1) {
                            y_index_2 = y_index_1 + l_index
                        } else if (x_index_1 == 2) {
                            y_index_2 = y_index_1 + r_index
                        } else if (x_index_1 == 3) {
                            y_index_2 = y_index_1 + b_index
                        }
                    } else {
                        y_index_2 = y_index_1
                    }

                } else if (second_index == true) {

                    if (x_index_1 == x_index_2) {

                        let y1 = ""
                        if (screen_index == 1) {
                            if (x_index_1 == 0) {
                                y1 = y_index_1 + h_index
                            } else if (x_index_1 == 1) {
                                y1 = y_index_1 + l_index
                            } else if (x_index_1 == 2) {
                                y1 = y_index_1 + r_index
                            } else if (x_index_1 == 3) {
                                y1 = y_index_1 + b_index
                            }

                        } else {
                            y1 = y_index_1
                        }

                        const s = screen_index + ';' + x_index_1 + ';' + y1 + ';' + screen_index_2 + ';' + x_index_2 + ';' + y_index_2
                        socket.send(JSON.stringify({"Event": "swapRequest", "Message": s}))
                        console.log("sent swapreq")
                    } else {
                        console.log("CANNOT SWAP INCOMPATIBLE BODY PARTS")
                        err_msg_index = 1
                    }

                    second_index = false
                }
            } else {
                console.log("enter")

                if (state_machine == 0) {
                    // if the current hoverable is a unit, continue
                    if (
                        (
                            (hoverable.Type == "unit") &&
                            (hoverable.I == my_team) &&
                            (hoverable.J == my_int) &&
                            (state[hoverable.I][hoverable.J][hoverable.K].Status == 0)
                        )
                        ||
                        (
                            (state[hoverable.I][hoverable.J][hoverable.K].Status == 1) &&
                            coordinateEqual(state[hoverable.I][hoverable.J][hoverable.K].Position, state[hoverable.I][hoverable.J][hoverable.K].Start)
                        )
                    ) {

                        findPath(state[my_team][my_int][selected_hoverable.K]);

                        selected_hoverable = hoverable;
                        state_machine = 1;
                        setupHoverable();
                    }
                } else if (state_machine == 1) {
                    // if the current hoverable is a POA, continue

                    socket.send(JSON.stringify({
                        "Event": 'act',
                        "Message": action_buffer,
                        "Team_index": my_team,
                        "Client_index": my_int,
                        "Hero_index": selected_hoverable.K,
                        Route: hoverable
                    }))
                    selected_hoverable = null;
                    state_machine = 0;
                    setupHoverable();
                }

            }
            break;

        case 'C':
        case 'c':
            if (view == 0) {
                view = 1
            } else if (view == 1) {
                view = 0
            }
            break;

        case 'A':
        case 'a':
        case "ArrowLeft":

            if (phase == "CHARACTER SELECTION") {

                x_index_1 = x_index_1 - 1
                if (x_index_1 < 0) {
                    x_index_1 = 3
                }
            } else if (match_data.length > 0) {
                calculateNextHoverable(1);
            }

            break;

        case 'D':
        case 'd':
        case "ArrowRight":

            if (phase == "CHARACTER SELECTION") {
                x_index_1 = x_index_1 + 1
                if (x_index_1 > 3) {
                    x_index_1 = 0
                }
            } else if (match_data.length > 0) {
                calculateNextHoverable(2);
            }

            break;

        case 'W':
        case 'w':
        case "ArrowUp":
            if (phase == "CHARACTER SELECTION") {

                if (screen_index == 1) {
                    if (x_index_1 == 0) { //h case
                        y_index_1 = y_index_1 - 1
                        if (y_index_1 < 0) {
                            if (h_index > 0) {
                                h_index = h_index - 1
                                y_index_1 = 0
                            } else {
                                screen_index = 0
                                y_index_1 = 4
                            }
                        }
                    } else if (x_index_1 == 1) {
                        y_index_1 = y_index_1 - 1
                        if (y_index_1 < 0) {
                            if (l_index > 0) {
                                l_index = l_index - 1
                                y_index_1 = 0
                            } else {
                                screen_index = 0
                                y_index_1 = 4
                            }
                        }
                    } else if (x_index_1 == 2) {
                        y_index_1 = y_index_1 - 1
                        if (y_index_1 < 0) {
                            if (r_index > 0) {
                                r_index = r_index - 1
                                y_index_1 = 0
                            } else {
                                screen_index = 0
                                y_index_1 = 4
                            }
                        }
                    } else if (x_index_1 == 3) {
                        y_index_1 = y_index_1 - 1
                        if (y_index_1 < 0) {
                            if (b_index > 0) {
                                b_index = b_index - 1
                                y_index_1 = 0
                            } else {
                                screen_index = 0
                                y_index_1 = 4
                            }
                        }
                    }
                } else {
                    y_index_1 = y_index_1 - 1
                    if (y_index_1 < 0) {
                        screen_index = 1
                        y_index_1 = 4
                    }
                }
            } else if (match_data.length > 0) {
                calculateNextHoverable(0);
            }
            break;

        case 'S':
        case 's':
        case "ArrowDown":
            if (phase == "CHARACTER SELECTION") {

            } else if (match_data.length > 0) {
                calculateNextHoverable(3);
            }
            break;

        case ' ':
            if (game_over == true) {
                returntomatchmaking()
            }
            break;

        case "Escape":
            if (settings_menu == 0) {
                settings_menu = 1
                document.getElementById("settings").insertAdjacentHTML('beforeend', "<button id='resign' onclick=resign()>Resign</button>");
            } else {
                settings_menu = 0
                let s = document.getElementById("settings").getElementsByTagName("button")
                s[0].remove()
            }
            break;

        case 'Q':
        case 'q':
        case "Backspace":
            if (state_machine == 1) {
                state_machine = 0;
                selected_hoverable = null;
                setupHoverable();
            }
            break;

        case 'b':
        case 'B':
            a2++;
            if (a2 == a1) {
                a2++;
            }

            if (a2 > intersections.length - 1) {
                a2 = 0;
                a1++;
                if (a1 > intersections.length - 1) {
                    a1 = 0;
                    a2 = 1;
                }
            }
            console.log(a1, a2, aStar(intersections[a1], intersections[a2]));
            const startTime = performance.now();
            path = aStar(intersections[a1], intersections[a2]);
            const endTime = performance.now();
            console.log(`Execution Time: ${endTime - startTime}ms`);
            break;

    }
}

var a1 = 0;
var a2 = 1;

// important data
var socket = null; // socket object for this tab
var my_team = -1; // the user's team index
var my_int = -1; // the user's player index

// state stuff
var animating_state = -1 // where are we at in the dataset
var phase = "null";
var turn = "null"; // the current phase/turn
var sent_time = null; // the time that the turn ends

var game_over = false;

var attacking_frames_count = 0

// the board state
var state = []
var camps = [];

var h_index = 0
var l_index = 0
var r_index = 0
var b_index = 0

var bench = null
var bench_index = 0

var screen_index = 0
var x_index_1 = 0
var y_index_1 = 0
var second_index = false
var screen_index_2 = 0
var x_index_2 = 0
var y_index_2 = 0

var err_msg_index = 0
var err_count = 0

// datasets
var match_data = new Array();
var atk_data = new Array();
var camp_data = new Array();
var cumu_data = new Array();

var time_queue = []
var turn_queue = []
var event_log = []

var damageNumbers = []; // UI related, holds the damage number animation

var frame_count = 0;
var state_machine = 0;
var even = 0; // just a small global variable to let us know when to flash camp icons
var action_buffer = ""

var hoverable = null;
var hoverable_coordinate = null;
var selected_hoverable = null;

var total_units_of_time = 0;

var view = 0;
var settings_menu = 0;

var obstacles = [];
var intersections = [];
var path = [];


function main() {

    if (!window["WebSocket"]) {
        alert("Error: Your browser does not support web sockets.")
    } else {
        socket = new WebSocket("ws://{{.host}}/msocket?matchid={{.mid}}");
        socket.onclose = function () {
            alert("Connection has been closed.");
        }

        socket.onmessage = function (e) {
            var receive_time = Date.now()
            var msg = eval('(' + e.data + ')');
            //console.log("msg_init", msg.Event, msg.Name, msg.MatchID, msg.Message, msg.TCH, msg.Turn, msg.Spawns, msg.POAs);
            console.log("msg received", msg.Event)

            var msg_date = msg.When.substring(0, 10);
            var msg_time = msg.When.substring(11, 19);

            if (msg.Event == 'clockSyncRequest') {
                console.log("dif", Date.parse(msg.When) - receive_time)
                socket.send(JSON.stringify({"Event": 'clockSyncResponse', "Message": receive_time.toString()}))
            } else if (msg.Event == 'entered') {
                event_log.push(msg_date + ' ' + msg_time + ' ' + msg.Name + ' has entered the game')
                if (event_log.length >= 25) {
                    event_log.shift()
                }
            } else if (msg.Event == 'left') {
                event_log.push(msg_date + ' ' + msg_time + ' ' + msg.Name + ' has left the game')
                if (event_log.length >= 25) {
                    event_log.shift()
                }
            } else if (msg.Event == 'assignInt') {
                let temp_string = msg.Message.split(";")
                my_team = parseInt(temp_string[0])
                my_int = parseInt(temp_string[1])
            } else if (msg.Event == 'ticker_start') {
                sent_time = Date.parse(msg.Message)
            } else if (msg.Event == 'bench') {
                bench = JSON.parse(msg.Message)
            } else if (msg.Event == 'update_phase') {
                if (phase == "null") {
                    phase = msg.Phase
                    if (phase == "CHARACTER SELECTION") {
                        document.getElementById("sel_but").insertAdjacentHTML('beforeend', "<button id='a1' onclick=endCharSelect()>finished</button>");
                    }
                } else if (phase == "CHARACTER SELECTION") {
                    var menu2 = document.getElementById("sel_but");
                    menu2.remove();
                    phase = msg.Phase
                }
            } else if (msg.Event == 'sel_finished') {
                event_log.push(msg_date + ' ' + msg_time + ' ' + msg.Name + ' finished unit selection.')
                if (event_log.length >= 25) {
                    event_log.shift()
                }
            } else if (msg.Event == 'obstacles') {
                const temp = msg.Message.split('\n')
                for (let i = 0; i < temp.length; i++) {
                    const temp = JSON.parse(temp[i]);
                    const curve = new Bezier(temp.X1, temp.Y1, temp.X2, temp.Y2, temp.X3, temp.Y3, temp.X4, temp.Y4);
                    console.log(i, curve)
                    obstacles.push(temp);
                }
            } else if (msg.Event == 'intersections') {
                const temp = msg.Message.split('\n')
                for (let i = 0; i < temp.length; i++) {
                    intersections.push(JSON.parse(temp[i]));
                }
            } else if (msg.Event == 'game_state') {
                console.log("received game_state")

                if (phase == "CHARACTER SELECTION") {
                    let temp = msg.TCH;
                    let md = [];
                    for (let i = 0; i < temp.length; i++) {
                        md[i] = new Array();
                        for (let j = 0; j < temp[i].length; j++) {
                            md[i][j] = new Array();
                            for (let k = 0; k < temp[i][j].length; k++) {
                                md[i][j][k] = JSON.parse(temp[i][j][k]);
                            }
                        }
                    }

                    let cdata = [];
                    for (let i = 0; i < msg.Camps.length; i++) {
                        cdata.push(JSON.parse(msg.Camps[i]))
                    }

                    state = md
                    camps = cdata

                } else {
                    turn_queue.push(msg.Turn)
                    time_queue.push(msg.UnitsOfTime)

                    // turn the JSON into javascript objects
                    let temp = msg.TCH;
                    let md = [];
                    for (let i = 0; i < temp.length; i++) {
                        md[i] = new Array();
                        for (let j = 0; j < temp[i].length; j++) {
                            md[i][j] = new Array();
                            for (let k = 0; k < temp[i][j].length; k++) {
                                md[i][j][k] = JSON.parse(temp[i][j][k]);
                            }
                        }
                    }

                    let cdata = [];
                    for (let i = 0; i < msg.Camps.length; i++) {
                        cdata.push(JSON.parse(msg.Camps[i]))
                    }


                    // this triggers if they refreshed the page
                    if (animating_state == -1) {
                        let temp_string = msg.Message.split(";")
                        my_team = parseInt(temp_string[0])
                        my_int = parseInt(temp_string[1])
                        state = md;
                        animating_state = 0
                        camps = cdata
                        setupHoverable()
                    }

                    match_data.push(md)
                    camp_data.push(cdata)

                    if (msg.Atk == null) {
                        atk_data.push([])
                    } else {
                        atk_data.push(msg.Atk)
                    }
                }
            } else if (msg.Event == 'game_over') {
                game_over = true
                document.getElementById("test").insertAdjacentHTML('beforeend', "<button id='b1' onclick=returntomatchmaking()>RETURN TO MATCHMAKING</button>");
            }
        }
    }
}

main()

function returntomatchmaking() {
    window.location.href = 'http://{{.host}}/waitroom'
}

function endCharSelect() {
    socket.send(JSON.stringify({"Event": 'endCharSel'}))
}

function resign() {
    socket.send(JSON.stringify({"Event": 'resign'}))
}

function setupHoverable() {

    let mostNorthActive = null;

    let closestCamp = null;
    let closestCampDist = Number.MAX_SAFE_INTEGER;

    if (state_machine == 0) {
        for (let i = 0; i < state[my_team][my_int].length; i++) {
            if (state[my_team][my_int][i].Status == 0) {
                if (mostNorthActive == null) {
                    mostNorthActive = state[my_team][my_int][i].Position
                } else if (mostNorthActive.Y < state[my_team][my_int][i].Position.Y) {
                    mostNorthActive = state[my_team][my_int][i].Position
                }
            }
        }
    } else if (state_machine == 1) {
        for (let i = 0; i < camps.length; i++) {
            if (camps[i].HP > 0) {
                let dist = getDistance(camps[i].XY, hoverable);
                if (closestCamp == null) {
                    closestCamp = camps[i].XY
                    closestCampDist = dist
                } else if (dist <= closestCampDist) {
                    if (dist == closestCampDist) {
                        if (camps[i].XY.Y < closestCamp.Y) {
                            closestCamp = camps[i].XY
                        }
                    } else {
                        closestCamp = camps[i].XY
                        closestCampDist = dist
                    }
                }
            }
        }
    }

    if (mostNorthActive != null) {
        hoverable = mostNorthActive;
    } else if (closestCamp != null) {
        hoverable = closestCamp;
    }

    if (hoverable != null) {
        if (hoverable.Type == "camp") {
            hoverable_coordinate = null;
        } else if (hoverable.Type == "unit") {
            hoverable_coordinate = state[hoverable.I][hoverable.J][hoverable.K]
        }
    }
}

function followHoverable() {
    if (hoverable_coordinate != null) {
        hoverable = hoverable_coordinate.Position
    }
}

function calculateNextHoverable(direction) {
    let hoverables = []

    if (state_machine == 0) {
        for (let i = 0; i < state.length; i++) {
            for (let j = 0; j < state[i].length; j++) {
                for (let k = 0; k < state[i][j].length; k++) {
                    hoverables.push(state[i][j][k].Position)
                }
            }
        }
    }

    if (camps.length > 0) {
        for (let i = 0; i < camps.length; i++) {
            hoverables.push(camps[i].XY)
        }
    }

    // next figure out the U,L,R,D
    let closestCardinalCoordinate = null;
    let closestCardinalDistance = Number.MAX_SAFE_INTEGER;

    let closestWrapCoordinate = null; // this handles wrapping
    let closestWrapDistance = Number.MAX_SAFE_INTEGER;

    let nextSameCoordinate = null;

    for (let i = 0; i < hoverables.length; i++) {
        let dist = getDistance(hoverable, hoverables[i])
        if (dist == 0) {
            if ((direction == 2) || (direction == 3)) {
                if (isFirstCoordinateGreater(hoverables[i], hoverable)) {
                    if (nextSameCoordinate == null) {
                        nextSameCoordinate = hoverables[i];
                    } else {
                        if (isFirstCoordinateGreater(nextSameCoordinate, hoverables[i])) {
                            nextSameCoordinate = hoverables[i];
                        }
                    }
                }
            } else {
                if (isFirstCoordinateGreater(hoverable, hoverables[i])) {
                    if (nextSameCoordinate == null) {
                        nextSameCoordinate = hoverables[i];
                    } else {
                        if (isFirstCoordinateGreater(hoverables[i], nextSameCoordinate)) {
                            nextSameCoordinate = hoverables[i];
                        }
                    }
                }
            }
        } else if (direction == 0) {
            if (hoverables[i].Y > hoverable.Y) {
                let dist = getDistance(hoverable, hoverables[i])
                if (closestCardinalDistance > dist) {
                    closestCardinalCoordinate = hoverables[i];
                    closestCardinalDistance = dist;
                } else if (closestCardinalDistance == dist) {
                    if (isFirstCoordinateGreater(hoverables[i], closestCardinalCoordinate)) {
                        closestCardinalCoordinate = hoverables[i];
                    }
                }
            }
        } else if (direction == 1) {
            if (hoverable.X > hoverables[i].X) {
                let dist = getDistance(hoverable, hoverables[i])
                if (closestCardinalDistance > dist) {
                    closestCardinalCoordinate = hoverables[i];
                    closestCardinalDistance = dist;
                } else if (closestCardinalDistance == dist) {
                    if (isFirstCoordinateGreater(hoverables[i], closestCardinalCoordinate)) {
                        closestCardinalCoordinate = hoverables[i];
                    }
                }
            }
        } else if (direction == 2) {
            if (hoverable.X < hoverables[i].X) {
                let dist = getDistance(hoverable, hoverables[i])
                if (closestCardinalDistance > dist) {
                    closestCardinalCoordinate = hoverables[i];
                    closestCardinalDistance = dist;
                } else if (closestCardinalDistance == dist) {
                    if (isFirstCoordinateGreater(closestCardinalCoordinate, hoverables[i])) {
                        closestCardinalCoordinate = hoverables[i];
                    }
                }
            }
        } else if (direction == 3) {
            if (hoverables[i].Y < hoverable.Y) {
                let dist = getDistance(hoverable, hoverables[i])
                if (closestCardinalDistance > dist) {
                    closestCardinalCoordinate = hoverables[i];
                    closestCardinalDistance = dist;
                } else if (closestCardinalDistance == dist) {
                    if (isFirstCoordinateGreater(closestCardinalCoordinate, hoverables[i])) {
                        closestCardinalCoordinate = hoverables[i];
                    }
                }
            }
        }

        if ((closestCardinalCoordinate == null) && (dist != 0)) {
            let wrap_x = 0;
            let wrap_y = 0;
            if (direction == 0) {
                wrap_y = -200;
            } else if (direction == 1) {
                wrap_x = 200;
            } else if (direction == 2) {
                wrap_x = -200;
            } else if (direction == 3) {
                wrap_y = 200;
            }

            let wrap = structuredClone(hoverable); // deep copy of coordinate
            wrap.X = wrap.X + wrap_x;
            wrap.Y = wrap.Y + wrap_y;

            let wrapDist = getDistance(wrap, hoverables[i]);
            if (closestWrapDistance > wrapDist) {
                closestWrapCoordinate = hoverables[i];
                closestWrapDistance = wrapDist;
            } else if (closestWrapDistance == wrapDist) {
                if ((direction == 3) || (direction == 2)) {
                    if (isFirstCoordinateGreater(closestWrapCoordinate, hoverables[i])) {
                        closestWrapCoordinate = hoverables[i];
                    }
                } else {
                    if (isFirstCoordinateGreater(hoverables[i], closestWrapCoordinate)) {
                        closestWrapCoordinate = hoverables[i];
                    }
                }
            }
        }
    }

    if (nextSameCoordinate != null) {
        hoverable = nextSameCoordinate;
    } else if (closestCardinalCoordinate != null) {
        hoverable = closestCardinalCoordinate;
    } else if (closestWrapCoordinate != null) {
        hoverable = closestWrapCoordinate;
    }

    if (hoverable != null) {
        if (hoverable.Type == "camp") {
            hoverable_coordinate = null;
        } else if (hoverable.Type == "unit") {
            hoverable_coordinate = state[hoverable.I][hoverable.J][hoverable.K]
        }
    }

}

function isFirstCoordinateGreater(c1, c2) {
    if ((c1.Type == "camp") && (c2.Type == "unit")) {
        return true;
    } else if ((c1.Type == "unit") && (c2.Type == "unit")) {
        if (c2.I < c1.I) {
            return true;
        } else if (c1.I == c2.I) {
            if (c2.J < c1.J) {
                return true;
            } else if (c2.J == c1.J) {
                if (c2.K < c1.K) {
                    return true;
                }
            }
        }
    } else if ((c1.Type == "camp") && (c2.Type == "camp")) {
        if (c2.I < c1.I) {
            return true;
        }
    }
    return false;
}


function findPath(unit) {

    console.log(unit.Position)

}

function lineIntersectsBezier(P1, P2, B) {

    var line = {p1: {x: P1.X, y: P1.Y}, p2: {x: P2.X, y: P2.Y}};

    const curve = new Bezier(B.X1, B.Y1, B.X2, B.Y2, B.X3, B.Y3, B.X4, B.Y4);

    const intersections = curve.intersects(line);

    if (intersections.length > 0) {
        return true
    }
    return false
}












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
        drawCampSelect()
        calcPos()
        drawAttack()
        drawPos()
        drawHoverable()
        drawCamps()
        drawPotentialPaths()
        drawPaths()
        //drawStats()
        drawDamageNumbers()
        updateDamageNumbers()
        drawTurn()
        drawUnitsOfTime()
        drawLog()
        if(view==0){
            drawHeatMap()
            drawCampHP()
        }
        if(path != []) {
            drawComplexPath(path)
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
                            state[i][j][k].Start = state[i][j][k].Position
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].Angle = match_data[animating_state][i][j][k].Angle;
                            state[i][j][k].DistanceLeft = getDistance(state[i][j][k].Position, state[i][j][k].Destination);
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
                        if(state[i][j][k].Destination != match_data[animating_state][i][j][k].Destination) {
                            state[i][j][k].Start = state[i][j][k].Position
                            state[i][j][k].Destination = match_data[animating_state][i][j][k].Destination;
                            state[i][j][k].Angle = match_data[animating_state][i][j][k].Angle;
                            state[i][j][k].DistanceLeft = getDistance(state[i][j][k].Position, state[i][j][k].Destination);
                        }
                        if(state[i][j][k].DistanceLeft != match_data[animating_state][i][j][k].DistanceLeft) {
                            all_distances_correct = false;
                            let increment = getXY(state[i][j][k].Angle, state[i][j][k].SPD * 0.015)
                            state[i][j][k].Position.X = state[i][j][k].Position.X + increment[0]
                            state[i][j][k].Position.Y = state[i][j][k].Position.Y + increment[1]
                            state[i][j][k].DistanceLeft = state[i][j][k].DistanceLeft - (state[i][j][k].SPD * 0.015)
                            if( state[i][j][k].DistanceLeft <= match_data[animating_state][i][j][k].DistanceLeft ){
                                state[i][j][k].Position = structuredClone(match_data[animating_state][i][j][k].Position)
                                state[i][j][k].DistanceLeft = match_data[animating_state][i][j][k].DistanceLeft
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
                    drawRotated('>',width+state[i][j][k].Position.X,height-state[i][j][k].Position.Y,state[i][j][k].Angle, "11", 3, "white")
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
    }

    ctx.fillStyle = "white";
}

function drawCampSelect() {
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

function drawPotentialPaths() {

    if(state_machine==0) {

    }
    else if(state_machine==1) {

        if((selected_hoverable!=null) && (selected_hoverable.Type == "unit") && (selected_hoverable.I == my_team) && (selected_hoverable.J == my_int)) {
            for(let i = 0; i < camps.length; i++) {
                if(camps[i].HP > 0) {
                    let colour = ""
                    ctx.beginPath();
                    if(coordinateEqual(camps[i].XY,hoverable)) {
                        colour = "white";
                    }
                    else {
                        colour = "grey";
                    }
                    ctx.strokeStyle = colour;
                    ctx.moveTo(width+selected_hoverable.X, height-selected_hoverable.Y);
                    ctx.lineTo(width+camps[i].XY.X, height-camps[i].XY.Y);
                    ctx.stroke();

                    // write distance
                    let dist = getDistance(selected_hoverable, camps[i].XY).toFixed(1);
                    let mid = getCenter(selected_hoverable, camps[i].XY);
                    let radians = 0;
                    if(selected_hoverable.X<=camps[i].XY.X) {
                        radians = getAngle(selected_hoverable, camps[i].XY)
                    } else {
                        radians = getAngle(camps[i].XY, selected_hoverable)
                    }

                    drawRotated(dist,width+mid[0],height-mid[1],radians, "8", 10, colour);

                    // write TTL, may not do this, not sure

                }


            }
        }
    }

}

function drawPaths() {

    for(let i = 0; i < state.length; i++) {
        for(let j = 0; j < state[i].length; j++) {
            for(let k = 0; k < state[i][j].length; k++) {

                if(state[i][j][k].HP>0 && state[i][j][k].State == 0) {
                    ctx.beginPath();
                    ctx.strokeStyle = "grey";
                    ctx.moveTo(width+state[i][j][k].Start.X, height-state[i][j][k].Start.Y);
                    ctx.lineTo(width+state[i][j][k].Destination.X, height-state[i][j][k].Destination.Y);
                    ctx.stroke();
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

                    if(state[j][k][i].HP<=0){ctx.fillStyle='grey'}
                    ctx.textAlign = "center";
                    ctx.fillRect(x_offset-1-45+(user_shift*user_count),y_offset - (0 | state[j][k][i].HP/2) + (i*unit_shift),3,0 | state[j][k][i].HP/2);
                    //ctx.fillRect(x_offset-1-30+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.HP/2) + (i*unit_shift),3,0 | state[j][k][i].L.HP/2);
                    //ctx.fillRect(x_offset-1-15+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.HP/2) + (i*unit_shift),3,0 | state[j][k][i].R.HP/2);
                    //ctx.fillRect(x_offset-1-0+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.HP/2) + (i*unit_shift),3,0 | state[j][k][i].B.HP/2);
                    ctx.fillText("HP",x_offset-22.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].HP,x_offset-45+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].HP/2))
                    //ctx.fillText(state[j][k][i].L.HP,x_offset-30+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.HP/2))
                    //ctx.fillText(state[j][k][i].R.HP,x_offset-15+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.HP/2))
                    //ctx.fillText(state[j][k][i].B.HP,x_offset-0+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.HP/2))
                    ctx.fillText("H",x_offset-45+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("L",x_offset-30+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("R",x_offset-15+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("B",x_offset-0+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+25+(user_shift*user_count),y_offset - (0 | state[j][k][i].ATK/2) + (i*unit_shift),3,0 | state[j][k][i].ATK/2);
                    //ctx.fillRect(x_offset-1+40+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].L.ATK/2);
                    //ctx.fillRect(x_offset-1+55+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].R.ATK/2);
                    //ctx.fillRect(x_offset-1+70+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.ATK/2) + (i*unit_shift),3,0 | state[j][k][i].B.ATK/2);
                    //ctx.fillText("ATK",x_offset+47.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].ATK,x_offset+25+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].ATK/2))
                    //ctx.fillText(state[j][k][i].L.ATK,x_offset+40+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.ATK/2))
                    //ctx.fillText(state[j][k][i].R.ATK,x_offset+55+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.ATK/2))
                    //ctx.fillText(state[j][k][i].B.ATK,x_offset+70+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.ATK/2))
                    ctx.fillText("H",x_offset+25+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("L",x_offset+40+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("R",x_offset+55+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("B",x_offset+70+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    ctx.fillRect(x_offset-1+95+(user_shift*user_count),y_offset - (0 | state[j][k][i].DEF/2) + (i*unit_shift),3,0 | state[j][k][i].DEF/2);
                    //ctx.fillRect(x_offset-1+110+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].L.DEF/2);
                    //ctx.fillRect(x_offset-1+125+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].R.DEF/2);
                    //ctx.fillRect(x_offset-1+140+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.DEF/2) + (i*unit_shift),3,0 | state[j][k][i].B.DEF/2);
                    //ctx.fillText("DEF",x_offset+117.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].DEF,x_offset+95+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].DEF/2))
                    //ctx.fillText(state[j][k][i].L.DEF,x_offset+110+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.DEF/2))
                    //ctx.fillText(state[j][k][i].R.DEF,x_offset+125+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.DEF/2))
                    //ctx.fillText(state[j][k][i].B.DEF,x_offset+140+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.DEF/2))
                    ctx.fillText("H",x_offset+95+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("L",x_offset+110+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("R",x_offset+125+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("B",x_offset+140+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    /*
                    //ctx.fillRect(x_offset-1+165+(user_shift*user_count),y_offset - (0 | state[j][k][i].H.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].H.ACC/2);
                    //ctx.fillRect(x_offset-1+180+(user_shift*user_count),y_offset - (0 | state[j][k][i].L.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].L.ACC/2);
                    //ctx.fillRect(x_offset-1+195+(user_shift*user_count),y_offset - (0 | state[j][k][i].R.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].R.ACC/2);
                    //ctx.fillRect(x_offset-1+210+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.ACC/2) + (i*unit_shift),3,0 | state[j][k][i].B.ACC/2);
                    //ctx.fillText("ACC",x_offset+187.5+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    //ctx.font = '9px monospace';
                    //ctx.fillText(state[j][k][i].H.ACC,x_offset+165+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].H.ACC/2))
                    //ctx.fillText(state[j][k][i].L.ACC,x_offset+180+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].L.ACC/2))
                    //ctx.fillText(state[j][k][i].R.ACC,x_offset+195+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].R.ACC/2))
                    //ctx.fillText(state[j][k][i].B.ACC,x_offset+210+(user_shift*user_count), y_offset-10 + (i*unit_shift) - (0 | state[j][k][i].B.ACC/2))
                    //ctx.fillText("H",x_offset+165+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("L",x_offset+180+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("R",x_offset+195+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.fillText("B",x_offset+210+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.font = '11px monospace';

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
                    */

                    ctx.fillRect(x_offset-1+515+(user_shift*user_count),y_offset - (0 | state[j][k][i].SPD/2) + (i*unit_shift),3,0 | state[j][k][i].SPD/2);
                    ctx.fillText("SPD",x_offset+515+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    ctx.font = '9px monospace';
                    ctx.fillText(state[j][k][i].SPD,x_offset+515+(user_shift*user_count), y_offset - 10 + (i*unit_shift) - (0 | state[j][k][i].SPD/2) )
                    ctx.fillText("B",x_offset+515+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    ctx.font = '11px monospace';

                    //ctx.fillRect(x_offset-1+540+(user_shift*user_count),y_offset - (0 | state[j][k][i].B.DOG/2) + (i*unit_shift),3,0 | state[j][k][i].B.DOG/2);
                    //ctx.fillText("DOG",x_offset+540+(user_shift*user_count), y_offset+25 + (i*unit_shift) )
                    //ctx.font = '9px monospace';
                    //ctx.fillText(state[j][k][i].B.DOG,x_offset+540+(user_shift*user_count), y_offset - 10 + (i*unit_shift) - (0 | state[j][k][i].B.DOG/2) )
                    //ctx.fillText("B",x_offset+540+(user_shift*user_count), y_offset+10 + (i*unit_shift) )
                    //ctx.font = '11px monospace';

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

function drawCampHP() {
    let x = 100
    let y = 700
    ctx.font = '11px monospace';

    for(let i = 0; i < camps.length; i++) {
        ctx.fillStyle = 'white';
        if(hoverable != null) {
            if((hoverable.Type=='camp') && (hoverable.I==i)) {
                ctx.fillStyle = 'red';
            }
        }
        ctx.fillText(i+": " + camps[i].HP, x, y+(i*10))
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

function drawTarget(x,y){

    ctx.strokeStyle = "red";
    ctx.strokeRect(x-10,y-10, 20,20);

}

function laser(x1,y1,x2,y2) {
    ctx.beginPath();
    ctx.strokeStyle = "yellow";
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.stroke();
}

function coordinateEqual(c1,c2) {
    if ((c1.X == c2.X) && (c1.Y == c2.Y)) {
        return true
    }
    return false
}

function getXY(theta, r) {
    return [r*Math.cos(theta), r*Math.sin(theta)]
}

function drawRotated(s,x,y,radians, size, y_shift, colour) {
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

function getCenter(c1,c2) {
    let x = (c1.X + c2.X)/2;
    let y = (c1.Y + c2.Y)/2;
    return [x,y]
}

function getAngle(c1,c2) {
    let y = c2.Y - c1.Y
    let x = c2.X - c1.X
    let theta = Math.atan2(y, x)
    if (theta < 0) {
        theta = theta + (2 * Math.PI)
    }
    return theta
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

    for(let i = 0; i < obstacles.length; i++ ){

        if(obstacles[i].Shape == "BEZIER") {
            ctx.beginPath();
            ctx.moveTo(width+obstacles[i].X1, height-obstacles[i].Y1);
            ctx.bezierCurveTo(
                width+obstacles[i].X2, height-obstacles[i].Y2,
                width+obstacles[i].X3, height-obstacles[i].Y3,
                width+obstacles[i].X4, height-obstacles[i].Y4
            );
            ctx.strokeStyle = "white";
            ctx.stroke();

        } else if([i].Shape == "CIRCLE") {

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

            const neighbor = intersections[ parseInt(currentNode.origin.Edges[i].slice(1), 10) ];

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

function drawComplexPath(path) {
    for (let i = 0; i < path.length - 1; i++) {
        ctx.beginPath();    // Begin a new path
        ctx.strokeStyle = "blue";
        ctx.moveTo(width+path[i][0], height - path[i][1]); // Move to the starting point
        ctx.lineTo(width+path[i + 1][0], height - path[i + 1][1]); // Draw a line to the ending point
        ctx.stroke();       // Render the line
    }
}


