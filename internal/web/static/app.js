const form = document.getElementById("shorten");
const input = document.getElementById("target");
const button = form.querySelector("button[type='submit']");
const result = document.getElementById("result");

form.addEventListener("submit", async (event) => {
	event.preventDefault();
	button.disabled = true;

	try {
		const response = await fetch("/api/links", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ target: input.value }),
		});

		if (!response.ok) {
			showError(await response.text());
			return;
		}

		const { short_url } = await response.json();
		showResult(short_url, input.value.trim());
	} catch (err) {
		showError("Couldn't reach the server.");
	} finally {
		button.disabled = false;
	}
});

function showResult(shortURL, target) {
	const link = el("a", "short-link", shortURL);
	link.href = shortURL;
	link.target = "_blank";
	link.rel = "noopener noreferrer";

	const copy = el("button", "btn btn-sm", "Copy");
	copy.type = "button";
	copy.addEventListener("click", async () => {
		try {
			await navigator.clipboard.writeText(shortURL);
			copy.textContent = "Copied";
			setTimeout(() => (copy.textContent = "Copy"), 1600);
		} catch (err) {
			// navigator.clipboard needs a secure context (https or localhost).
			copy.textContent = "Press ⌘C";
			getSelection().selectAllChildren(link);
		}
	});

	const head = el("div", "result-head");
	head.append(el("span", "result-label", "Short link"), copy);

	const body = el("div", "result-body");
	body.append(link, el("p", "target", "→ " + target));

	const card = el("div", "result");
	card.append(head, body);

	result.replaceChildren(card);
}

function showError(message) {
	result.replaceChildren(el("p", "error", message.trim()));
}

function el(tag, className, text) {
	const node = document.createElement(tag);
	node.className = className;
	if (text !== undefined) node.textContent = text;
	return node;
}
