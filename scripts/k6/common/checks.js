import { check } from "k6";

/**
 * Common checks for API responses
 */
export const apiChecks = {
	status200: (res) => check(res, { "status is 200": (r) => r.status === 200 }),

	successTrue: (res) =>
		check(res, { "success is true": (r) => r.json("success") === true }),

	validJSON: (res) =>
		check(res, {
			"is valid json": (r) => {
				try {
					r.json();
					return true;
				} catch (_e) {
					return false;
				}
			},
		}),

	// Feature specific checks
	rootResponse: (res) =>
		check(res, {
			"message matches": (r) =>
				r.json("message") === "Welcome to Fiber Starter API",
			"api version link present": (r) => r.json("data.api") === "/api/v1",
			"docs link present": (r) => r.json("data.docs") === "/docs",
		}),
};
