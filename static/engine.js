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
	//ctx.clearRect(0,0,innerWidth,innerHeight);
	ctx.beginPath();
	ctx.arc(x,y,radius,0,Math.PI*2,false);
	ctx.strokeStyle='blue';
	ctx.stroke();
	
	if (x + radius > c.width || x - radius < 0) {
		dx = -dx;
	}
	if (y + radius > c.height || y - radius < 0) {
		dy = -dy;
	}
	x += dx;
	y += dy;
}
anime()
