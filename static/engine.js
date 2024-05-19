console.log("loaded js");

var c = document.getElementById("myCanvas");
const ctx = c.getContext("2d");

ctx.beginPath();
ctx.rect(20, 20, 150, 100);
ctx.stroke();


var x = 200;
function anime() {
	requestAnimationFrame(anime);
	c.beginPath();
	c.arc(x,200,30,0,Math.PI*2,false);
	c.strokeStyle='blue';
	c.stroke();
	x += 1;
	console.log("hi")
}
anime()
