for(let i of document.getElementsByClassName("tagname")) {
	let tag = i.dataset.tag;
	let matching = []
	for (let j of document.getElementsByClassName("post")) {
		if (j.title.split(" ").filter(e=>e==tag).length > 0) {
			matching.push(j)
		}
	}

	i.onmouseenter = () => {
		matching.forEach(p=>p.classList.add("active"))
	};

	i.onmouseleave = () => {
		matching.forEach(p=>p.classList.remove("active"))
	};
}
