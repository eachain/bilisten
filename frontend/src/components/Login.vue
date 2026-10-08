<script setup lang="ts">
  import { HugeiconsIcon } from '@hugeicons/vue';
  import { CheckIcon, Refresh01Icon } from '@hugeicons/core-free-icons';
  import { onActivated, onDeactivated, ref } from "vue";
  import { genLoginQRCode, type PollQRCodeResult, qrcodePoll, type QRCodeResult, QRCodeStatus } from "../libs/login";
  import * as qrcode from "qrcode";
  import { isLogined } from "../libs/backend";
  import { noticeAlert } from '../libs/alerts';

  const loginQRCode = ref();
  const status = ref(''); // 'polling', 'scanned', 'expired'

  let continuePoll = false;

  const stopPoll = () => (continuePoll = false);

  const refreshQRCode = async () => {
    if (await isLogined()) {
      window.location.hash = "#/";
      return;
    }

    status.value = 'polling';

    let result: QRCodeResult;
    try {
      result = await genLoginQRCode();
    } catch (error) {
      noticeAlert({level: 'alert-error', message: `获取登录二维码失败：${error}`});
      console.log(`get qrcode: ${error}`);
      return;
    }
    // const result = {
    //   "url": "https://account.bilibili.com/h5/account-h5/auth/scan-web?navhide=1&callback=close&qrcode_key=c25d85057e68ee77419a487cea7342d7&from=main-fe-header",
    //   "qrcode_key": "c25d85057e68ee77419a487cea7342d7",
    // };
    qrcode.toCanvas(loginQRCode.value, result.url, (error) => {
      if (error) {
        console.log("generate qrcode to canvas:", error);
        return;
      }
      console.log("generate qrcode to canvas: ok");
      continuePoll = true;
      startPolling(result.qrcode_key)
        .catch(() => status.value = 'expired')
        .finally(stopPoll);
    });
  }

  onActivated(refreshQRCode);

  onDeactivated(stopPoll);

  const startPolling = async (qrcode_key: string) => {
    while (continuePoll) {
      await new Promise((resolve) => setTimeout(resolve, 2000));

      let result: PollQRCodeResult;
      try {
        result = await qrcodePoll(qrcode_key);
      } catch(error) {
        console.log(`poll qrcode: ${error}`);
        continue;
      }
      console.log("poll qrcode result:", result);
      if (result.code === QRCodeStatus.notYetScanned) {
        continue;
      }
      if (result.code === QRCodeStatus.ok) {
        noticeAlert({level: 'alert-info', message: '登录成功'});
        window.location.hash = "#/";
        break;
      }
      if (result.code === QRCodeStatus.expired) {
        status.value = 'expired';
        break;
      }
      if (result.code === QRCodeStatus.scanned) {
        status.value = 'scanned';
        continue;
      }
      console.log("unrecognized qrcode status:", result.code, result.message);
      break;
    }
  };
</script>

<template>
  <div class="bg-image qr-wrapper w-full h-5/6 max-h-5/6 flex flex-col items-center gap-4">
    <span>扫描二维码登录</span>
    <div class="relative">
      <canvas ref="loginQRCode" />
      <div
        v-if="status === 'scanned'"
        class="absolute inset-0 z-10 w-full h-full flex flex-col items-center justify-center gap-2 bg-base-100/80 backdrop-blur-sm"
      >
        <span class="btn btn-xs btn-circle btn-success">
          <HugeiconsIcon :icon="CheckIcon" :size="16" />
        </span>
        <span>扫码成功</span>
        <span>请在手机上确认</span>
      </div>

      <div
        v-if="status === 'expired'"
        class="absolute inset-0 z-10 w-full h-full flex flex-col items-center justify-center gap-2 bg-base-100/80 backdrop-blur-sm"
        @click.stop="refreshQRCode"
      >
        <span class="btn btn-xs btn-circle">
          <HugeiconsIcon :icon="Refresh01Icon" :size="16" />
        </span>
        <span>二维码已失效</span>
        <span>点击刷新</span>
      </div>
    </div>
    <span>请使用哔哩哔哩客户端扫码登录</span>
  </div>
</template>


<style scoped lang="css">
.bg-image {
	background-image: url('../assets/bi-phone.png');
	background-repeat: no-repeat;
	background-size: contain;
}
</style>