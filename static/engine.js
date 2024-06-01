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

    // try printing start time
    if(startCount != null) {
        ctx.fillStyle = 'black';
        ctx.fillRect((c.width/2)-(ctx.measureText(start).width/2),10,ctx.measureText(start).width,ctx.measureText('M').width);
        ctx.font = '12px Arial';
        ctx.fillStyle = 'white';
        ctx.fillText(startCount,(c.width/2) - (ctx.measureText(start).width/2), 20);
        console.log(0, start)
        console.log(1, start.replaceAll("’",""))
        

    }

}
anime()


