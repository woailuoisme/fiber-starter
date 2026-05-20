import { apiChecks } from "../common/checks.js";
import { config } from "../common/config.js";
import { rootApi } from "../modules/root.js";

export const options = {
	scenarios: {
		smoke_test: {
			executor: "shared-iterations",
			vus: config.smoke.VUS,
			iterations: config.smoke.ITERATIONS,
			maxDuration: config.smoke.MAX_DURATION,
		},
	},
	thresholds: {
		...config.thresholds,
		http_req_duration: [`p(95)<${config.load.TARGET_DURATION}`],
	},
};

export default function () {
	const res = rootApi.getHome();

	apiChecks.status200(res);
	apiChecks.successTrue(res);
	apiChecks.rootResponse(res);
}
