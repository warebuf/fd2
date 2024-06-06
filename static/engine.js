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
    }

}
anime()

function drawPos() {
    ctx.font = '11px monospace';
    ctx.fillStyle = 'white';
    for(let i = 0; i < match_data[current_state].length; i++) { // team
        for(let j = 0; j < match_data[current_state][i].length; j++) { // client
            for(let k = 0; k < match_data[current_state][i][j].length; k++) {
                if(match_data[current_state][i][j][k].Health > 0) {
                    let road = 'o' + "-".repeat(99) + 'x';
                    road = road.substring(0,
                        Math.round(match_data[current_state][i][j][k].Position)) +
                        (match_data[current_state][i][j][k].Direction ? '>' : '<') +
                        road.substring(Math.round(match_data[current_state][i][j][k].Position) + 1);
                    ctx.fillText(road, 250, ((i+j)*600) + 50 + (k*10));
                }
            }

        }
    }
}