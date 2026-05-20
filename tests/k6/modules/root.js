import http from "k6/http";
import { config } from "../common/config.js";

/**
 * Root API Module
 */
export const rootApi = {
	getHome: () => {
		const url = `${config.BASE_URL.replace(/\/$/, "")}/`;
		return http.get(url);
	},
};
