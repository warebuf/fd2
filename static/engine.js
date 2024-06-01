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


window.addEventListener('keydown', function (e) {
    if(e.key === 'Enter') {
        console.log("enter pressed")
    }

})


function anime() {
    window.requestAnimFrame(anime);
    ctx.fillStyle = '#696969';
    //ctx.clearRect(0,0,innerWidth,innerHeight);
    ctx.fillRect(0,0,innerWidth,innerHeight);

    let start = Date.now().toLocaleString('en-CH');
    ctx.font = '12px Arial';

    // print the current time
    ctx.fillStyle = 'black';
    ctx.fillRect((c.width/2)-(ctx.measureText(start).width/2),90,ctx.measureText(start).width,ctx.measureText('M').width);
    //ctx.fillRect((c.width/2)-(ctx.measureText(start).width/2),(c.height/2)-ctx.measureText('M').width,ctx.measureText(start).width,ctx.measureText('M').width);
    ctx.fillStyle = 'white';
    ctx.fillText(start,(c.width/2) - (ctx.measureText(start).width/2), 100);
    //ctx.fillText(start,(c.width/2) - (ctx.measureText(start).width/2), c.height/2);

    if(startCount != null) {
        remaining_time = startCount - start.replaceAll("’","")
        ctx.fillStyle = 'black';
        ctx.fillRect((c.width/2)-(ctx.measureText(remaining_time).width/2),10,ctx.measureText(remaining_time).width,ctx.measureText('M').width);
        ctx.font = '12px Arial';
        ctx.fillStyle = 'white';
        ctx.fillText(remaining_time,(c.width/2) - (ctx.measureText(remaining_time).width/2), 20);

        // need to change it so that it not only sends ur input when time expires, but sends it if you disconnect
        if(remaining_time<=0 && once) {
            socket.send(JSON.stringify({ "Event": 'timeUpMsg', "Message": "test" }))
            once = false
        }
    }

}
anime()

