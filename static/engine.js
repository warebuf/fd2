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



window.addEventListener('keydown', function (e) {
    if(e.key === 'Enter') {
        console.log("enter pressed")
        socket.send(JSON.stringify({ "Event": 'taptap', "Message": "taptap" }))
    }

})

var once = true

function anime() {
    window.requestAnimFrame(anime);
    ctx.fillStyle = '#303030';
    //ctx.clearRect(0,0,innerWidth,innerHeight);
    ctx.fillRect(0,0,innerWidth,innerHeight);

    let start = Date.now().toLocaleString('en-CH');
    ctx.font = '12px Arial';

    // print the current time
    ctx.textAlign = "left";
    ctx.fillStyle = 'black';
    ctx.fillRect((c.width/2)-(ctx.measureText(start).width/2),10,ctx.measureText(start).width,ctx.measureText('M').width);
    //ctx.fillRect((c.width/2)-(ctx.measureText(start).width/2),(c.height/2)-ctx.measureText('M').width,ctx.measureText(start).width,ctx.measureText('M').width);
    ctx.fillStyle = 'white';
    ctx.fillText(start,(c.width/2) - (ctx.measureText(start).width/2), 20);
    //ctx.fillText(start,(c.width/2) - (ctx.measureText(start).width/2), c.height/2);

    if(startCount != null) {
        remaining_time = startCount - start.replaceAll("’","")
        ctx.fillStyle = 'black';
        ctx.fillRect((c.width/2)-(ctx.measureText(remaining_time).width/2),30,ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '12px Arial';
        ctx.fillStyle = 'white';
        ctx.fillText(remaining_time,(c.width/2) - (ctx.measureText(remaining_time).width/2), 40);

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
                    ctx.fillText(road, 1100, ((i+j)*115) + 82 + (k*20));
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
                ctx.fillRect(1075, ((i+j)*115) + 71 + (k*20), 5, 5);
                if (true) {
                    c.fillStyle = 'green';
                } else {
                    c.fillStyle = 'red';
                }
                ctx.fillRect(1070, ((i+j)*115) + 76 + (k*20), 5, 7); // make it 270,5,7 OR 269,6,4
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    ctx.fillStyle = 'red';
                }
                ctx.fillRect(1080, ((i+j)*115) + 76 + (k*20), 5, 7);
                if (true) {
                    ctx.fillStyle = 'green';
                } else {
                    c.fillStyle = 'red';
                }
                ctx.fillRect(1075, ((i+j)*115) + 81 + (k*20), 5, 7);
            }
        }
    }

}


