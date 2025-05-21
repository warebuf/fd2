const canvas = document.getElementById("myCanvas");
const ctx = canvas.getContext("2d");

// Resize canvas to fill the screen
function resize() {
	canvas.width = window.innerWidth;
	canvas.height = window.innerHeight;
	}
	window.addEventListener("resize", resize);
	resize();

	function drawRandomNumber() {
	const x = Math.random() * canvas.width;
	const y = Math.random() * canvas.height;
	const fontSize = Math.random() * 40 + 10;
	const value = Math.floor(Math.random() * 1000);

	ctx.font = `${fontSize}px monospace`;
	ctx.fillStyle = `hsl(${Math.random() * 360}, 100%, 70%)`;
	ctx.fillText(value, x, y);
}

function animate() {
	// fade out effect
	ctx.fillStyle = "rgba(0, 0, 0, 0.1)";
	ctx.fillRect(0, 0, canvas.width, canvas.height);

	// draw multiple random numbers per frame
	for (let i = 0; i < 1; i++) {
		drawRandomNumber();
	}

requestAnimationFrame(animate);
}

animate();

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

