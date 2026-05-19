/**
 * k6 Configuration File
 */

export const config = {
	BASE_URL: __ENV.BASE_URL || "http://localhost:3300",

	// Smoke Test defaults
	smoke: {
		VUS: Number(__ENV.VUS || 1),
		ITERATIONS: Number(__ENV.ITERATIONS || 20),
		MAX_DURATION: __ENV.MAX_DURATION || "30s",
	},

	// Load Test defaults
	load: {
		VUS: Number(__ENV.VUS || 10),
		DURATION: __ENV.DURATION || "1m",
		TARGET_DURATION: Number(__ENV.TARGET_DURATION || 1000),
	},

	// Common Thresholds: keep p95 under 1s for common HTTP paths.
	thresholds: {
		http_req_failed: ["rate<0.01"],
		http_req_duration: ["p(95)<1000"],
	},
};
