import http from "k6/http";
import { check, sleep } from "k6";

const baseURL = __ENV.BASE_URL || "http://localhost:18080";
const reference = __ENV.SPARE_PART_REFERENCE || "BRK-PAD-4521";

export const options = {
  stages: [
    { duration: "30s", target: 50 },
    { duration: "3m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
  },
};

export default function () {
  const response = http.get(`${baseURL}/spare-part/${reference}`);

  check(response, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(0.1);
}
