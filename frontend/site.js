// Only public configuration comes from the API. These files need no build step.
window.tinychessConfig = fetch("/api/config")
  .then((response) => {
    if (!response.ok) throw new Error("Configuration request failed");
    return response.json();
  })
  .catch(() => ({ coachEnabled: false, version: "unavailable" }));

window.tinychessConfig.then((config) => {
  const version = document.getElementById("version");
  if (version) {
    version.textContent = config.version;
    version.href = `https://github.com/dustywusty/tinychess/tree/${encodeURIComponent(config.version)}`;
  }
});
