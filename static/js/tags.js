let filtered = [];
let resetBtn = document.getElementById("resetfilter");

let tags = e => e.title.split(" ");

function findThumbByTag(tag) {
	let matched = [];

	for (let i of document.getElementsByClassName("post"))
		if (tags(i).includes(tag)) matched.push(i);

	return matched;
}

function tagHasActive(tag) {
	for (let i of document.getElementsByClassName("post")) {
		if (!tags(i).includes(tag)) continue;
		if (i.style.display != "none") return true;
	}
	return false;
}

function updateTaglist() {
	let hasEnabled = filtered.length > 0;
	if (hasEnabled) resetBtn.style.display = "inline";
	else resetBtn.style.display = "none";

	for(let i of document.getElementsByClassName("tagname")) {
		if (!hasEnabled) {
			i.classList.remove("inactive");
			i.classList.remove("filter");
			continue;
		}

		let tag = i.dataset.tag;

		if (tagHasActive(tag)) {
			i.classList.remove("inactive");
			if (filtered.includes(tag))
				i.classList.add("filter");
			else
				i.classList.remove("filter");
		} else {
			i.classList.add("inactive");
			i.classList.remove("filter");
		}
	}
}

function updateFilter() {
	let hasEnabled = filtered.length > 0;
	for (let i of document.getElementsByClassName("post")) {
		if (!hasEnabled) { i.style.display = ""; continue; }

		let t = tags(i);
		i.style.display = filtered.every(e=>t.includes(e)) ? "" : "none";
	}
}

function filter(tag) {
	if (filtered.includes(tag)) return;
	filtered.push(tag);

	updateFilter();
	updateTaglist();
}

function unfilter(tag) {
	if (!filtered.includes(tag)) return;
	filtered = filtered.filter(e=>e!=tag);

	updateFilter();
	updateTaglist();
}

function resetFilter() {
	filtered = [];

	updateFilter();
	updateTaglist();
}

resetBtn.onclick = resetFilter;

for(let i of document.getElementsByClassName("tagname")) {
	let tag = i.dataset.tag;
	let matching = findThumbByTag(tag);

	i.onmouseenter = () => {
		matching.forEach(p=>p.classList.add("active"))
	};

	i.onclick = (e) => {
		e.preventDefault();
		if (i.classList.contains("filter"))
			unfilter(tag);
		else
			filter(tag);
	};

	i.onmouseleave = () => {
		matching.forEach(p=>p.classList.remove("active"))
	};
}
