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
	ctx.strokeStyle='rgb('+Math.floor(Math.random()*255)+','+Math.floor(Math.random()*255)+','+Math.floor(Math.random()*255)+')';
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

function removeB1() {

	// send message to server to create match
	var e1 = document.getElementById("formatselect").value;
	socket.send(JSON.stringify({ "Event": 'createPermissionList', "Message": 'ffa' }))

	var button = document.getElementById("b1");
	button.remove();
	var menu1 = document.getElementById("opponent");
	menu1.remove();
	var menu2 = document.getElementById("format");
	menu2.remove();

	document.getElementById("mid_div").insertAdjacentHTML('beforeend',"<button id='b2' onclick=removeB2()>Cancel Matchmaking</button>");

}

function removeB2() {
	var button = document.getElementById("b2");
	button.remove();
	document.getElementById("mid_div").insertAdjacentHTML('beforeend',"<button id='b1' onclick=removeB1()>Create a Match</button>");
}

