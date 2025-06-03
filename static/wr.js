const canvas = document.getElementById("myCanvas");
const ctx = canvas.getContext("2d");

// Resize canvas to fill the screen
function resize() {
	canvas.width = window.innerWidth;
	canvas.height = window.innerHeight;
}
window.addEventListener("resize", resize);
resize();

function removeB1() {

// send message to server to create match
socket.send(JSON.stringify({ "Event": 'createPermissionList', "Message": 'ffa' }))

var button = document.getElementById("b1");
button.remove();

document.getElementById("mid_div").insertAdjacentHTML('beforeend',"<button id='b2' onclick=removeB2()>Cancel Matchmaking</button>");

}

function removeB2() {
var button = document.getElementById("b2");
button.remove();
document.getElementById("mid_div").insertAdjacentHTML('beforeend',"<button id='b1' onclick=removeB1()>Create a Match</button>");
}

