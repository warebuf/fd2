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

function removeB1() {

	// send message to server to create match
	var e1 = document.getElementById("formatselect").value;
	var e2 = document.getElementById("opponentselect").value;
	socket.send(JSON.stringify({ "Event": 'createMatch', "Message": e1+e2 }))


	var button = document.getElementById("b1");
	button.remove();
	var menu1 = document.getElementById("opponent");
	menu1.remove();
	var menu2 = document.getElementById("format");
	menu2.remove();

	document.getElementById("test").insertAdjacentHTML('beforeend',"<button id='b2' onclick=removeB2()>Cancel Matchmaking</button>");

}

function removeB2() {
	var button = document.getElementById("b2");
	button.remove();
	document.getElementById("test").insertAdjacentHTML('beforeend',"<button id='b1' onclick=removeB1()>Connect to a Match</button>");

	document.getElementById("container").insertAdjacentHTML('beforeend',
		"<div id=\"format\">\n" +
		"\t\t<label>Choose a Format:</label>\n" +
		"\t\t<select id=\"formatselect\">\n" +
		"\t\t\t<option value=\"ffa\">Free For All</option>\n" +
		"\t\t\t<option value=\"tea\">Team</option>\n" +
		"\t\t\t<option value=\"1vx\">1vX</option>\n" +
		"\t\t</select>\n" +
		"\n" +
		"\t</div>");
	
	document.getElementById("container").insertAdjacentHTML('beforeend',
		"<div id=\"opponent\"> <label>Choose Number of Opponent:</label> <select id=\"opponentselect\"> " +
		"			<option value=\"1\">1</option>" +
		"			<option value=\"2\">2</option>" +
		"			<option value=\"3\">3</option>" +
		"			<option value=\"9\">9</option>" +
		"			<option value=\"99\">99</option>" +
			"</select>	</div>");
}