function drawStats() {
    ctx.fillStyle = 'white'
    ctx.textAlign = "center";

    ctx.fillText("T C H D", 10, 60);
    ctx.fillText("POS", 50, 60);
    ctx.fillText("ACT", 75, 60);

    ctx.fillText("HP", 110, 60) // H HP
    ctx.fillText("ATK", 135, 60) // H ATK
    ctx.fillText("DEF", 160, 60) // H DEF
    ctx.fillText("ACC", 185, 60) // H ACC
    ctx.fillText("CRT", 210, 60) // H CRT
    ctx.fillText("MOB", 235, 60) // H MOB
    ctx.fillText("CD", 260, 60) // H CD
    ctx.fillText("CLU", 285, 60) // H CLU
    ctx.fillText("USE", 310, 60) // H USE
    ctx.fillText("WGH", 335, 60) // H WEIGHT

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
    ctx.fillText("SPE", 935, 60) // LEG SPE
    ctx.fillText("ACL", 960, 60) // LEG CRIT
    ctx.fillText("ANT", 985, 60) // LEG MOBI
    ctx.fillText("END", 1010, 60) // LEG CD

    for(let i = 0; i < match_data[current_state].length; i++) {
        for(let j = 0; j < match_data[current_state][i].length; j++) {
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                ctx.fillText(i + " " +j + " " + k + " " + match_data[current_state][i][j][k].Direction, 10, ((i+j)*115) + 82 + (k*20));
                ctx.fillText(match_data[current_state][i][j][k].Position, 50, ((i+j)*115) + 82 + (k*20))
                ctx.fillText("-", 75, ((i+j)*115) + 82 + (k*20))


                ctx.fillText("100", 110, ((i+j)*115) + 82 + (k*20)) // H HP
                ctx.fillText("100", 135, ((i+j)*115) + 82 + (k*20)) // H ATK
                ctx.fillText("100", 160, ((i+j)*115) + 82 + (k*20)) // H DEF
                ctx.fillText("100", 185, ((i+j)*115) + 82 + (k*20)) // H ACC
                ctx.fillText("100", 210, ((i+j)*115) + 82 + (k*20)) // H CRIT
                ctx.fillText("100", 235, ((i+j)*115) + 82 + (k*20)) // H MOBI
                ctx.fillText("100", 260, ((i+j)*115) + 82 + (k*20)) // H CD
                ctx.fillText("100", 285, ((i+j)*115) + 82 + (k*20)) // H CLU
                ctx.fillText("0/0", 310, ((i+j)*115) + 82 + (k*20)) // H USE
                ctx.fillText("100", 335, ((i+j)*115) + 82 + (k*20)) // H WEIGHT

                ctx.fillText("L", 375, ((i+j)*115) + 77 + (k*20)) // L L
                ctx.fillText("99", 400, ((i+j)*115) + 77 + (k*20)) // L HP
                ctx.fillText("9", 425, ((i+j)*115) + 77 + (k*20)) // L ATK
                ctx.fillText("99", 450, ((i+j)*115) + 77 + (k*20)) // L DEF
                ctx.fillText("9", 475, ((i+j)*115) + 77 + (k*20)) // L ACC
                ctx.fillText("99", 500, ((i+j)*115) + 77 + (k*20)) // L CRIT
                ctx.fillText("9", 525, ((i+j)*115) + 77 + (k*20)) // L MOBI
                ctx.fillText("99", 550, ((i+j)*115) + 77 + (k*20)) // L CD
                ctx.fillText("9", 575, ((i+j)*115) + 77 + (k*20)) // L CLU
                ctx.fillText("0/0", 600, ((i+j)*115) + 77 + (k*20)) // L USE
                ctx.fillText("100", 625, ((i+j)*115) + 77 + (k*20)) // L WEIGHT

                ctx.fillText("R", 375, ((i+j)*115) + 87 + (k*20))   // R L
                ctx.fillText("100", 400, ((i+j)*115) + 87 + (k*20)) // R HP
                ctx.fillText("100", 425, ((i+j)*115) + 87 + (k*20)) // R ATK
                ctx.fillText("100", 450, ((i+j)*115) + 87 + (k*20)) // R DEF
                ctx.fillText("100", 475, ((i+j)*115) + 87 + (k*20)) // R ACC
                ctx.fillText("100", 500, ((i+j)*115) + 87 + (k*20)) // R CRIT
                ctx.fillText("100", 525, ((i+j)*115) + 87 + (k*20)) // R MOBI
                ctx.fillText("100", 550, ((i+j)*115) + 87 + (k*20)) // R CD
                ctx.fillText("100", 575, ((i+j)*115) + 87 + (k*20)) // R CLU
                ctx.fillText("0/0", 600, ((i+j)*115) + 87 + (k*20)) // R USE
                ctx.fillText("100", 625, ((i+j)*115) + 87 + (k*20)) // R WEIGHT


                ctx.fillText("100", 660, ((i+j)*115) + 82 + (k*20)) // LEG HP
                ctx.fillText("100", 685, ((i+j)*115) + 82 + (k*20)) // LEG ATK
                ctx.fillText("100", 710, ((i+j)*115) + 82 + (k*20)) // LEG DEF
                ctx.fillText("100", 735, ((i+j)*115) + 82 + (k*20)) // LEG ACC
                ctx.fillText("100", 760, ((i+j)*115) + 82 + (k*20)) // LEG CRIT
                ctx.fillText("100", 785, ((i+j)*115) + 82 + (k*20)) // LEG MOBI
                ctx.fillText("100", 810, ((i+j)*115) + 82 + (k*20)) // LEG CD
                ctx.fillText("100", 835, ((i+j)*115) + 82 + (k*20)) // LEG  CLU
                ctx.fillText("0/0", 860, ((i+j)*115) + 82 + (k*20)) // LEG USE
                ctx.fillText("100", 885, ((i+j)*115) + 82 + (k*20)) // LEG WEIGHT
                ctx.fillText("100", 910, ((i+j)*115) + 82 + (k*20)) // LEG DOG
                ctx.fillText("100", 935, ((i+j)*115) + 82 + (k*20)) // LEG SPE
                ctx.fillText("100", 960, ((i+j)*115) + 82 + (k*20)) // LEG CRIT
                ctx.fillText("100", 985, ((i+j)*115) + 82 + (k*20)) // LEG MOBI
                ctx.fillText("100", 1010, ((i+j)*115) + 82 + (k*20)) // LEG CD
                
            }
        }
    }
}