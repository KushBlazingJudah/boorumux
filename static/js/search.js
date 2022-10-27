function addTag(name) {
	let tags = document.getElementById("q").value.split(" ").filter((e)=>e!=name)
	tags.push(name)
	document.getElementById("q").value = tags.join(" ")
}

function delTag(name) {
	document.getElementById("q").value = document.getElementById("q").value.split(" ").filter((e)=>e!=name).join(" ")
}
