import { apiChecks } from "../common/checks.js";
import { config } from "../common/config.js";
import { rootApi } from "../modules/root.js";

export const options = {
	scenarios: {
		load_test: {
			executor: "constant-vus",
			vus: config.load.VUS,
			duration: config.load.DURATION,
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
