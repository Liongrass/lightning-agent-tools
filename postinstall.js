#!/usr/bin/env node

// postinstall.js downloads the pre-built lightning-mcp-server binary for the
// current platform and architecture from GitHub Releases. This follows the
// same pattern used by esbuild, turbo, and other Go/Rust projects distributed
// via npm.
//
// The binary is placed in bin/ so the npm "bin" entry in package.json can
// reference it directly, making `npx lightning-mcp-server` work out of the
// box.

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const os = require("os");

// PLATFORM_MAP translates Node.js platform/arch identifiers to the Go build
// system naming convention used by the release workflow.
const PLATFORM_MAP = {
	"darwin-arm64": "darwin-arm64",
	"darwin-x64": "darwin-amd64",
	"linux-arm64": "linux-arm64",
	"linux-x64": "linux-amd64",
	"win32-x64": "windows-amd64",
};

// REPO is the GitHub repository that hosts the release binaries.
const REPO = "lightninglabs/lightning-agent-kit";

// BINARY_NAME is the name of the binary inside the release tarball.
const BINARY_NAME = "lightning-mcp-server";

// getPackageVersion reads the version from package.json to determine which
// GitHub release tag to download from.
function getPackageVersion() {
	const pkg = JSON.parse(
		fs.readFileSync(path.join(__dirname, "package.json"), "utf8"),
	);
	return pkg.version;
}

// getPlatformKey returns the platform-architecture key for the current
// system, e.g. "darwin-arm64" or "linux-x64".
function getPlatformKey() {
	const platform = os.platform();
	const arch = os.arch();
	return `${platform}-${arch}`;
}

// getBinaryPath returns the file path where the downloaded binary will be
// placed within the package's bin directory.
function getBinaryPath() {
	const ext = os.platform() === "win32" ? ".exe" : "";
	return path.join(__dirname, "bin", `${BINARY_NAME}${ext}`);
}

// fetchBuffer follows HTTP redirects (GitHub Releases redirects to a CDN)
// and returns the final response body as a buffer.
function fetchBuffer(url) {
	return new Promise((resolve, reject) => {
		const opts = { headers: { "User-Agent": "npm-postinstall" } };
		https.get(url, opts, (res) => {
			if (res.statusCode >= 300 && res.statusCode < 400 &&
				res.headers.location) {
				fetchBuffer(res.headers.location)
					.then(resolve, reject);
				return;
			}
			if (res.statusCode !== 200) {
				reject(new Error(
					`Download failed: HTTP ${res.statusCode} ` +
					`from ${url}`,
				));
				return;
			}
			const chunks = [];
			res.on("data", (chunk) => chunks.push(chunk));
			res.on("end", () => resolve(Buffer.concat(chunks)));
			res.on("error", reject);
		}).on("error", reject);
	});
}

// extractTarGz extracts a .tar.gz buffer to the specified directory using
// the system tar command. The binary is expected at the root of the archive.
function extractTarGz(buffer, destDir) {
	const tmpFile = path.join(
		os.tmpdir(), `${BINARY_NAME}-${Date.now()}.tar.gz`,
	);
	fs.writeFileSync(tmpFile, buffer);
	try {
		execSync(`tar xzf "${tmpFile}" -C "${destDir}"`, {
			stdio: "pipe",
		});
	} finally {
		fs.unlinkSync(tmpFile);
	}
}

// main downloads and installs the correct binary for the current platform.
async function main() {
	const platformKey = getPlatformKey();
	const goTarget = PLATFORM_MAP[platformKey];

	if (!goTarget) {
		console.error(
			`Unsupported platform: ${platformKey}. ` +
			`Supported: ${Object.keys(PLATFORM_MAP).join(", ")}`,
		);
		console.error(
			"You can build from source instead:\n" +
			"  cd mcp-server && make build",
		);
		process.exit(1);
	}

	const version = getPackageVersion();
	const tag = `v${version}`;
	const assetName = `${BINARY_NAME}-${goTarget}.tar.gz`;
	const url =
		`https://github.com/${REPO}/releases/download/${tag}/${assetName}`;

	console.log(
		`Downloading ${BINARY_NAME} ${tag} for ${goTarget}...`,
	);

	try {
		const buffer = await fetchBuffer(url);

		// Ensure the bin directory exists before extracting.
		const binDir = path.join(__dirname, "bin");
		fs.mkdirSync(binDir, { recursive: true });

		extractTarGz(buffer, binDir);

		// Make the binary executable on Unix systems.
		const binaryPath = getBinaryPath();
		if (os.platform() !== "win32") {
			fs.chmodSync(binaryPath, 0o755);
		}

		console.log(`Installed ${BINARY_NAME} to ${binaryPath}`);
	} catch (err) {
		console.error(
			`Failed to download ${BINARY_NAME}: ${err.message}`,
		);
		console.error("");
		console.error("You can build from source instead:");
		console.error("  cd mcp-server && make build");
		console.error("");
		console.error(
			"Or download from: " +
			`https://github.com/${REPO}/releases/tag/${tag}`,
		);
		process.exit(1);
	}
}

main();
