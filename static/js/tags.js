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

	for(let i of document.getElementsByClassName("tagname")) {
		let tag = i.dataset.tag;

		if (tagHasActive(tag) || !hasEnabled) {
			i.classList.remove("inactive");
			if (filtered.includes(tag) && hasEnabled)
				i.classList.add("filter");
			else
				i.classList.remove("filter");
		} else {
			i.classList.add("inactive");
			i.classList.remove("filter");
		}
	}
}

function filter(tag) {
	if (filtered.includes(tag)) return;
	filtered.push(tag);

	let exclude = findThumbByTag(tag);

	for (let i of document.getElementsByClassName("post")) {
		if (!exclude.includes(i)) i.style.display = "none";
	}

	updateTaglist();
}

function resetFilter() {
	filtered = [];
	for (let i of document.getElementsByClassName("post")) {
		i.style.display = "";
	}

	updateTaglist();
}

resetBtn.onclick = (e) => {
	resetFilter();
	resetBtn.style.display = "none";
};

for(let i of document.getElementsByClassName("tagname")) {
	let tag = i.dataset.tag;
	let matching = findThumbByTag(tag);

	i.onmouseenter = () => {
		matching.forEach(p=>p.classList.add("active"))
	};

	i.onclick = (e) => {
		e.preventDefault();
		filter(tag);
	};

	i.onmouseleave = () => {
		matching.forEach(p=>p.classList.remove("active"))
	};
}
