import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration');

export const options = {
  vus: 100,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(99)<50'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.01'],
  },
};

export default function() {
  const res = http.get('http://localhost:8080/api/users');
  
  requestDuration.add(res.timings.duration);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has body': (r) => r.body.length > 0,
  });
  
  errorRate.add(!success);
  
  sleep(0.1);
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, opts) {
  const indent = opts.indent || '';
  let output = '\n' + indent + '=== Load Test Summary ===\n\n';
  
  const metrics = data.metrics;
  
  output += indent + `HTTP Requests:\n`;
  output += indent + `  Total: ${metrics.http_reqs?.values?.runs || 0}\n`;
  output += indent + `  Failed: ${metrics.http_req_failed?.values?.runs || 0}\n`;
  output += indent + `  Failed Rate: ${((metrics.http_req_failed?.values?.rate || 0) * 100).toFixed(2)}%\n\n`;
  
  output += indent + `Response Times:\n`;
  output += indent + `  p99: ${metrics.http_req_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  output += indent + `  avg: ${metrics.http_req_duration?.values?.avg?.toFixed(2) || 0}ms\n`;
  output += indent + `  max: ${metrics.http_req_duration?.values?.max?.toFixed(2) || 0}ms\n\n`;
  
  output += indent + `Virtual Users: ${options.vus}\n`;
  output += indent + `Duration: ${options.duration}\n`;
  
  return output;
}
