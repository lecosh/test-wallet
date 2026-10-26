import http from 'k6/http';
import exec from 'k6/execution';
import { check } from 'k6';

const baseURL = 'http://api:8000/api/v1';
const rate = Number(1000);
const seconds = Number(60);
const planned = rate * seconds;

export const options = {
  scenarios: {
    deposits: {
      executor: 'constant-arrival-rate',
      rate,
      timeUnit: '1s',
      duration: `${seconds}s`,
      preAllocatedVUs: 1000,
      gracefulStop: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate==0'],
    checks: ['rate==1'],
    dropped_iterations: ['count==0'],
    // Deposits, plus wallet creation and the final balance request
    http_reqs: [`count==${planned + 2}`],
  },
};

export function setup() {
  const response = http.post(`${baseURL}/wallets/create`, null, { timeout: '10s' });
  if (response.status !== 201) {
    throw new Error(`Cannot create test wallet: HTTP ${response.status}`);
  }
  const wallet = response.json();
  if (typeof wallet.walletId !== 'string' || !wallet.walletId || wallet.balance !== 0) {
    throw new Error('New wallet must have an ID and a zero balance');
  }
  return wallet.walletId;
}

export default function (walletId) {
  if (exec.scenario.iterationInTest >= planned) return;

  const response = http.post(
    `${baseURL}/wallets`,
    JSON.stringify({ walletId, operationType: 'DEPOSIT', amount: 1 }),
    { headers: { 'Content-Type': 'application/json' }, timeout: '10s' },
  );
  check(response, { 'deposit accepted': r => r.status === 200 });
}

export function teardown(walletId) {
  const response = http.get(
    `${baseURL}/wallets/${walletId}`, { timeout: '10s' });
  check(response, {
    'final balance includes every deposit': r =>
      r.status === 200 && r.json('walletId') === walletId && r.json('balance') === planned,
  });
}
