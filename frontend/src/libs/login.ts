import { cancelAlert, noticeAlert } from './alerts';
import { addCommomParams, callAPI, WebIfaceEndpoint } from './api';
import { logger, longPollLoginEvent } from './backend';

export type QRCodeResult = {
  url: string;
  qrcode_key: string;
};

export async function genLoginQRCode(): Promise<QRCodeResult> {
  logger.info(`generate login qrcode`);
  const params = new URLSearchParams();
  params.append('source', 'main-fe-header');
  params.append('go_url', 'https://www.bilibili.com/');
  params.append('web_location', '333.1007');
  params.append(
    'x-bili-locale-json',
    '{"c_locale":{"language":"zh","script":"Hans"},"always_translate":true}',
  );
  params.append('x-bili-redirect', '1');
  params.append(
    'x-bili-device-req-json',
    '{"platform":"web","device":"pc","spmid":"333.1007","mobi_app":"web_cn"}',
  );

  const api =
    'https://passport.bilibili.com/x/passport-login/web/qrcode/generate';
  const url = `${api}?${params.toString()}`;
  return await callAPI<QRCodeResult>({ url });
}

export type PollQRCodeResult = {
  url?: string;
  refresh_token: string;
  code: number;
  message: string;
};

export const QRCodeStatus = Object.freeze({
  ok: 0,
  expired: 86038,
  notYetScanned: 86101,
  scanned: 86090, // awaiting app confirmation
});

export async function qrcodePoll(
  qrcode_key: string,
): Promise<PollQRCodeResult> {
  const params = new URLSearchParams();
  params.append('qrcode_key', qrcode_key);
  params.append('source', 'main-fe-header');
  params.append('go_url', 'https://www.bilibili.com/');
  params.append('x-bili-redirect', '1');
  addCommomParams(params);

  const api = 'https://passport.bilibili.com/x/passport-login/web/qrcode/poll';
  const url = `${api}?${params.toString()}`;
  return await callAPI<PollQRCodeResult>({ url });
}

let isUserLogout = false;

export async function logout() {
  isUserLogout = true;
  const url = 'https://passport.bilibili.com/login/exit/v2';
  await callAPI<{}>({ url, method: 'POST' });
}

export type UserInfo = {
  isLogin: boolean;
  face: string;
  uname: string;
  mid: bigint;
};

export async function getUserInfo(): Promise<UserInfo> {
  const url = `${WebIfaceEndpoint}/nav`;
  return await callAPI<UserInfo>({ url });
}

const loginStatusCbs = new Map<string, () => void>();

export function registerOnLoginStatusChanged(name: string, cb: () => void) {
  loginStatusCbs.set(name, cb);
}

export function noticeLoginStatusChanged() {
  loginStatusCbs.forEach((cb) => cb());
}

let hasCheckStarted = false;

export async function autoCheckLoginStatus() {
  if (hasCheckStarted) {
    return;
  }
  hasCheckStarted = true;

  let alertId: number | undefined = undefined;
  let last = await longPollLoginEvent();
  while (true) {
    const now = await longPollLoginEvent();
    if (last === now) {
      continue;
    }
    if (last && !now) {
      if (!isUserLogout) {
        alertId = noticeAlert({
          level: 'alert-info',
          message: '登录状态已过期，请重新扫码登录',
          timeout: 30000,
        });
        const tmpAlertId = alertId;
        setTimeout(() => {
          if (tmpAlertId === alertId) {
            alertId = undefined;
          }
        }, 30000);
      }
      isUserLogout = false;
    } else if (!last && now) {
      if (alertId) {
        cancelAlert(alertId);
        alertId = undefined;
      }
    }
    last = now;
    noticeLoginStatusChanged();
  }
}
