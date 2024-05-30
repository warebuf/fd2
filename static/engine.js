console.log("loaded js");

var c = document.getElementById("myCanvas");
const ctx = c.getContext("2d");

//c.width = window.innerWidth;
//c.height = window.innerHeight;

c.width = document.body.clientWidth;
c.height = document.body.clientHeight;

window.requestAnimationFrame = window.requestAnimationFrame || window.mozRequestAnimationFrame || window.webkitRequestAnimationFrame || function(f){return setTimeout(f, 1000/60)};
window.cancelAnimationFrame = window.cancelAnimationFrame || window.mozCancelAnimationFrame || function(requestID){clearTimeout(requestID)};


var x = 200;
var y = 200;
var dx = 4;
var dy = 4;
var radius = 30;
function anime() {
    window.requestAnimationFrame(anime);
    ctx.clearRect(0,0,innerWidth,innerHeight);

    let start = Date.now().toLocaleString('en-CH');
    c.fillStyle = 'black';
    c.fillRect((canvas.width/2)-(c.measureText(start).width/2),(canvas.height/2)-c.measureText('M').width,c.measureText(start).width,c.measureText('M').width);
    c.fillStyle = 'white';
    c.fillText(start,(canvas.width/2) - (c.measureText(start).width/2), canvas.height/2);

}
anime()


